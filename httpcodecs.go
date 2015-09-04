package relay

import (
	"io"
	"strings"
)

// HTTPCodeco represents a generic http codec
type HTTPCodeco struct {
	HTTPEncoder
	HTTPDecoder
}

// NewHTTPCodec returns a new http codec
func NewHTTPCodec(e HTTPEncoder, d HTTPDecoder) HTTPCodec {
	return &HTTPCodeco{e, d}
}

// BasicHTTPCodec returns a new codec based on the basic deocder and encoder
func BasicHTTPCodec() HTTPCodec {
	return NewHTTPCodec(&BasicHTTPEncoder{}, &BasicHTTPDecoder{})
}

// BasicHTTPEncoder provides a simple encoder of standard http responses
type BasicHTTPEncoder struct {
	HTTPEncoder
}

func setUpHeadings(r *HTTPRequest) {
	agent, ok := r.Req.Header["User-Agent"]

	if ok {
		ag := strings.Join(agent, ";")
		msie := strings.Index(ag, ";MSIE")
		trident := strings.Index(ag, "Trident/")

		if msie != -1 || trident != -1 {
			r.Res.Header().Set("X-XSS-Protection", "0")
		}
	}

	origin, ok := r.Req.Header["Origin"]

	if ok {
		r.Res.Header().Set("Access-Control-Allow-Credentials", "true")
		r.Res.Header().Set("Access-Control-Allow-Origin", strings.Join(origin, ";"))
	} else {
		r.Res.Header().Set("Access-Control-Allow-Origin", "*")
	}
}

func loadData(r *HTTPRequest) (*Message, error) {
	msg := Message{}
	msg.Method = r.Req.Method

	content, ok := r.Req.Header["Content-Type"]

	if ok {
		muxcontent := strings.Join(content, ";")

		if strings.Index(muxcontent, "application/x-www-form-urlencode") != -1 {
			if err := r.Req.ParseForm(); err != nil {
				return nil, err
			}

			msg.MessageType = "form"
			msg.Method = r.Req.Method
			msg.Form = r.Req.Form
			msg.PostForm = r.Req.PostForm

			return &msg, nil
		}

		if strings.Index(muxcontent, "multipart/form-data") != -1 {
			if err := r.Req.ParseMultipartForm(32 << 20); err != nil {
				return nil, err
			}

			msg.MessageType = "multipart"
			msg.Multipart = r.Req.MultipartForm
			return &msg, nil
		}
	}

	if r.Req.Body == nil {
		return nil, nil
	}

	data := make([]byte, r.Req.ContentLength)
	_, err := r.Req.Body.Read(data)

	if err != nil && err != io.EOF {
		return nil, err
	}

	msg.MessageType = "body"
	msg.Payload = data

	return &msg, nil
}

// Encode encodes and writes the payload and writes it appropriately to the writer
func (b *BasicHTTPEncoder) Encode(r *HTTPRequest, payload interface{}) (int, error) {
	setUpHeadings(r)
	return r.Res.Write(payload.([]byte))
}

// BasicHTTPDecoder provides a basic http encoder for response to http-requests
type BasicHTTPDecoder struct {
	HTTPDecoder
}

// Decode decodes the data from the object received and returns a enw Message object
func (d *BasicHTTPDecoder) Decode(r *HTTPRequest) (*Message, error) {
	return loadData(r)
}
