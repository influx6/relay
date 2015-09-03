package relay

import (
	"net/http"
	"strings"
	"sync"

	"github.com/influx6/flux"
	"github.com/influx6/reggy"
)

// RHandler provides a custom route handler for http request with params
type RHandler func(http.ResponseWriter, *http.Request, flux.Collector)

// Routable provides an interface for handling structs that accept routing
type Routable interface {
	Handle(http.ResponseWriter, *http.Request, flux.Collector)
}

//Route define a specific route with its handler
type Route struct {
	*reggy.ClassicMatchMux
	handler RHandler
	method  string
}

// PanicHandler provides a panic function type for requests
type PanicHandler func(http.ResponseWriter, *http.Request, interface{})

//Routes is the base struct for defining interlinking routes
type Routes struct {
	added        map[string]int
	routes       []*Route
	FailHandler  http.HandlerFunc
	PanicHandler PanicHandler
	ro           sync.RWMutex
}

//NewRoutes returns a new Routes instance
func NewRoutes() *Routes {
	return BuildRoutes(func(res http.ResponseWriter, req *http.Request) {
		http.NotFound(res, req)
	}, nil)
}

//BuildRoutes returns a new Routes instance
func BuildRoutes(failed http.HandlerFunc, panic PanicHandler) *Routes {
	rs := Routes{
		added:        make(map[string]int),
		FailHandler:  failed,
		PanicHandler: panic,
	}
	return &rs
}

//ServeHTTP handles the a request cycle
func (r *Routes) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	flux.RecoveryHandler("Route:ServeHTTP", func() error {
		r.ro.RLock()
		for _, no := range r.routes {
			state, params := no.Validate(req.URL.Path)
			if !state {
				continue
			}
			if no.method != "" && strings.ToLower(no.method) != strings.ToLower(req.Method) {
				break
			}
			r.wrap(no, res, req, flux.Collector(params))
			return nil
		}
		r.ro.RUnlock()
		r.doFail(res, req, nil)
		return nil
	})
}

// RouteOPTIONS sets the handler to only requests of this method
func (r *Routes) RouteOPTIONS(pattern string, h Routable) {
	r.Add("options", pattern, h.Handle)
}

// RoutePOST sets the handler to only requests of this method
func (r *Routes) RoutePOST(pattern string, h Routable) {
	r.Add("post", pattern, h.Handle)
}

// RoutePATCH sets the handler to only requests of this method
func (r *Routes) RoutePATCH(pattern string, h Routable) {
	r.Add("patch", pattern, h.Handle)
}

// RouteDELETE sets the handler to only requests of this method
func (r *Routes) RouteDELETE(pattern string, h Routable) {
	r.Add("delete", pattern, h.Handle)
}

// RoutePUT sets the handler to only requests of this method
func (r *Routes) RoutePUT(pattern string, h Routable) {
	r.Add("put", pattern, h.Handle)
}

// RouteHEAD sets the handler to only requests of this method
func (r *Routes) RouteHEAD(pattern string, h Routable) {
	r.Add("head", pattern, h.Handle)
}

// RouteGET sets the handler to only requests of this method
func (r *Routes) RouteGET(pattern string, h Routable) {
	r.Add("get", pattern, h.Handle)
}

// OPTIONS sets the handler to only requests of this method
func (r *Routes) OPTIONS(pattern string, h RHandler) {
	r.Add("options", pattern, h)
}

// POST sets the handler to only requests of this method
func (r *Routes) POST(pattern string, h RHandler) {
	r.Add("post", pattern, h)
}

// PATCH sets the handler to only requests of this method
func (r *Routes) PATCH(pattern string, h RHandler) {
	r.Add("patch", pattern, h)
}

// DELETE sets the handler to only requests of this method
func (r *Routes) DELETE(pattern string, h RHandler) {
	r.Add("delete", pattern, h)
}

// PUT sets the handler to only requests of this method
func (r *Routes) PUT(pattern string, h RHandler) {
	r.Add("put", pattern, h)
}

// HEAD sets the handler to only requests of this method
func (r *Routes) HEAD(pattern string, h RHandler) {
	r.Add("head", pattern, h)
}

// GET sets the handler to only requests of this method
func (r *Routes) GET(pattern string, h RHandler) {
	r.Add("get", pattern, h)
}

// Add adds a route into the sets of routes, method can be "" to allow all methods to be handled
func (r *Routes) Add(method, pattern string, h RHandler) {
	r.ro.Lock()
	defer r.ro.Unlock()
	if _, ok := r.added[pattern]; !ok {
		r.added[pattern] = len(r.routes)
		r.routes = append(r.routes, &Route{
			ClassicMatchMux: reggy.CreateClassic(pattern),
			handler:         h,
			method:          method,
		})
	}
}

func (r *Routes) recover(res http.ResponseWriter, req *http.Request) {
	if r.PanicHandler == nil {
		return
	}
	if err := recover(); err != nil {
		r.PanicHandler(res, req, err)
	}
}

func (r *Routes) doFail(res http.ResponseWriter, req *http.Request, ps flux.Collector) {
	if r.FailHandler == nil {
		return
	}
	defer r.recover(res, req)
	r.FailHandler(res, req)
}

func (r *Routes) wrap(rw *Route, res http.ResponseWriter, req *http.Request, ps flux.Collector) {
	defer r.recover(res, req)
	rw.handler(res, req, ps)
}
