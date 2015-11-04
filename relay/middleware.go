package relay

import (
	"errors"
	"log"
	"net/http"
	"os"
)

// ErrNotHTTPParameter is returned when an HTTPort receives a wrong interface type
var ErrNotHTTP = errors.New("interface type is not HTTPRequest")

// Context provides a resource holder for each request and response pair with
//internal mappings for storing data
type Context struct {
	*SyncCollector
	Req *http.Request
	Res ResponseWriter
	Log *log.Logger
	// sock *SocketWorker
}

// NewContext returns a new http context
func NewContext(res http.ResponseWriter, req *http.Request) *Context {
	return NewContextWith(res, req, nil)
}

// NewContextWith returns a new http context with a custom logger
func NewContextWith(res http.ResponseWriter, req *http.Request, loga *log.Logger) *Context {
	if loga == nil {
		loga = log.New(os.Stdout, "[Relay] ", 0)
	}

	cx := Context{
		SyncCollector: NewSyncCollector(),
		Req:           req,
		Res:           NewResponseWriter(res),
		Log:           loga,
	}
	return &cx
}

//FlatChains define a simple flat chain
type FlatChains interface {
	ChainHandleFunc(h http.HandlerFunc) FlatChains
	ChainHandler(h http.Handler) FlatChains
	ChainFlat(h FlatHandler) FlatChains
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
	log  *log.Logger
}

//FlatChainIdentity returns a chain that calls the next automatically
func FlatChainIdentity(lg *log.Logger) FlatChains {
	return NewFlatChain(func(c *Context, nx NextHandler) {
		nx(c)
	}, lg)
}

//NewFlatChain returns a new flatchain instance
func NewFlatChain(fx FlatHandler, loga *log.Logger) *FlatChain {
	return &FlatChain{
		op:  fx,
		log: loga,
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
	r.op(c, r.handleNext)
}

// Handle calls the next chain if any
func (r *FlatChain) Handle(res http.ResponseWriter, req *http.Request, co Collector) {
	c := NewContextWith(res, req, r.log)
	c.Copy(co)
	r.HandleContext(c)
}

// ServeHTTP calls the next chain if any
func (r *FlatChain) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	r.Handle(res, req, nil)
}

//ChainHandleFunc returns a new flatchain using a http.HandlerFunc as a chain wrap
func (r *FlatChain) ChainHandleFunc(h http.HandlerFunc) FlatChains {
	fh := FlatChainWrap(h, r.log)
	r.Chain(fh)
	return fh
}

//ChainFlat returns a new flatchain using a provided FlatHandler
func (r *FlatChain) ChainFlat(h FlatHandler) FlatChains {
	fh := NewFlatChain(h, r.log)
	r.Chain(fh)
	return fh
}

//ChainHandler returns a new flatchain using a http.Handler as a chain wrap
func (r *FlatChain) ChainHandler(h http.Handler) FlatChains {
	fh := FlatHandlerWrap(h, r.log)
	r.Chain(fh)
	return fh
}

// handleNext calls the next chain in the link if available
func (r *FlatChain) handleNext(c *Context) {
	if r.next != nil {
		r.next.HandleContext(c)
	}
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

//FlatHandlerWrap provides a chain wrap for http.Handler with an optional log argument
func FlatHandlerWrap(h http.Handler, lg *log.Logger) FlatChains {
	return NewFlatChain(func(c *Context, nx NextHandler) {
		h.ServeHTTP(c.Res, c.Req)
		nx(c)
	}, lg)
}

//FlatChainWrap provides a chain wrap for http.Handler with an optional log argument
func FlatChainWrap(h http.HandlerFunc, lg *log.Logger) FlatChains {
	return NewFlatChain(func(c *Context, nx NextHandler) {
		h(c.Res, c.Req)
		nx(c)
	}, lg)
}
