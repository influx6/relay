// render.go embodies rendering codecs for different message patterns eg json, html templates

package relay

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
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

//MessageDecoder provides the message decoding decoder for *Context objects
var MessageDecoder = NewHTTPDecoder(func(req *Context) (*Message, error) {
	return loadData(req)
})

// UseHTTPEncoder wires up the MessageDecoder as an automatic decoder
func UseHTTPEncoder(enc Encoder) HTTPCodec {
	return NewHTTPCodec(enc, MessageDecoder)
}

//SimpleEncoder provides simple encoding that checks if the value given is a string or []byte else returns an error
var SimpleEncoder = NewEncoder(func(w io.Writer, d interface{}) (int, error) {

	switch d.(type) {
	case string:
		so, _ := d.(string)
		return w.Write([]byte(so))
	case []byte:
		bo, _ := d.([]byte)
		return w.Write(bo)
	}

	return 0, NewCustomError("SimpleEncoder", "data is neither a 'string' or '[]byte' type ")
})

// BasicHTTPCodec returns a new codec based on the basic deocder and encoder
var BasicHTTPCodec = NewHTTPCodec(SimpleEncoder, MessageDecoder)

// Head provides a basic message status and head values
type Head struct {
	Status  int
	Content string
	Headers http.Header
}

// BasicHeadEncoder provides a basic encoder for writing the response headers and
// status information,it provides the flexibility to build even more reusable and
// useful header writing methods
var BasicHeadEncoder = NewHeadEncoder(func(c *Context, h *Head) error {
	WriteHead(c, h)
	return nil
})

// WriteHead uses a context and writes into the resposne with a head struct
func WriteHead(c *Context, h *Head) {
	WriteRawHead(c.Res, h)
}

//WriteRawHead writes a head struct into a ResponseWriter
func WriteRawHead(c http.ResponseWriter, h *Head) {
	c.WriteHeader(h.Status)
	//copy over the headers
	for k, v := range h.Headers {
		for _, vs := range v {
			c.Header().Add(k, vs)
		}
	}

	//write the Content-Type if not a empty string, we do it down here to preserve
	//the given value incase there was an over-write in the headers provided
	//TODO: decide wether to use .Add() or .Set()

	if h.Content != "" {
		c.Header().Add("Content-Type", h.Content)
	}
}

// Text provides a basic text messages
type Text struct {
	*Head
	Data string
}

//TextEncoder provides the jsonp encoder for encoding json messages
var TextEncoder = NewEncoder(func(w io.Writer, d interface{}) (int, error) {
	tx, ok := d.(*Text)

	if !ok {
		return 0, NewCustomError("TextEncoder", "received type is not a Text{}")
	}

	return w.Write([]byte(tx.Data))
})

// TextRender returns a text struct for rendering
func TextRender(status int, data string) *Text {
	return &Text{
		Data: data,
		Head: &Head{
			Status:  status,
			Content: ContentText,
		},
	}
}

// JSONP provides a basic jsonp messages
type JSONP struct {
	*Head
	Indent   bool
	Callback string
	Data     interface{}
}

//JSONPEncoder provides the jsonp encoder for encoding json messages
var JSONPEncoder = NewEncoder(func(w io.Writer, d interface{}) (int, error) {
	var res []byte
	var err error

	jop, ok := d.(*JSONP)

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

	var fos []byte
	fos = append(fos, []byte(jop.Callback+"(")...)
	fos = append(fos, res...)
	fos = append(fos, []byte(");")...)

	return w.Write(fos)
})

// JSONPRender returns a jsonp struct for rendering
func JSONPRender(status int, indent bool, callback string, data interface{}) *JSONP {
	return &JSONP{
		Indent:   indent,
		Callback: callback,
		Data:     data,
		Head: &Head{
			Status:  status,
			Content: ContentJSONP,
		},
	}
}

// JSON provides a basic json messages
type JSON struct {
	*Head
	Indent       bool
	UnEscapeHTML bool
	Stream       bool
	Data         interface{}
}

//JSONEncoder provides the jsonp encoder for encoding json messages
var JSONEncoder = NewEncoder(func(w io.Writer, d interface{}) (int, error) {

	jso, ok := d.(*JSON)

	if !ok {
		return 0, NewCustomError("JSON", "Wrong type,expected JSON type")
	}

	if jso.Stream {
		return 1, json.NewEncoder(w).Encode(jso.Data)
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

	w.Write(res)

	return 0, nil
})

// JSONRender returns a json struct for rendering
func JSONRender(status int, data interface{}, indent, stream, unescape bool) *JSON {
	return &JSON{
		Indent:       indent,
		Stream:       stream,
		UnEscapeHTML: unescape,
		Data:         data,
		Head: &Head{
			Status:  status,
			Content: ContentJSON,
		},
	}
}

// HTML provides a basic html messages
type HTML struct {
	*Head
	Layout   string
	Binding  interface{}
	Template *template.Template
}

//HTMLEncoder provides the jsonp encoder for encoding json messages
var HTMLEncoder = NewEncoder(func(w io.Writer, d interface{}) (int, error) {

	hop, ok := d.(*HTML)

	if !ok {
		return 0, NewCustomError("HTML", "encoder received wrong type,expected HTML struct type")
	}

	bou := bufPool.Get()

	if err := hop.Template.ExecuteTemplate(bou, hop.Layout, hop.Binding); err != nil {
		return 0, err
	}

	nd, err := bou.WriteTo(w)

	bufPool.Put(bou)

	return int(nd), err
})

// HTMLRender returns a html struct for rendering
func HTMLRender(status int, layout string, binding interface{}, tl *template.Template) *HTML {
	return &HTML{
		Layout:   layout,
		Binding:  binding,
		Template: tl,
		Head: &Head{
			Status:  status,
			Content: ContentHTML,
		},
	}
}

// XML provides a basic html messages
type XML struct {
	*Head
	Indent bool
	Prefix []byte
	Data   interface{}
}

//XMLEncoder provides the jsonp encoder for encoding json messages
var XMLEncoder = NewEncoder(func(w io.Writer, d interface{}) (int, error) {

	jso, ok := d.(*XML)

	if !ok {
		return 0, NewCustomError("XML", "Wrong type,expected XML type")
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

	return w.Write(res)
})

// XMLRender returns a html struct for rendering
func XMLRender(status int, indent bool, data interface{}, prefix []byte) *XML {
	return &XML{
		Indent: indent,
		Prefix: prefix,
		Data:   data,
		Head: &Head{
			Status:  status,
			Content: ContentXML,
		},
	}
}

// ErrInvalidByteType is returned when the interface is ot a []byte
var ErrInvalidByteType = errors.New("interface not a []byte")

//ByteEncoder provides the basic websocket message encoder for encoding json messages
var ByteEncoder = NewEncoder(func(w io.Writer, bu interface{}) (int, error) {
	bo, ok := bu.([]byte)

	if !ok {
		return 0, ErrInvalidByteType
	}

	return w.Write(bo)
})

//ByteSocketDecoder provides the basic websocket decoder which justs returns a decoder
var ByteSocketDecoder = NewSocketDecoder(func(t int, bu []byte) (interface{}, error) {
	return bu, nil
})

// BasicSocketCodec returns a codec using the socket encoder and decoder
var BasicSocketCodec = NewSocketCodec(ByteEncoder, ByteSocketDecoder)
