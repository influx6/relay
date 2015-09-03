package relay

import (
	"net/http"

	"github.com/influx6/flux"
)

// PortHandler provides the interface member description
type PortHandler interface {
	Handle(http.ResponseWriter, *http.Request, flux.Collector)
}
