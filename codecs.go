package relay

import (
	"io"
)

// Codecs provide a flexible approach to defining different rendering and message systems for replies without overloading the request handling code. Codecs follow the basic idea that their is an encoder and a decoder that produce and consume sets of data and handle the writing for you. This way we can create custom message patterns for response without ease and testability

// Encoders accept interface{} as the data type because this allows flexible data messages to be constructed without restrictions

//Encoder takes an inteface and encodes it into a []byte slice
type Encoder interface {
	Encode(io.Writer, interface{}) (int, error)
}

//EncodeHandler provides a handler for http encoding function
type EncodeHandler func(io.Writer, interface{}) (int, error)

//encodeBuild provides a base struct for HTTPEncoder
type encodeBuild struct {
	handler EncodeHandler
}

// Encode provides the HTTPEncoder encode function
func (e *encodeBuild) Encode(r io.Writer, b interface{}) (int, error) {
	return e.handler(r, b)
}

// NewEncoder provides a nice way of build an Encoder
func NewEncoder(fx EncodeHandler) Encoder {
	return &encodeBuild{fx}
}

// HeadEncoder takes an inteface and encodes it into a []byte slice
type HeadEncoder interface {
	Encode(*Context, *Head) error
}

// HeadEncodeHandler provides a handler for http encoding function
type HeadEncodeHandler func(*Context, *Head) error

//headEncodeBuild provides a base struct for HTTPEncoder
type headEncodeBuild struct {
	handler HeadEncodeHandler
}

// Encode provides the HTTPEncoder encode function
func (e *headEncodeBuild) Encode(c *Context, b *Head) error {
	return e.handler(c, b)
}

//NewHeadEncoder provides a nice way of build an http headers encoders
func NewHeadEncoder(fx HeadEncodeHandler) HeadEncoder {
	return &headEncodeBuild{fx}
}

// HTTPDecoder provides a single member rule that takes a []byte and decodes it into
//its desired format
type HTTPDecoder interface {
	Decode(*Context) (*Message, error)
}

//HTTPDecodeHandler provides a base decoder function type
type HTTPDecodeHandler func(*Context) (*Message, error)

//decodeBuild provides a base struct for HTTPEncoder
type decodeBuild struct {
	handler HTTPDecodeHandler
}

// Decode provides the HTTPDecoder encode function
func (e *decodeBuild) Decode(r *Context) (*Message, error) {
	return e.handler(r)
}

//NewHTTPDecoder provides a nice way of build an HTTPEncoder
func NewHTTPDecoder(fx HTTPDecodeHandler) HTTPDecoder {
	return &decodeBuild{fx}
}

//HTTPCodec provides a interface define method for custom message formats
type HTTPCodec interface {
	Encoder
	HTTPDecoder
}

// httpCodec represents a generic http codec
type httpCodec struct {
	Encoder
	HTTPDecoder
}

// NewHTTPCodec returns a new http codec
func NewHTTPCodec(e Encoder, d HTTPDecoder) HTTPCodec {
	return &httpCodec{e, d}
}

// SocketDecoder decodes a WebsocketMessage
type SocketDecoder interface {
	Decode(int, []byte) (interface{}, error)
}

// SocketDecodeHandler provides an handler for SocketEncoder
type SocketDecodeHandler func(int, []byte) (interface{}, error)

//decodeBuild provides a base struct for HTTPEncoder
type socDecodeBuild struct {
	handler SocketDecodeHandler
}

// Decode provides the HTTPDecoder encode function
func (e *socDecodeBuild) Decode(d int, bu []byte) (interface{}, error) {
	return e.handler(d, bu)
}

//NewSocketDecoder provides a nice way of build an SocketDecoder
func NewSocketDecoder(fx SocketDecodeHandler) SocketDecoder {
	return &socDecodeBuild{fx}
}

//SocketCodec provides a interface define method for custom message formats
type SocketCodec interface {
	Encoder
	SocketDecoder
}

// SocketCodeco represents a generic http codec
type socketCodec struct {
	Encoder
	SocketDecoder
}

// NewSocketCodec returns a new http codec
func NewSocketCodec(e Encoder, d SocketDecoder) SocketCodec {
	return &socketCodec{e, d}
}
