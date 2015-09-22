package relay

import (
	"errors"
	"net/http"
)

// ErrNotHTTPParameter is returned when an HTTPort receives a wrong interface type
var ErrNotHTTP = errors.New("interface type is not HTTPRequest")

// Context provides a resource holder for each request and response pair with
//internal mappings for storing data
type Context struct {
	*SyncCollector
	Req *http.Request
	Res ResponseWriter
	// sock *SocketWorker
}

// NewContext returns a new http context
func NewContext(res http.ResponseWriter, req *http.Request) *Context {
	cx := Context{
		SyncCollector: NewSyncCollector(),
		Req:           req,
		Res:           NewResponseWriter(res),
	}
	return &cx
}

//FlatChains define a simple flat chain
type FlatChains interface {
	ChainHandleFunc(h http.HandlerFunc) FlatChains
	ChainHandler(h http.Handler) FlatChains
	ServeHTTP(http.ResponseWriter, *http.Request)
	Handle(http.ResponseWriter, *http.Request, Collector)
	HandleContext(*Context)
	Chain(FlatChains) FlatChains
	NChain(FlatChains) FlatChains
}

// NextHandler provides next call for flat chains
type NextHandler func(*Context)

// FlatHandler provides a handler for flatchain
type FlatHandler func(*Context, NextHandler)

// FlatChain provides a simple middleware like
type FlatChain struct {
	op   FlatHandler
	next FlatChains
}

//FlatChainIdentity returns a chain that calls the next automatically
func FlatChainIdentity() FlatChains {
	return NewFlatChain(func(c *Context, nx NextHandler) {
		nx(c)
	})
}

//NewFlatChain returns a new flatchain instance
func NewFlatChain(fx FlatHandler) *FlatChain {
	return &FlatChain{
		op: fx,
	}
}

// Chain sets the next flat chains else passes it down to the last chain to set as next chain,returning itself
func (r *FlatChain) Chain(rx FlatChains) FlatChains {
	if r.next == nil {
		r.next = rx
	} else {
		r.next.Chain(rx)
	}
	return r
}

// NChain sets the next flat chains else passes it down to the last chain to set as next chain,returning the the supplied chain
func (r *FlatChain) NChain(rx FlatChains) FlatChains {
	if r.next == nil {
		r.next = rx
		return rx
	}

	return r.next.NChain(rx)
}

// HandleContext calls the next chain if any
func (r *FlatChain) HandleContext(c *Context) {
	r.op(c, func(c *Context) {
		if r.next != nil {
			r.next.HandleContext(c)
		}
	})
}

// Handle calls the next chain if any
func (r *FlatChain) Handle(res http.ResponseWriter, req *http.Request, co Collector) {
	c := NewContext(res, req)
	c.Copy(co)
	r.HandleContext(c)
}

// ServeHTTP calls the next chain if any
func (r *FlatChain) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	r.Handle(res, req, nil)
}

//ChainHandleFunc returns a new flatchain using a http.HandlerFunc as a chain wrap
func (r *FlatChain) ChainHandleFunc(h http.HandlerFunc) FlatChains {
	fh := FlatChainWrap(h)
	r.Chain(fh)
	return fh
}

//ChainHandler returns a new flatchain using a http.Handler as a chain wrap
func (r *FlatChain) ChainHandler(h http.Handler) FlatChains {
	fh := FlatHandlerWrap(h)
	r.Chain(fh)
	return fh
}

//ChainFlats chains second flats to the first flatchain and returns the first flatchain
func ChainFlats(mo FlatChains, so ...FlatChains) FlatChains {
	for _, sof := range so {
		func(do FlatChains) {
			mo.Chain(do)
		}(sof)
	}
	return mo
}

//FlatHandlerWrap provides a chain wrap for http.Handler
func FlatHandlerWrap(h http.Handler) FlatChains {
	return NewFlatChain(func(c *Context, nx NextHandler) {
		h.ServeHTTP(c.Res, c.Req)
		nx(c)
	})
}

//FlatChainWrap provides a chain wrap for http.Handler
func FlatChainWrap(h http.HandlerFunc) FlatChains {
	return NewFlatChain(func(c *Context, nx NextHandler) {
		h(c.Res, c.Req)
		nx(c)
	})
}
