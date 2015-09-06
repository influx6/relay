package relay

import (
	"errors"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/influx6/ds"
)

var dheaders = http.Header(map[string][]string{
	"Access-Control-Allow-Credentials": []string{"true"},
	"Access-Control-Allow-Origin":      []string{"*"},
})

var dupgrade = websocket.Upgrader{
	ReadBufferSize:  1024 * 1024,
	WriteBufferSize: 1024 * 1024,
}

type portnet struct {
	id   string
	port PortHandler
}

type netport struct {
	path string
	port interface{}
}

// Equals checks if the given value is equal
func (n *netport) Equals(v interface{}) bool {
	if nop, ok := v.(*netport); ok {
		return nop.path == n.path
	}

	if sop, ok := v.(string); ok {
		return n.path == sop
	}

	return false
}

// Controller provides a nice overlay on top of the behaviour of a requestlevel
type Controller struct {
	*Routes
	sockets, https ds.Sets
	tag            string
}

// NewController returns a controller with the default config settings
func NewController(name string) *Controller {
	return &Controller{
		Routes:  NewRoutes(name),
		tag:     name,
		sockets: ds.SafeSet(),
		https:   ds.SafeSet(),
	}
}

// GetSocket gets the websocket port for the specific pattern
func (c *Controller) GetSocket(pattern string) (*WebsocketPort, error) {
	ind, ok := c.sockets.Find(pattern)
	if !ok {
		return nil, ErrorNotFind
	}
	port := c.sockets.Get(ind).(*netport).port
	return port.(*WebsocketPort), nil
}

// GetHTTP gets the http port for the specific pattern
func (c *Controller) GetHTTP(pattern string) (*HTTPort, error) {
	ind, ok := c.https.Find(pattern)
	if !ok {
		return nil, ErrorNotFind
	}
	port := c.https.Get(ind).(*netport).port
	return port.(*HTTPort), nil
}

// BindSocket returns a WebsocketPort that provides a underline buffering strategy to control socket requests handling throttling to a specific address. It requries the supply of a codec but if not supplied uses a default socket codec
func (c *Controller) BindSocket(mo, pattern string, fx SocketHandler, codec SocketCodec) (*WebsocketPort, error) {
	if codec == nil {
		codec = BasicSocketCodec
	}
	return c.BindSocketFor(mo, pattern, fx, codec, dupgrade, dheaders)
}

// ErrPatternBound is returned when the pattern is bound
var ErrPatternBound = errors.New("Pattern is already bound")

// BindHTTP binds a pattern/route to a websocket port and registers that into the controllers router, requiring the supply of a codec for handling encoding/decoding process but if not supplied uses a default http codec
func (c *Controller) BindHTTP(mo, pattern string, fx HTTPHandler, codec HTTPCodec) (*HTTPort, error) {
	if c.HasRoute(pattern) {
		return nil, ErrPatternBound
	}

	if codec == nil {
		codec = BasicHTTPCodec
	}

	hs := NewHTTPort(codec, fx)
	c.sockets.Add(&netport{pattern, hs}, -1)

	c.Add(mo, pattern, hs.Handle)
	return hs, nil
}

// BindSocketFor binds a pattern/route to a websocket port and registers that into the controllers router and allows a more refined and control configuration for the socket connection
func (c *Controller) BindSocketFor(mo, pattern string, fx SocketHandler, codec SocketCodec, up websocket.Upgrader, headers http.Header) (*WebsocketPort, error) {
	if c.HasRoute(pattern) {
		return nil, ErrPatternBound
	}

	ws := NewWebsocketPort(codec, &up, headers, fx)
	c.sockets.Add(&netport{pattern, ws}, -1)
	c.Add(mo, pattern, ws.Handle)
	return ws, nil
}
