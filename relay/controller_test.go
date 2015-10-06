package relay

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestController(t *testing.T) {

	admin := NewController("/admin")

	admin.GET("/:id", func(res http.ResponseWriter, req *http.Request, ps Collector) {
		expect(t, ps.Get("id"), "troje_base")
	})

	req, _ := http.NewRequest("GET", "http://localhost:3000/admin/troje_base", nil)

	rec := httptest.NewRecorder()
	admin.ServeHTTP(rec, req)
}

func TestControllerBinders(t *testing.T) {
	var ws sync.WaitGroup
	ws.Add(1)

	admin := NewController("/admin")

	admin.BindHTTP("get", "/:id", func(req *Context, nx NextHandler) {
		ws.Done()
		expect(t, req.Get("id"), "troje_base")
		nx(req)
	})

	req, _ := http.NewRequest("GET", "http://localhost:3000/admin/troje_base", nil)

	rec := httptest.NewRecorder()
	admin.ServeHTTP(rec, req)

	ws.Wait()
}
