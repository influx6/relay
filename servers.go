package relay

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

type (
	//TCPKeepAliveListener provides the same internal wrapping as http keep alive functionalities
	TCPKeepAliveListener struct {
		*net.TCPListener
	}
)

var (
	//ErrTimeout provides a timeout error
	ErrTimeout = errors.New("Timeout on Connection")
	//threshold returns a chan of time
	threshold = func(ds time.Duration) <-chan time.Time {
		return time.After(ds)
	}
	//ErrorNotFind stands for errors when value not find
	ErrorNotFind = errors.New("NotFound!")
	//ErrorBadRequestType stands for errors when the interface{} recieved can not
	//be type asserted as a *http.Request object
	ErrorBadRequestType = errors.New("type is not a *http.Request")
	//ErrorBadHTTPPacketType stands for errors when the interface{} received is not a
	//bad request type
	ErrorBadHTTPPacketType = errors.New("type is not a HTTPPacket")
	//ErrorNoConnection describe when a link connection does not exists"
	ErrorNoConnection = errors.New("NoConnection")
	//ErrBadConn represent a bad connection
	ErrBadConn = errors.New("Bad Connection Received")
)

//LoadTLS loads a tls.Config from a key and cert file path
func LoadTLS(cert, key string) (*tls.Config, error) {
	var config = &tls.Config{}
	config.Certificates = make([]tls.Certificate, 1)

	c, err := tls.LoadX509KeyPair(cert, key)

	if err != nil {
		return nil, err
	}

	config.Certificates[0] = c
	return config, nil
}

//Accept sets the keep alive features of the tcp listener
func (t *TCPKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := t.AcceptTCP()
	if err != nil {
		return nil, err
	}

	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

//KeepAliveListener returns a new TCPKeepAliveListener
func KeepAliveListener(t *net.TCPListener) *TCPKeepAliveListener {
	return &TCPKeepAliveListener{t}
}

//MakeListener returns a new net.Listener for http.Request
func MakeListener(addr, ts string, conf *tls.Config) (net.Listener, error) {

	var l net.Listener
	var err error

	if conf == nil {
		l, err = tls.Listen(ts, addr, conf)
	} else {
		l, err = net.Listen(ts, addr)
	}

	if err != nil {
		return nil, err
	}

	return l, nil
}

//MakeBaseListener returns a new net.Listener(*TCPKeepAliveListener) for http.Request
func MakeBaseListener(addr string, conf *tls.Config) (net.Listener, error) {

	var l net.Listener
	var err error

	if conf != nil {
		l, err = tls.Listen("tcp", addr, conf)
	} else {
		l, err = net.Listen("tcp", addr)
	}

	if err != nil {
		return nil, err
	}

	tl, ok := l.(*net.TCPListener)

	if !ok {
		return nil, ErrBadConn
	}

	return KeepAliveListener(tl), nil
}

//MakeBaseServer returns a new http.Server using the provided listener
func MakeBaseServer(l net.Listener, handle http.Handler, c *tls.Config) (*http.Server, net.Listener, error) {

	tl, ok := l.(*net.TCPListener)

	if !ok {
		return nil, nil, fmt.Errorf("Listener is not type *net.TCPListener")
	}

	s := &http.Server{
		Addr:           tl.Addr().String(),
		Handler:        handle,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
		TLSConfig:      c,
	}

	s.SetKeepAlivesEnabled(true)
	go s.Serve(KeepAliveListener(tl))

	return s, tl, nil
}

//CreateHTTP returns a http server using the giving address
func CreateHTTP(addr string, handle http.Handler) (*http.Server, net.Listener, error) {
	l, err := net.Listen("tcp", addr)

	if err != nil {
		return nil, nil, err
	}

	return MakeBaseServer(l, handle, nil)
}

//CreateTLS returns a http server using the giving address
func CreateTLS(addr string, conf *tls.Config, handle http.Handler) (*http.Server, net.Listener, error) {
	l, err := tls.Listen("tcp", addr, conf)

	if err != nil {
		return nil, nil, err
	}

	return MakeBaseServer(l, handle, conf)
}

//LunchHTTP returns a http server using the giving address
func LunchHTTP(addr string, handle http.Handler) error {
	_, _, err := CreateHTTP(addr, handle)
	return err
}

//LunchTLS returns a http server using the giving address
func LunchTLS(addr string, conf *tls.Config, handle http.Handler) error {
	_, _, err := CreateTLS(addr, conf, handle)
	return err
}
