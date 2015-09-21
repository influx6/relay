package relay

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influx6/flux"
)

func TestHTTPEncoder(t *testing.T) {
	enc := BasicHTTPCodec

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://localhost:300/foo", nil)

	rw := NewContext(res, req)

	if _, err := enc.Encode(rw.Res, []byte("home")); err != nil {
		flux.FatalFailed(t, "Encoder failed to write: %+s", err)
	}

	expect(t, res.Code, rw.Res.Status())
}

func TestHTTPDecoder(t *testing.T) {
	enc := BasicHTTPCodec

	ho := []byte("dead!")

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://localhost:300/foo", bytes.NewBuffer(ho))

	rw := NewContext(res, req)

	if _, err := enc.Encode(rw.Res, []byte("home")); err != nil {
		flux.FatalFailed(t, "Encoder failed to write: %+s", err)
	}

	bu, err := enc.Decode(rw)

	if err != nil {
		flux.FatalFailed(t, "Unable to get message %+s", err)
	}

	expect(t, string(bu.Payload), string(ho))
}
