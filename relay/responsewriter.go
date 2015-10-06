package relay

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

// modes.go contains specific modification structs or decorators over standard
//go interfaces for useful bits

// ResponseWriter provides a clean interface decorated over the http.ResponseWriter
type ResponseWriter interface {
	http.ResponseWriter
	http.Flusher
	http.Hijacker
	// http.CloseNotifier
	Status() int
	Size() int
	Written() bool
	WritePayload(int, []byte) error
}

// responseWriter provides the concrete implementation of ResponseWriter
type responseWriter struct {
	w      http.ResponseWriter
	status int
	size   int
}

// NewResponseWriter returns a new responseWriter
func NewResponseWriter(w http.ResponseWriter) ResponseWriter {
	rw := responseWriter{w: w}
	return &rw
}

// ErrNotHijackable is returned when a response writer can not be hijacked
var ErrNotHijackable = errors.New("ResponseWriter cant be Hijacked")

// Hijack checks if it can hijack the internal http.ResponseWriter
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hw, ok := rw.w.(http.Hijacker); ok {
		return hw.Hijack()
	}
	return nil, nil, ErrNotHijackable
}

// WritePayload writes a payload ignoring the type
func (rw *responseWriter) WritePayload(c int, p []byte) error {
	_, err := rw.w.Write(p)
	return err
}

// WriteHeader writes the status code for the http response
func (rw *responseWriter) WriteHeader(c int) {
	if rw.Written() {
		return
	}
	rw.status = c
	rw.w.WriteHeader(c)
}

// Header returns the response header
func (rw *responseWriter) Header() http.Header {
	return rw.w.Header()
}

// Write writes the supplied data into the internal resposne writer
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.Written() {
		rw.status = http.StatusOK
	}

	n, err := rw.w.Write(b)
	rw.size += n
	return n, err
}

// Status returns the status of the resposne,defaults to 200 OK
func (rw *responseWriter) Status() int {
	return rw.status
}

// Flush calls the http.Flusher flush method
func (rw *responseWriter) Flush() {
	if fw, ok := rw.w.(http.Flusher); ok {
		fw.Flush()
	}
}

// CloseNotify returns a receive-only channel
func (rw *responseWriter) CloseNotify() <-chan bool {
	return rw.w.(http.CloseNotifier).CloseNotify()
}

// Written returns true/false if the status code has been written
func (rw *responseWriter) Written() bool {
	return rw.status != 0
}

// Size returns the size of written data
func (rw *responseWriter) Size() int {
	return rw.size
}
