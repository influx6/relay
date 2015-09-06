package relay

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influx6/flux"
)

func expect(t *testing.T, v, m interface{}) {
	if v != m {
		flux.FatalFailed(t, "Value %+v and %+v are not a match", v, m)
		return
	}
	flux.LogPassed(t, "Value %+v and %+v are a match", v, m)
}

func TestHTTPEncoder(t *testing.T) {
	enc := BasicHTTPCodec

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://localhost:300/foo", nil)

	rw := NewHTTPRequest(req, res, nil, enc)

	if _, err := enc.Encode(rw, []byte("home")); err != nil {
		flux.FatalFailed(t, "Encoder failed to write: %+s", err)
	}

	expect(t, res.Code, rw.Res.Status())
}

func TestHTTPDecoder(t *testing.T) {
	enc := BasicHTTPCodec

	ho := []byte("dead!")

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://localhost:300/foo", bytes.NewBuffer(ho))

	rw := NewHTTPRequest(req, res, nil, enc)

	if _, err := enc.Encode(rw, []byte("home")); err != nil {
		flux.FatalFailed(t, "Encoder failed to write: %+s", err)
	}

	bu, err := rw.Message()

	if err != nil {
		flux.FatalFailed(t, "Unable to get message %+s", err)
	}

	expect(t, string(bu.Payload), string(ho))
}

func TestHTTPPort(t *testing.T) {
	enc := BasicHTTPCodec

	ho := []byte("dead!")

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://localhost:300/foo", bytes.NewBuffer(ho))

	rq := NewHTTPort(enc, func(ro *HTTPRequest) {
		expect(t, ro.Req.URL.Path, req.URL.Path)
		expect(t, req.Method, ro.Req.Method)
	})

	rq.Handle(res, req, nil)
}
