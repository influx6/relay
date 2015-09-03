package relay

// SocketCodeco represents a generic http codec
type SocketCodeco struct {
	SocketEncoder
	SocketDecoder
}

// NewSocketCodec returns a new http codec
func NewSocketCodec(e SocketEncoder, d SocketDecoder) SocketCodec {
	return &SocketCodeco{e, d}
}

// BasicSocketCodec returns a codec using the socket encoder and decoder
func BasicSocketCodec() SocketCodec {
	return NewSocketCodec(&BasicSocketEncoder{}, &BasicSocketDecoder{})
}

// BasicSocketEncoder provides an simple encoder for websocket messages
type BasicSocketEncoder struct {
}

// Encode takes the data and type and encodes them appropriately into the socket as a reply
func (b *BasicSocketEncoder) Encode(w *Websocket, t int, bu []byte) (int, error) {
	var err error
	var size int

	err = w.Conn.WriteMessage(t, bu)

	if err == nil {
		size = len(bu)
	}

	return size, err
}

// BasicSocketDecoder provides an simple encoder for websocket messages
type BasicSocketDecoder struct{}

// Decode decodes the socket data and returns a new appropriate result as a []byte. This basic decoder just returns the slice
func (b *BasicSocketDecoder) Decode(t int, bu []byte) ([]byte, error) {
	return bu, nil
}
