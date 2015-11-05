package relay

import "net/http/pprof"

// PProf provide pprofile handlers
type PProf struct{}

// NewPProf provides an instantiated endpoint for pprofiles
func NewPProf(r *ChainRouter) *PProf {
	db := &PProf{}
	r.Rule("get head", "/debug/pprof/*", db.Index)
	return db
}

// Index provides the pprof Index endpoint
func (p *PProf) Index(c *Context, next NextHandler) {
	pprof.Index(c.Res, c.Req)
	next(c)
}

// Profile provides the pprof Profile endpoint
func (p *PProf) Profile(c *Context, next NextHandler) {
	pprof.Profile(c.Res, c.Req)
	next(c)
}

// Symbol provides the pprof Symbol endpoint
func (p *PProf) Symbol(c *Context, next NextHandler) {
	pprof.Symbol(c.Res, c.Req)
	next(c)
}

// Trace provides the pprof Trace endpoint
func (p *PProf) Trace(c *Context, next NextHandler) {
	pprof.Trace(c.Res, c.Req)
	next(c)
}
