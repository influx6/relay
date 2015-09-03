package main

import (
	"log"
	"net/http"

	"github.com/influx6/relay"
)

type Pub struct {
	*relay.Controller
}

func (p *Pub) Index(req *relay.HTTPRequest) {
	m, err := req.Message()
	log.Printf("%+s %+s", m, err)
	req.Write([]byte("Welcome Home!"))
}

func main() {
	pub := &Pub{relay.NewController()}

	routes := relay.NewRoutes()

	routes.RouteGET("/", pub.HTTP(pub.Index))
	http.ListenAndServe(":3030", routes)
}
