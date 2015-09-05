package relay

// Codecs provide a flexible approach to defining different rendering and message systems for replies without overloading the request handling code. Codecs follow the basic idea that their is an encoder and a decoder that produce and consume sets of data and handle the writing for you. This way we can create custom message patterns for response without ease and testability

// Encoders accept interface{} as the data type because this allows flexible data messages to be constructed without restrictions

//HTTPCodec provides a interface define method for custom message formats
type HTTPCodec interface {
	HTTPEncoder
	HTTPDecoder
}

//HTTPEncoder takes an inteface and encodes it into a []byte slice
type HTTPEncoder interface {
	Encode(*HTTPRequest, interface{}) (int, error)
}

//HTTPEncodeHandler provides a handler for http encoding function
type HTTPEncodeHandler func(*HTTPRequest, interface{}) (int, error)

//encodeBuild provides a base struct for HTTPEncoder
type encodeBuild struct {
	handler HTTPEncodeHandler
}

// Encode provides the HTTPEncoder encode function
func (e *encodeBuild) Encode(r *HTTPRequest, b interface{}) (int, error) {
	return e.handler(r, b)
}

//NewHTTPEncoder provides a nice way of build an HTTPEncoder
func NewHTTPEncoder(fx HTTPEncodeHandler) HTTPEncoder {
	return &encodeBuild{fx}
}

// HTTPDecoder provides a single member rule that takes a []byte and decodes it into
//its desired format
type HTTPDecoder interface {
	Decode(*HTTPRequest) (*Message, error)
}

//HTTPDecodeHandler provides a base decoder function type
type HTTPDecodeHandler func(*HTTPRequest) (*Message, error)

//decodeBuild provides a base struct for HTTPEncoder
type decodeBuild struct {
	handler HTTPDecodeHandler
}

// Decode provides the HTTPDecoder encode function
func (e *decodeBuild) Decode(r *HTTPRequest) (*Message, error) {
	return e.handler(r)
}

//NewHTTPDecoder provides a nice way of build an HTTPEncoder
func NewHTTPDecoder(fx HTTPDecodeHandler) HTTPDecoder {
	return &decodeBuild{fx}
}

// httpCodec represents a generic http codec
type httpCodec struct {
	HTTPEncoder
	HTTPDecoder
}

// NewHTTPCodec returns a new http codec
func NewHTTPCodec(e HTTPEncoder, d HTTPDecoder) HTTPCodec {
	return &httpCodec{e, d}
}

//SocketEncoder takes an the arguments and encodes it with its own given operation into a acceptable format for the websocket
type SocketEncoder interface {
	Encode(*Websocket, int, interface{}) (int, error)
}

// SocketEncodeHandler provides an handler for socketencoder
type SocketEncodeHandler func(*Websocket, int, interface{}) (int, error)

//socEncodeBuild provides a base struct for HTTPEncoder
type socEncodeBuild struct {
	handler SocketEncodeHandler
}

// Decode provides the HTTPDecoder encode function
func (e *socEncodeBuild) Encode(w *Websocket, d int, do interface{}) (int, error) {
	return e.handler(w, d, do)
}

//NewSocketEncoder provides a nice way of build an SocketEncoder
func NewSocketEncoder(fx SocketEncodeHandler) SocketEncoder {
	return &socEncodeBuild{fx}
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
	SocketEncoder
	SocketDecoder
}

// SocketCodeco represents a generic http codec
type socketCodec struct {
	SocketEncoder
	SocketDecoder
}

// NewSocketCodec returns a new http codec
func NewSocketCodec(e SocketEncoder, d SocketDecoder) SocketCodec {
	return &socketCodec{e, d}
}
