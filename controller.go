package relay

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var dheaders = http.Header(map[string][]string{
	"Access-Control-Allow-Credentials": []string{"true"},
	"Access-Control-Allow-Origin":      []string{"*"},
})

var dupgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Controller provides a nice overlay on top of the behaviour of a requestlevel
type Controller struct {
}

// NewController returns a controller with the default config settings
func NewController() *Controller {
	return &Controller{}
}

// Websocket returns a WebsocketPort that provides a underline buffering strategy to control
//socket requests handling throttling to a specific address
func (c *Controller) Websocket(fx SocketHandler) *WebsocketPort {
	return c.WebsocketAction(fx, BasicSocketCodec(), dheaders, dupgrade)
}

// HTTP returns a HTTPort that provides a underline buffering strategy to control
//requests handling throttling to a specifc address
func (c *Controller) HTTP(fx HTTPHandler) *HTTPort {
	return c.HTTPAction(fx, BasicHTTPCodec())
}

// WebsocketAction returns a WebsocketPort that provides a underline buffering strategy to control
//socket requests handling throttling to a specific address
func (c *Controller) WebsocketAction(fx SocketHandler, codec SocketCodec, headers http.Header, up websocket.Upgrader) *WebsocketPort {
	return NewWebsocketPort(codec, &up, headers, fx)
}

// HTTPAction returns a HTTPort that provides a underline buffering strategy to control
//requests handling throttling to a specifc address
func (c *Controller) HTTPAction(fx HTTPHandler, codec HTTPCodec) *HTTPort {
	return NewHTTPort(codec, fx)
}
