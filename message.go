package relay

import (
	"mime/multipart"
	"net/url"

	"github.com/influx6/flux"
)

// Message represent a message data
type Message struct {
	MessageType string
	Method      string
	Payload     []byte
	Form        url.Values
	PostForm    url.Values
	Multipart   *multipart.Form
	Params      flux.Collector
}
