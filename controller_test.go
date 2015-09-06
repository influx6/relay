package relay

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/influx6/flux"
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

	_, err := admin.BindHTTP("get", "/:id", func(req *HTTPRequest) {
		ws.Done()
		msg, err := req.Message()

		if err != nil {
			flux.FatalFailed(t, "Unable to parse message: %s", err.Error())
		}

		expect(t, msg.Params.Get("id"), "troje_base")
	}, nil)

	if err != nil {
		flux.FatalFailed(t, "Unable to add http handler: %s", err.Error())
	}

	req, _ := http.NewRequest("GET", "http://localhost:3000/admin/troje_base", nil)

	rec := httptest.NewRecorder()
	admin.ServeHTTP(rec, req)

	ws.Wait()
}
