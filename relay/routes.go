package relay

import (
	"log"
	"net/http"
	"sync"

	"github.com/influx6/reggy"
)

// RHandler provides a custom route handler for http request with params
type RHandler func(http.ResponseWriter, *http.Request, Collector)

//WrapRouteHandlerFunc wraps http handler into a router RHandler
func WrapRouteHandlerFunc(r http.HandlerFunc) RHandler {
	return func(res http.ResponseWriter, req *http.Request, _ Collector) {
		http.Redirect(res, req, "/404", 302)
	}
}

//WrapRouteHandler wraps http handler into a router RHandler
func WrapRouteHandler(r http.Handler) RHandler {
	return func(res http.ResponseWriter, req *http.Request, _ Collector) {
		r.ServeHTTP(res, req)
	}
}

// PanicHandler provides a panic function type for requests
type PanicHandler func(http.ResponseWriter, *http.Request, interface{})

// ChainRouta provides a single routing rule instance for Router
type ChainRouta struct {
	*reggy.ClassicMatchMux
	Handler RHandler
}

// ChainRouter provides an alternative routing strategy of registered ChainRoutas using the FlatChains, its process is when any stack matches, its passed the requests to that handler and continues on, but if non matching is found,it executes a failure routine i.e it matches as many as possible unless non matches
type ChainRouter struct {
	FlatChains
	paths []*ChainRouta
	wg    sync.RWMutex
	Fail  RHandler
	Log   *log.Logger
}

// NewChainRouter returns a new ChainRouter instance
func NewChainRouter(fail RHandler, lg *log.Logger) *ChainRouter {
	sa := ChainRouter{
		FlatChains: FlatChainIdentity(lg),
		paths:      make([]*ChainRouta, 0),
		Fail:       fail,
		Log:        lg,
	}
	return &sa
}

// BareRule defines a matching rule for a specified pattern
func (r *ChainRouter) BareRule(mo, pattern string, fx RHandler) {
	methods := GetMethods(mo)
	patt := reggy.CreateClassic(pattern)
	cr := &ChainRouta{ClassicMatchMux: patt, Handler: BuildMatchesMethod(methods, fx)}
	r.wg.Lock()
	defer r.wg.Unlock()
	r.paths = append(r.paths, cr)
}

// Rule defines a matching rule which returns a flatchain
func (r *ChainRouter) Rule(mo, pattern string, fx FlatHandler) FlatChains {
	if fx == nil {
		fx = IdentityCall
	}
	methods := GetMethods(mo)
	patt := reggy.CreateClassic(pattern)
	fr := FlatRouteBuild(methods, patt, fx, r.Log)
	cr := &ChainRouta{ClassicMatchMux: patt, Handler: fr.Handle}
	r.wg.Lock()
	defer r.wg.Unlock()
	r.paths = append(r.paths, cr)
	return fr
}

// ServeHTTP provides the handling of http requests and meets the http.Handler interface
func (r *ChainRouter) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	r.FlatChains.ServeHTTP(w, rq)
	//check if any of our regexp matches
	var found bool
	var c map[string]interface{}

	for _, ra := range r.paths {
		ok, params := ra.ClassicMatchMux.Validate(rq.URL.Path)
		c = params
		if ok {
			found = true
			ra.Handler(w, rq, Collector(params))
		}
	}

	if !found {
		if r.Fail == nil {
			http.NotFound(w, rq)
		} else {
			r.Fail(w, rq, c)
		}
	}

}

// Handle meets the FlatHandler interface for serving http requests
func (r *ChainRouter) Handle(w http.ResponseWriter, rq *http.Request, _ Collector) {
	r.ServeHTTP(w, rq)
}

// Redirect redirects all incoming request to the path
func Redirect(path string) FlatChains {
	return NewFlatChain(func(c *Context, nx NextHandler) {
		http.Redirect(c.Res, c.Req, path, http.StatusTemporaryRedirect)
		nx(c)
	}, nil)
}
