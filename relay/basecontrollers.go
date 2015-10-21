package relay

import (
	"log"
	"net/http/pprof"
)

// NewPProfController provides an instantiated endpoint for pprofiles
func NewPProfController() *PProfController {
	db := &PProfController{NewController("/debug")}
	// db.BindHTTP("get head", "/pprof/profile", db.Profile)
	// db.BindHTTP("get head", "/pprof/trace", db.Trace)
	// db.BindHTTP("get head", "/pprof/symbol", db.Symbol)
	db.BindHTTP("get head", "/pprof/*", db.Index)
	return db
}

// PProfController provide pprofile handlers
type PProfController struct {
	*Controller
}

// Profile provides the pprof Profile endpoint
func (p *PProfController) Profile(c *Context, next NextHandler) {
	log.Printf("profile getting hit")
	pprof.Profile(c.Res, c.Req)
	next(c)
}

// Index provides the pprof Index endpoint
func (p *PProfController) Index(c *Context, next NextHandler) {
	pprof.Index(c.Res, c.Req)
	next(c)
}

// Symbol provides the pprof Symbol endpoint
func (p *PProfController) Symbol(c *Context, next NextHandler) {
	pprof.Symbol(c.Res, c.Req)
	next(c)
}

// Trace provides the pprof Trace endpoint
func (p *PProfController) Trace(c *Context, next NextHandler) {
	pprof.Trace(c.Res, c.Req)
	next(c)
}
