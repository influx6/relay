package main

import (
	"net/http"

	"github.com/influx6/relay"
)

type Pub struct {
	*relay.Controller
}

func (p *Pub) Index(req *relay.Context, nxt relay.NextHandler) {
	req.Res.Write([]byte("Welcome Home!"))
	nxt(req)
}

func main() {
	pub := &Pub{relay.NewController("")}

	routes := relay.NewRoutes("")

	routes.RouteGET("/", pub.BindHTTP("", "/", pub.Index))
	http.ListenAndServe(":3030", routes)
}
