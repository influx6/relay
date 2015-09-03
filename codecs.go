package relay

//HTTPEncoder takes an inteface and encodes it into a []byte slice
type HTTPEncoder interface {
	Encode(*HTTPRequest, []byte) (int, error)
}

// HTTPDecoder provides a single member rule that takes a []byte and decodes it into
//its desired format
type HTTPDecoder interface {
	Decode(*HTTPRequest) (*Message, error)
}

//HTTPCodec provides a interface define method for custom message formats
type HTTPCodec interface {
	HTTPEncoder
	HTTPDecoder
}

//SocketEncoder takes an inteface and encodes it into a []byte slice
type SocketEncoder interface {
	Encode(*Websocket, int, []byte) (int, error)
}

// SocketDecoder decodes a WebsocketMessage
type SocketDecoder interface {
	Decode(int, []byte) ([]byte, error)
}

//SocketCodec provides a interface define method for custom message formats
type SocketCodec interface {
	SocketEncoder
	SocketDecoder
}
