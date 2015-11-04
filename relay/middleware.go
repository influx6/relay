package relay

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/influx6/reggy"
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
	fh := FlatChainHandlerWrap(h, r.log)
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

//FlatChainHandlerWrap provides a chain wrap for http.Handler with an optional log argument
func FlatChainHandlerWrap(h http.Handler, lg *log.Logger) FlatChains {
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

//FlatHandlerFuncWrap provides a chain wrap for http.Handler with an optional log argument
func FlatHandlerFuncWrap(h http.HandlerFunc) FlatHandler {
	return (func(c *Context, nx NextHandler) {
		h(c.Res, c.Req)
		nx(c)
	})
}

//FlatHandlerWrap provides a chain wrap for http.Handler with an optional log argument
func FlatHandlerWrap(h http.Handler) FlatHandler {
	return (func(c *Context, nx NextHandler) {
		h.ServeHTTP(c.Res, c.Req)
		nx(c)
	})
}

// FlatPass uses FlatRoute but passes on the next caller immediately
func FlatPass(methods, pattern string, lg *log.Logger) FlatChains {
	return FlatRoute(methods, pattern, func(c *Context, n NextHandler) { n(c) }, lg)
}

// FlatHandleHandler returns a flatchain that wraps a http.Handler
func FlatHandleHandler(methods, pattern string, r http.Handler, lg *log.Logger) FlatChains {
	return FlatRoute(methods, pattern, FlatHandlerWrap(r), lg)
}

// FlatHandleFunc returns a FlatChain that wraps a http.HandleFunc for execution
func FlatHandleFunc(methods, pattern string, r http.HandlerFunc, lg *log.Logger) FlatChains {
	return FlatRoute(methods, pattern, FlatHandlerFuncWrap(r), lg)
}

// FlatRoute provides a new routing system based on the middleware stack and if a request matches
// then its passed down the chain else ignored
func FlatRoute(methods, pattern string, fx FlatHandler, lg *log.Logger) FlatChains {
	var rmethods = GetMethods(methods)
	var rxc = reggy.CreateClassic(pattern)

	return NewFlatChain(func(c *Context, next NextHandler) {
		req := c.Req
		method := req.Method
		url := req.URL.Path

		//ok we have no method restrictions or we have the method,so we can continue else ignore
		if len(rmethods) == 0 || HasMethod(rmethods, method) {

			ok, param := rxc.Validate(url)

			// not good, so ignore
			if !ok {
				return
			}

			//ok we got a hit, copy over the parameters
			c.Copy(param)

			//call the handler with the context and next pass
			fx(c, next)
		}
	}, lg)
}

// ChooseFlat provides a binary operation for handling routing using flatchains,it inspects the requests where
// if it matches its validation parameters, the `pass` chain is called else calls the 'fail' chain if no match but still passes down the requests through the returned chain
func ChooseFlat(methods, pattern string, pass, fail FlatChains, lg *log.Logger) FlatChains {
	var rmethods = GetMethods(methods)
	var rxc = reggy.CreateClassic(pattern)

	return NewFlatChain(func(c *Context, next NextHandler) {
		req := c.Req
		res := c.Res
		method := req.Method
		url := req.URL.Path

		//ok we have no method restrictions or we have the method,so we can continue else ignore
		if len(rmethods) == 0 || HasMethod(rmethods, method) {
			ok, param := rxc.Validate(url)
			// not good, so ignore
			if !ok {
				fail.Handle(res, req, Collector(param))
			} else {
				//ok we got a hit, copy over the parameters
				pass.Handle(res, req, Collector(param))
			}
		}
		next(c)
	}, lg)
}

// ThenFlat provides a binary operation for handling routing using flatchains,it inspects the requests where
// if it matches the given criteria passes off to the supplied Chain else passes it down its own chain scope
func ThenFlat(methods, pattern string, pass FlatChains, log *log.Logger) FlatChains {
	var rmethods = GetMethods(methods)
	var rxc = reggy.CreateClassic(pattern)

	return NewFlatChain(func(c *Context, next NextHandler) {
		req := c.Req
		res := c.Res
		method := req.Method
		url := req.URL.Path

		//ok we have no method restrictions or we have the method,so we can continue else ignore
		if len(rmethods) == 0 || HasMethod(rmethods, method) {
			ok, param := rxc.Validate(url)
			// not good, so ignore
			if ok {
				//ok we got a hit, copy over the parameters
				pass.Handle(res, req, Collector(param))
			} else {
				c.Copy(param)
				next(c)
			}
		} else {
			next(c)
		}
	}, log)
}

// PanicFlatHandler returns a new FlatHandler which handles panics which may occure
func PanicFlatHandler(fx FlatHandler, p PanicHandler) FlatHandler {
	return func(c *Context, next NextHandler) {
		defer func() {
			if err := recover(); err != nil {
				p(c.Res, c.Req, err)
			}
		}()

		fx(c, next)
	}
}

// PanicContextHandler returns a new FlatHandler which handles panics which may occure
func PanicContextHandler(fx FlatHandler, p func(*Context, interface{})) FlatHandler {
	return func(c *Context, next NextHandler) {
		defer func() {
			if err := recover(); err != nil {
				p(c, err)
			}
		}()

		fx(c, next)
	}
}
