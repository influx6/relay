package relay

import (
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/influx6/flux"
	"github.com/influx6/reggy"
)

// RHandler provides a custom route handler for http request with params
type RHandler func(http.ResponseWriter, *http.Request, Collector)

//WrapRouteHandlerFunc wraps http handler into a router RHandler
func WrapRouteHandlerFunc(r http.HandlerFunc) RHandler {
	return func(res http.ResponseWriter, req *http.Request, _ Collector) {
		r(res, req)
	}
}

//WrapRouteHandler wraps http handler into a router RHandler
func WrapRouteHandler(r http.Handler) RHandler {
	return func(res http.ResponseWriter, req *http.Request, _ Collector) {
		r.ServeHTTP(res, req)
	}
}

// Routable provides an interface for handling structs that accept routing
type Routable interface {
	Handle(http.ResponseWriter, *http.Request, Collector)
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
	namespace    string
	added        map[string]int
	routes       []*Route
	FailHandler  http.HandlerFunc
	PanicHandler PanicHandler
	ro           sync.RWMutex
}

//NewRoutes returns a new Routes instance
func NewRoutes(ns string) *Routes {
	return BuildRoutes(ns, nil, nil)
}

//BuildRoutes returns a new Routes instance
func BuildRoutes(ns string, failed http.HandlerFunc, panic PanicHandler) *Routes {
	rs := Routes{
		namespace:    ns,
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
		defer r.ro.RUnlock()
		mod := strings.ToLower(req.Method)
		for _, no := range r.routes {
			if no.method == "" || strings.ToLower(no.method) == mod {
				var ro string

				// check if namespace is not empty then combine the namespace else just use url path

				if r.namespace != "" {
					ro = fmt.Sprintf("%s/%s", r.namespace, req.URL.Path)
				} else {
					ro = req.URL.Path
				}

				// state, params := no.Validate(req.URL.Path)
				state, params := no.Validate(ro)
				if !state {
					continue
				}
				r.wrap(no, res, req, Collector(params))
				// break
				return nil
			}
		}
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

// RouteWithMethods sets the handler to only requests of this method
func (r *Routes) RouteWithMethods(pattern string, methods []string, h Routable) {
	r.WithMethods(pattern, methods, h.Handle)
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

// WithMethods provides a route handler for dealing with routes that can handle more than one method
func (r *Routes) WithMethods(pattern string, methods []string, h RHandler) {
	r.Add("", pattern, func(res http.ResponseWriter, req *http.Request, c Collector) {

		mo := strings.ToLower(req.Method)
		var found bool

		for _, amo := range methods {
			if mo == strings.ToLower(amo) {
				found = true
				break
			}
			continue
		}

		if !found {
			r.doFail(res, req, c)
			return
		}

		h(res, req, c)
	})
}

// ServeDir serves up a directory to the request
func (r *Routes) ServeDir(pattern, dir, strip string) {
	r.WithMethods(pattern+"/*", []string{"get", "head"}, func(res http.ResponseWriter, req *http.Request, c Collector) {

		requested := reggy.CleanPath(req.URL.Path)
		file := strings.TrimPrefix(requested, strip)

		if err := ServeFile("index.html", dir, file, res, req); err != nil {
			r.doFail(res, req, c)
		}

	})
}

// ServeDirIndex serves up a directory to the request
func (r *Routes) ServeDirIndex(pattern, indexFile, dir, strip string) {
	r.WithMethods(pattern+"/*", []string{"get", "head"}, func(res http.ResponseWriter, req *http.Request, c Collector) {

		// _, file := path.Split(pao)
		requested := reggy.CleanPath(req.URL.Path)
		file := strings.TrimPrefix(requested, strip)

		if err := ServeFile(indexFile, dir, file, res, req); err != nil {
			r.doFail(res, req, c)
		}

	})
}

//ServeFile adds a only route for handling file requests
func (r *Routes) ServeFile(pattern, file string) {
	r.Add("", pattern, func(res http.ResponseWriter, req *http.Request, c Collector) {

		dir, file := path.Split(file)

		mo := strings.ToLower(req.Method)

		if mo != "get" && mo != "head" {
			r.doFail(res, req, c)
			return
		}

		if err := ServeFile("", dir, file, res, req); err != nil {
			r.doFail(res, req, c)
		}

	})
}

// Add adds a route into the sets of routes, method can be "" to allow all methods to be handled
func (r *Routes) Add(mo, pattern string, h RHandler) {
	r.ro.Lock()
	defer r.ro.Unlock()
	if _, ok := r.added[pattern]; !ok {
		r.added[pattern] = len(r.routes)
		r.routes = append(r.routes, &Route{
			ClassicMatchMux: reggy.CreateClassic(pattern),
			handler:         h,
			method:          mo,
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

func (r *Routes) doFail(res http.ResponseWriter, req *http.Request, ps Collector) {
	defer r.recover(res, req)
	if r.FailHandler == nil {
		http.NotFound(res, req)
	} else {
		r.FailHandler(res, req)
	}
}

func (r *Routes) wrap(rw *Route, res http.ResponseWriter, req *http.Request, ps Collector) {
	defer r.recover(res, req)
	rw.handler(res, req, ps)
}
