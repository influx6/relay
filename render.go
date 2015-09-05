// render.go embodies rendering codecs for different message patterns eg json, html templates

package relay

import (
	"mime/multipart"
	"net/url"
	"text/template"
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

	return nil, nil
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
func BasicHTTPCodec() HTTPCodec {
	return NewHTTPCodec(SimpleEncoder, MessageDecoder)
}

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

	if tx.Content != "" {
		r.Res.Header().Add("Content-Type", tx.Content)
	}

	return r.Res.Write([]byte(tx.Data))
})

// JSONP provides a basic jsonp messages
type JSONP struct {
	Head
	Callback string
	Data     interface{}
}

//JSONPEncoder provides the jsonp encoder for encoding json messages
var JSONPEncoder = NewHTTPEncoder(func(req *HTTPRequest, d interface{}) (int, error) {

	return 0, nil
})

// JSON provides a basic json messages
type JSON struct {
	Head
	Data   interface{}
	Indent bool
}

//JSONEncoder provides the jsonp encoder for encoding json messages
var JSONEncoder = NewHTTPEncoder(func(req *HTTPRequest, d interface{}) (int, error) {

	return 0, nil
})

// HTML provides a basic html messages
type HTML struct {
	Head
	Name     string
	Template *template.Template
}

//HTMLEncoder provides the jsonp encoder for encoding json messages
var HTMLEncoder = NewHTTPEncoder(func(req *HTTPRequest, d interface{}) (int, error) {

	return 0, nil
})

// XML provides a basic html messages
type XML struct {
	Head
	Indent bool
	Prefix []byte
}

//XMLEncoder provides the jsonp encoder for encoding json messages
var XMLEncoder = NewHTTPEncoder(func(req *HTTPRequest, d interface{}) (int, error) {

	return 0, nil
})
