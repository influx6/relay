package relay

import (
	"io"
	"net/http"
	"strings"
)

//IsWebSocketRequest returns true if a http.Request object is based on
//websocket standards
func IsWebSocketRequest(r *http.Request) bool {
	var _ interface{}
	_, hasUpgrade := r.Header["Upgrade"]
	_, hasSec := r.Header["Sec-Websocket-Version"]
	_, hasExt := r.Header["Sec-Websocket-Extensions"]
	_, hasKey := r.Header["Sec-Websocket-Key"]
	return hasUpgrade && hasSec && hasExt && hasKey
}

func setUpHeadings(r *HTTPRequest) {
	agent, ok := r.Req.Header["User-Agent"]

	if ok {
		ag := strings.Join(agent, ";")
		msie := strings.Index(ag, ";MSIE")
		trident := strings.Index(ag, "Trident/")

		if msie != -1 || trident != -1 {
			// r.Res.Header().Set("X-XSS-Protection", "0")
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
