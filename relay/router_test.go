package relay

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func BenchmarkRouter(t *testing.B) {
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		router := NewChainRouter(nil, nil)

		req, _ := http.NewRequest("GET", "http://localhost:3000/boo/bat", nil)

		req2, _ := http.NewRequest("POST", "http://localhost:3000/boo/post", nil)

		req3, _ := http.NewRequest("PATCH", "http://localhost:3000/boo/patch", nil)

		req4, _ := http.NewRequest("DELETE", "http://localhost:3000/boo/delete", nil)

		req5, _ := http.NewRequest("PUT", "http://localhost:3000/boo/put", nil)

		req6, _ := http.NewRequest("OPTIONS", "http://localhost:3000/boo/options", nil)

		req7, _ := http.NewRequest("HEAD", "http://localhost:3000/boo/4", nil)

		router.Rule("get", "/boo/:id", func(c *Context, next NextHandler) {
			time.Sleep(20 * time.Millisecond)
		})

		router.Rule("post", "/boo/:id", func(c *Context, next NextHandler) {
			time.Sleep(20 * time.Millisecond)
		})

		router.Rule("patch", "/boo/:id", func(c *Context, next NextHandler) {
			time.Sleep(20 * time.Millisecond)
		})

		router.Rule("delete", "/boo/:id", func(c *Context, next NextHandler) {
			time.Sleep(20 * time.Millisecond)
		})

		router.Rule("option", "/boo/:id", func(c *Context, next NextHandler) {
			time.Sleep(20 * time.Millisecond)
		})

		router.Rule("put", "/boo/:id", func(c *Context, next NextHandler) {
			time.Sleep(20 * time.Millisecond)
		})

		router.Rule("head", `/header/{id:[\d+]}`, func(c *Context, next NextHandler) {
			time.Sleep(20 * time.Millisecond)
		})

		router.Rule("get head  connect", "/goo/:id", func(c *Context, next NextHandler) {
			time.Sleep(20 * time.Millisecond)
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
}

func TestRouter(t *testing.T) {
	router := NewChainRouter(nil, nil)

	req, _ := http.NewRequest("GET", "http://localhost:3000/boo/bat", nil)

	req2, _ := http.NewRequest("POST", "http://localhost:3000/boo/post", nil)

	req3, _ := http.NewRequest("PATCH", "http://localhost:3000/boo/patch", nil)

	req4, _ := http.NewRequest("DELETE", "http://localhost:3000/boo/delete", nil)

	req5, _ := http.NewRequest("PUT", "http://localhost:3000/boo/put", nil)

	req6, _ := http.NewRequest("OPTIONS", "http://localhost:3000/boo/options", nil)

	req7, _ := http.NewRequest("HEAD", "http://localhost:3000/boo/4", nil)

	router.Rule("bat", "/boo/:id", func(c *Context, next NextHandler) {
		expect(t, c.Get("id"), "bat")
	})

	router.Rule("post", "/boo/:id", func(c *Context, next NextHandler) {
		expect(t, c.Get("id"), "post")
	})

	router.Rule("put", "/boo/:id", func(c *Context, next NextHandler) {
		expect(t, c.Get("id"), "put")
	})

	router.Rule("patch", "/boo/:id", func(c *Context, next NextHandler) {
		expect(t, c.Get("id"), "patch")
	})

	router.Rule("delete", "/boo/:id", func(c *Context, next NextHandler) {
		expect(t, c.Get("id"), "delete")
	})

	router.Rule("option", "/boo/:id", func(c *Context, next NextHandler) {
		expect(t, c.Get("id"), "options")
	})

	router.Rule("head", `/header/{id:[\d+]}`, func(c *Context, next NextHandler) {
		expect(t, c.Get("id"), 4)
	})

	router.Rule("get head  connect", "/goo/:id", func(c *Context, next NextHandler) {
		// expect(t, c.Get("id"), "options")
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
