package relay

import (
	"errors"
	"net/http"

	"github.com/gorilla/websocket"
)

// ErrPatternBound is returned when the pattern is bound
var ErrPatternBound = errors.New("Pattern is already bound")

var dupgrade = websocket.Upgrader{
	ReadBufferSize:  1024 * 1024,
	WriteBufferSize: 1024 * 1024,
}

// Controller provides a nice overlay on top of the behaviour of a requestlevel
type Controller struct {
	*Routes
}

// NewController returns a controller with the default config settings
func NewController(name string) *Controller {
	return &Controller{
		Routes: NewRoutes(name),
	}
}

// BindSocket returns provides a handle that turns http requests into websocket requests using relay.socketworkers
func (c *Controller) BindSocket(mo, pattern string, fx SocketHandler, ho http.Header) FlatChains {
	do := dupgrade
	so := NewSockets(&do, ho, fx)
	c.Add(mo, pattern, so.Handle)
	return so
}

//UpgradeSocket provides a refined control of the arguments passed to the relay.NewSocket provider
func (c *Controller) UpgradeSocket(mo, pattern string, fx SocketHandler, up websocket.Upgrader, ho http.Header) FlatChains {
	so := NewSockets(&up, ho, fx)
	c.Add(mo, pattern, so.Handle)
	return so
}

// BindHTTP binds a pattern/route to a websocket port and registers that into the controllers router, requiring the supply of a codec for handling encoding/decoding process but if not supplied uses a default http codec
func (c *Controller) BindHTTP(mo, pattern string, fx FlatHandler) FlatChains {
	hs := NewFlatChain(fx)
	c.Add(mo, pattern, hs.Handle)
	return hs
}
