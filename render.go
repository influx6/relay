// render.go embodies rendering codecs for different message patterns eg json, html templates

package relay

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"mime/multipart"
	"net/url"
	"text/template"
)

const (
	// ContentBinary header value for binary data.
	ContentBinary = "application/octet-stream"
	// ContentHTML header value for HTML data.
	ContentHTML = "text/html"
	// ContentJSON header value for JSON data.
	ContentJSON = "application/json"
	// ContentJSONP header value for JSONP data.
	ContentJSONP = "application/javascript"
	// ContentLength header constant.
	ContentLength = "Content-Length"
	// ContentText header value for Text data.
	ContentText = "text/plain"
	// ContentType header constant.
	ContentType = "Content-Type"
	// ContentXHTML header value for XHTML data.
	ContentXHTML = "application/xhtml+xml"
	// ContentXML header value for XML data.
	ContentXML = "text/xml"
	// Default character encoding.
	defaultCharset = "charset=UTF-8"
)

// Message represent a message data
type Message struct {
	MessageType string
	Method      string
	Payload     []byte
	Form        url.Values
	PostForm    url.Values
	Multipart   *multipart.Form
	Params      Collector
}

//MessageDecoder provides the message decoding decoder for *HTTPRequest objects
var MessageDecoder = NewHTTPDecoder(func(req *HTTPRequest) (*Message, error) {
	return loadData(req)
})

//SimpleEncoder provides simple encoding that checks if the value given is a string or []byte else returns an error
var SimpleEncoder = NewHTTPEncoder(func(req *HTTPRequest, d interface{}) (int, error) {

	switch d.(type) {
	case string:
		so, _ := d.(string)
		return req.Res.Write([]byte(so))
	case []byte:
		bo, _ := d.([]byte)
		return req.Res.Write(bo)
	}

	return 0, NewCustomError("SimpleEncoder", "data is neither a 'string' or '[]byte' type ")
})

// BasicHTTPCodec returns a new codec based on the basic deocder and encoder
var BasicHTTPCodec = NewHTTPCodec(SimpleEncoder, MessageDecoder)

// Head provides a basic message status and head values
type Head struct {
	Status  int
	Content string
}

// Text provides a basic text messages
type Text struct {
	Head
	Data string
}

//TextEncoder provides the jsonp encoder for encoding json messages
var TextEncoder = NewHTTPEncoder(func(r *HTTPRequest, d interface{}) (int, error) {
	setUpHeadings(r)

	tx, ok := d.(Text)

	if !ok {
		return 0, NewCustomError("TextEncoder", "received type is not a Text{}")
	}

	r.Res.WriteHeader(tx.Status)

	if tx.Content == "" {
		r.Res.Header().Add("Content-Type", ContentText)
	} else {
		r.Res.Header().Add("Content-Type", tx.Content)
	}

	return r.Res.Write([]byte(tx.Data))
})

// JSONP provides a basic jsonp messages
type JSONP struct {
	Head
	Indent   bool
	Callback string
	Data     interface{}
}

//JSONPEncoder provides the jsonp encoder for encoding json messages
var JSONPEncoder = NewHTTPEncoder(func(req *HTTPRequest, d interface{}) (int, error) {

	var res []byte
	var err error

	jop, ok := d.(JSONP)

	if !ok {
		return 0, NewCustomError("JSONP", "encoder expected JSONP type")
	}

	if jop.Indent {
		res, err = json.MarshalIndent(jop.Data, "", "  ")
		res = append(res, '\n')
	} else {
		res, err = json.Marshal(jop.Data)
	}

	if err != nil {
		return 0, err
	}

	if jop.Content == "" {
		req.Res.Header().Add("Content-Type", ContentJSONP)
	} else {
		req.Res.Header().Add("Content-Type", jop.Content)
	}
	// req.Res.Header().Add("Content-Type", jop.Content)
	req.Res.WriteHeader(jop.Status)

	var fos []byte
	fos = append(fos, []byte(jop.Callback+"(")...)
	fos = append(fos, res...)
	fos = append(fos, []byte(");")...)

	return req.Res.Write(fos)
})

// JSON provides a basic json messages
type JSON struct {
	Head
	Indent       bool
	UnEscapeHTML bool
	Stream       bool
	Data         interface{}
}

//JSONEncoder provides the jsonp encoder for encoding json messages
var JSONEncoder = NewHTTPEncoder(func(req *HTTPRequest, d interface{}) (int, error) {

	jso, ok := d.(JSON)

	if !ok {
		return 0, NewCustomError("JSON", "Wrong type,expected JSON type")
	}

	if jso.Content == "" {
		req.Res.Header().Add("Content-Type", ContentJSON)
	} else {
		req.Res.Header().Add("Content-Type", jso.Content)
	}

	// req.Res.Header().Add("Content-Type", ContentJSON)
	if jso.Stream {
		// req.Res.Header().Add("Content-Type", jso.Content)
		req.Res.WriteHeader(jso.Status)
		return 1, json.NewEncoder(req.Res).Encode(jso.Data)
	}

	var res []byte
	var err error

	if jso.Indent {
		res, err = json.MarshalIndent(jso.Data, "", "  ")
		res = append(res, '\n')
	} else {
		res, err = json.Marshal(jso.Data)
	}

	if err != nil {
		return 0, err
	}

	// Unescape HTML if needed.
	if jso.UnEscapeHTML {
		res = bytes.Replace(res, []byte("\\u003c"), []byte("<"), -1)
		res = bytes.Replace(res, []byte("\\u003e"), []byte(">"), -1)
		res = bytes.Replace(res, []byte("\\u0026"), []byte("&"), -1)
	}

	req.Res.Header().Add("Content-Type", jso.Content)
	req.Res.WriteHeader(jso.Status)
	req.Res.Write(res)

	return 0, nil
})

// HTML provides a basic html messages
type HTML struct {
	Head
	Name     string
	Layout   string
	Binding  interface{}
	Template *template.Template
}

//HTMLEncoder provides the jsonp encoder for encoding json messages
var HTMLEncoder = NewHTTPEncoder(func(req *HTTPRequest, d interface{}) (int, error) {

	hop, ok := d.(HTML)

	if !ok {
		return 0, NewCustomError("HTML", "encoder received wrong type,expected HTML struct type")
	}

	bou := bufPool.Get()

	if err := hop.Template.ExecuteTemplate(bou, hop.Name, hop.Binding); err != nil {
		return 0, err
	}

	if hop.Content == "" {
		req.Res.Header().Add("Content-Type", ContentHTML)
	} else {
		req.Res.Header().Add("Content-Type", hop.Content)
	}

	nd, err := bou.WriteTo(req.Res)

	bufPool.Put(bou)

	return int(nd), err
})

// XML provides a basic html messages
type XML struct {
	Head
	Indent bool
	Prefix []byte
	Data   interface{}
}

//XMLEncoder provides the jsonp encoder for encoding json messages
var XMLEncoder = NewHTTPEncoder(func(req *HTTPRequest, d interface{}) (int, error) {

	jso, ok := d.(XML)

	if !ok {
		return 0, NewCustomError("XML", "Wrong type,expected XML type")
	}

	if jso.Content == "" {
		req.Res.Header().Add("Content-Type", ContentXML)
	} else {
		req.Res.Header().Add("Content-Type", jso.Content)
	}

	var res []byte
	var err error

	if jso.Indent {
		res, err = xml.MarshalIndent(jso.Data, "", "  ")
		res = append(res, '\n')
	} else {
		res, err = xml.Marshal(jso.Data)
	}

	if err != nil {
		return 0, err
	}

	return req.Res.Write(res)
})

// ErrInvalidByteType is returned when the interface is ot a []byte
var ErrInvalidByteType = errors.New("interface not a []byte")

//ByteSocketEncoder provides the basic websocket message encoder for encoding json messages
var ByteSocketEncoder = NewSocketEncoder(func(w *Websocket, t int, bu interface{}) (int, error) {
	var err error
	var size int

	bo, ok := bu.([]byte)

	if !ok {
		return 0, ErrInvalidByteType
	}

	err = w.Conn.WriteMessage(t, bo)

	if err == nil {
		size = len(bo)
	}

	return size, err
})

//ByteSocketDecoder provides the basic websocket decoder which justs returns a decoder
var ByteSocketDecoder = NewSocketDecoder(func(t int, bu []byte) (interface{}, error) {
	return bu, nil
})

// BasicSocketCodec returns a codec using the socket encoder and decoder
var BasicSocketCodec = NewSocketCodec(ByteSocketEncoder, ByteSocketDecoder)
