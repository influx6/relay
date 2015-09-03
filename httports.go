package relay

import (
	"errors"
	"net/http"

	"github.com/influx6/flux"
)

// ErrNotHTTPParameter is returned when an HTTPort receives a wrong interface type
var ErrNotHTTP = errors.New("interface type is not HTTPRequest")

// HTTPRequest provides a request object from a httport
type HTTPRequest struct {
	codec  HTTPCodec
	Req    *http.Request
	Res    ResponseWriter
	Params flux.Collector
}

// NewHTTPRequest returns a new HTTPMessage instance
func NewHTTPRequest(req *http.Request, res http.ResponseWriter, params flux.Collector, codec HTTPCodec) *HTTPRequest {
	return &HTTPRequest{
		Req:    req,
		Res:    NewResponseWriter(res),
		codec:  codec,
		Params: params,
	}
}

// Write encodes and writes the given data returns the (int,error) of the total writes while
func (m *HTTPRequest) Write(bw []byte) (int, error) {
	return m.codec.Encode(m, bw)
}

// Message returns the data of the socket
func (m *HTTPRequest) Message() (*Message, error) {
	msg, err := m.codec.Decode(m)
	if msg.Params == nil {
		msg.Params = m.Params
	}
	return msg, err
}

//HTTPReactorHandler is a function type for the *HTTPRequest
type HTTPReactorHandler func(*HTTPort, *HTTPRequest)

//HTTPHandler is a function type for the *HTTPRequest
type HTTPHandler func(*HTTPRequest)

// HTTPort provides a port for handling http-type request
type HTTPort struct {
	codec   HTTPCodec
	handler HTTPHandler
}

// Handle handles the reception of http request and returns a HTTPRequest object
func (h *HTTPort) Handle(res http.ResponseWriter, req *http.Request, params flux.Collector) {
	rwq := NewHTTPRequest(req, res, params, h.codec)
	h.handler(rwq)
}

// NewHTTPort returns a new http port
func NewHTTPort(codec HTTPCodec, h HTTPHandler) (hp *HTTPort) {
	hp = &HTTPort{
		codec:   codec,
		handler: h,
	}
	return
}