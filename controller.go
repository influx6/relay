package relay

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var defaultConfig = &ControllerConfig{
	Headers: http.Header(map[string][]string{
		"Access-Control-Allow-Credentials": []string{"true"},
		"Access-Control-Allow-Origin":      []string{"*"},
	}),
	Upgrader: &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	},
	HttpCodec:   BasicHTTPCodec(),
	SocketCodec: BasicSocketCodec(),
}

//ControllerConfig provides a configuration for websockets
type ControllerConfig struct {
	Headers     http.Header
	Upgrader    *websocket.Upgrader
	HttpCodec   HTTPCodec
	SocketCodec SocketCodec
}

// Controller provides a nice overlay on top of the behaviour of a requestlevel
type Controller struct {
	config *ControllerConfig
}

// NewController returns a controller with the default config settings
func NewController() *Controller {
	return BuildController(defaultConfig)
}

// BuildController returns an instance of controller
func BuildController(config *ControllerConfig) *Controller {
	return &Controller{config: config}
}

// Websocket returns a WebsocketPort that provides a underline buffering strategy to control
//socket requests handling throttling to a specific address
func (c *Controller) Websocket(fx SocketHandler) *WebsocketPort {
	return NewWebsocketPort(c.config.SocketCodec, c.config.Upgrader, c.config.Headers, fx)
}

// HTTP returns a HTTPort that provides a underline buffering strategy to control
//requests handling throttling to a specifc address
func (c *Controller) HTTP(fx HTTPHandler) *HTTPort {
	return NewHTTPort(c.config.HttpCodec, fx)
}
