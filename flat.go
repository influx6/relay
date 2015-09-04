package relay

import "net/http"

//FlatChains define a simple flat chain
type FlatChains interface {
	ChainHandleFunc(h http.HandlerFunc) FlatChains
	ChainHandler(h http.Handler) FlatChains
	Handle(http.ResponseWriter, *http.Request, Collector)
	Chain(FlatChains) FlatChains
	NChain(FlatChains) FlatChains
}

// NextHandler provides next call for flat chains
type NextHandler func(http.ResponseWriter, *http.Request, Collector)

// FlatHandler provides a handler for flatchain
type FlatHandler func(http.ResponseWriter, *http.Request, Collector, NextHandler)

// FlatChain provides a simple middleware like
type FlatChain struct {
	op   FlatHandler
	next FlatChains
}

//FlatChainIdentity returns a chain that calls the next automatically
func FlatChainIdentity() FlatChains {
	return NewFlatChain(func(res http.ResponseWriter, req *http.Request, c Collector, nx NextHandler) {
		nx(res, req, c)
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

// Handle calls the next chain if any
func (r *FlatChain) Handle(res http.ResponseWriter, req *http.Request, c Collector) {
	r.op(res, req, c, func(res http.ResponseWriter, req *http.Request, param Collector) {
		if r.next != nil {
			r.next.Handle(res, req, c)
		}
	})
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
	return NewFlatChain(func(res http.ResponseWriter, req *http.Request, c Collector, nx NextHandler) {
		h.ServeHTTP(res, req)
		nx(res, req, c)
	})
}

//FlatChainWrap provides a chain wrap for http.Handler
func FlatChainWrap(h http.HandlerFunc) FlatChains {
	return NewFlatChain(func(res http.ResponseWriter, req *http.Request, c Collector, nx NextHandler) {
		h(res, req)
		nx(res, req, c)
	})
}
