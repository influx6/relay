package relay

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influx6/flux"
)

func TestRouter(t *testing.T) {
	router := NewRoutes()

	req, _ := http.NewRequest("GET", "http://localhost:3000/boo/bat", nil)

	req2, _ := http.NewRequest("POST", "http://localhost:3000/boo/post", nil)

	req3, _ := http.NewRequest("PATCH", "http://localhost:3000/boo/patch", nil)

	req4, _ := http.NewRequest("DELETE", "http://localhost:3000/boo/delete", nil)

	req5, _ := http.NewRequest("PUT", "http://localhost:3000/boo/put", nil)

	req6, _ := http.NewRequest("OPTIONS", "http://localhost:3000/boo/options", nil)

	req7, _ := http.NewRequest("HEAD", "http://localhost:3000/boo/4", nil)

	router.GET("/boo/:id", func(res http.ResponseWriter, req *http.Request, ps flux.Collector) {
		expect(t, ps.Get("id"), "bat")
	})

	router.POST("/boo/:id", func(res http.ResponseWriter, req *http.Request, ps flux.Collector) {
		expect(t, ps.Get("id"), "post")
	})

	router.PUT("/boo/:id", func(res http.ResponseWriter, req *http.Request, ps flux.Collector) {
		expect(t, ps.Get("id"), "put")
	})

	router.PATCH("/boo/:id", func(res http.ResponseWriter, req *http.Request, ps flux.Collector) {
		expect(t, ps.Get("id"), "patch")
	})

	router.DELETE("/boo/:id", func(res http.ResponseWriter, req *http.Request, ps flux.Collector) {
		expect(t, ps.Get("id"), "delete")
	})

	router.OPTIONS("/boo/:id", func(res http.ResponseWriter, req *http.Request, ps flux.Collector) {
		expect(t, ps.Get("id"), "options")
	})

	router.HEAD(`/header/{id:[\d+]}`, func(res http.ResponseWriter, req *http.Request, ps flux.Collector) {
		expect(t, ps.Get("id"), 4)
	})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	router.ServeHTTP(rec, req2)
	router.ServeHTTP(rec, req3)
	router.ServeHTTP(rec, req4)
	router.ServeHTTP(rec, req5)
	router.ServeHTTP(rec, req6)
	router.ServeHTTP(rec, req7)

}