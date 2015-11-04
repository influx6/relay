package relay

import (
	"log"
	"net/http"
	"time"
)

// LoggerHandler creates a new FlatHandler using a giving log instance
func LoggerHandler() FlatHandler {
	return func(c *Context, next NextHandler) {
		start := time.Now()
		req := c.Req
		res := c.Res

		addr := req.Header.Get("X-Real-IP")
		if addr == "" {
			addr = req.Header.Get("X-Forwarded-For")
			if addr == "" {
				addr = req.RemoteAddr
			}
		}

		c.Log.Printf("Started %s %s for %s", req.Method, req.URL.Path, addr)

		rw := res.(ResponseWriter)
		next(c)

		c.Log.Printf("Completed %s(::%s) -> %v %s in %v @ %s\n", req.URL.Path, req.Method, rw.Status(), http.StatusText(rw.Status()), time.Since(start), addr)
	}
}

// Logger returns a new logger chain for logger incoming requests using a custom logger
func Logger(log *log.Logger) FlatChains {
	return NewFlatChain(LoggerHandler(), log)
}
