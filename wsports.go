package relay

import (
	"errors"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/influx6/flux"
)

// ErrClosed is returned to indicated an already closed struct
var ErrClosed = errors.New("Already Closed")

// ErrInvalidType is returned when the type required is not met
var ErrInvalidType = errors.New("Unsupported Type")

// Websocket provides a cover for websocket connection
type Websocket struct {
	Conn   *websocket.Conn
	Req    *http.Request
	Res    http.ResponseWriter
	Params flux.Collector
}

// WebsocketMessage provides small abstraction for processing a message
type WebsocketMessage struct {
	codec   SocketCodec
	payload []byte
	mtype   int
	Socket  *Websocket
	Worker  *SocketWorker
}

// Message returns the data of the socket
func (m *WebsocketMessage) Message() ([]byte, error) {
	return m.codec.Decode(m.mtype, m.payload)
}

// MessageType returns the type of the message
func (m *WebsocketMessage) MessageType() int {
	return m.mtype
}

// Write encodes and writes the given data returns the (int,error) of the total writes while
func (m *WebsocketMessage) Write(bw []byte) (int, error) {
	return m.codec.Encode(m.Socket, m.mtype, bw)
}

// SocketWorker provides a workpool for socket connections
type SocketWorker struct {
	codec  SocketCodec
	data   chan interface{}
	closer chan bool
	queue  *flux.Queue
	wo     *Websocket
	ro, rd sync.Mutex
	closed bool
}

// NewSocketWorker returns a new socketworker instance
func NewSocketWorker(wo *Websocket, codec SocketCodec) *SocketWorker {
	data := make(chan interface{})

	sw := SocketWorker{
		codec:  codec,
		wo:     wo,
		closer: make(chan bool),
		data:   data,
		queue:  flux.NewQueue(data),
	}

	go sw.manage()
	return &sw
}

// Messages returns a receive only channel for socket messages
func (s *SocketWorker) Messages() <-chan interface{} {
	return s.data
}

// Equals returns true/false if interface matches
func (s *SocketWorker) Equals(b interface{}) bool {
	if sc, ok := b.(*websocket.Conn); ok {
		return s.wo.Conn == sc
	}

	if ws, ok := b.(*Websocket); ok {
		return s.wo == ws
	}

	if sw, ok := b.(*SocketWorker); ok {
		return s == sw
	}

	return false
}

// CloseNotify returns a chan that is used to notify closing of socket
func (s *SocketWorker) CloseNotify() chan bool {
	return s.closer
}

// Close closes the socket read goroutine and notifies a closure using the close channel
func (s *SocketWorker) Close() error {
	if s.closed {
		return ErrClosed
	}
	s.ro.Lock()
	s.closed = true
	close(s.closer)
	s.ro.Unlock()
	return s.wo.Conn.Close()
}

func (s *SocketWorker) manage() {
	defer s.Close()
	for {
		select {
		case <-s.closer:
			return
		default:
			tp, do, err := s.wo.Conn.ReadMessage()

			if err != nil {
				return
			}

			s.queue.Enqueue(&WebsocketMessage{
				codec:   s.codec,
				payload: do,
				mtype:   tp,
				Socket:  s.wo,
			})
		}
	}
}

// SocketStore provides a map store for SocketHub
type SocketStore map[*SocketWorker]bool

// SocketHubHandler provides a function type that encapsulates the socket hub message operations
type SocketHubHandler func(*SocketHub, *WebsocketMessage)

// SocketHub provides a central command for websocket message handling,its the base struct through which different websocket messaging procedures can be implemented on, it provides a go-routine approach,by taking each new websocket connection,stashing it then receiving data from it for processing
type SocketHub struct {
	flux.Reactor
	so      sync.RWMutex
	sockets SocketStore
	handler SocketHubHandler
}

// NewSocketHub returns a new SocketHub instance,allows the passing of a codec for encoding and decoding data
func NewSocketHub(fx SocketHubHandler) (sh *SocketHub) {
	sh = &SocketHub{
		sockets: make(SocketStore),
		handler: fx,
	}
	return
}

// AddConnection adds a new socket connection
func (s *SocketHub) AddConnection(ws *SocketWorker) {
	var ok bool

	s.so.RLock()
	ok = s.sockets[ws]
	s.so.RUnlock()

	if ok {
		return
	}

	s.so.Lock()
	s.sockets[ws] = true
	s.so.Unlock()

	go s.manageSocket(ws)
}

// SocketWorkerHandler provides a function type that encapsulates the socket workers
type SocketWorkerHandler func(*SocketWorker)

// Distribute propagates through the set of defined websocket workers and
//calls a function on it
func (s *SocketHub) Distribute(hsx SocketWorkerHandler, except *SocketWorker) {
	s.so.RLock()
	for wo := range s.sockets {
		if wo != except {
			go hsx(wo)
		}
	}
	s.so.RUnlock()
}

// manageSocket takes a socket and spawns a go-routine to manage the operations of the socket,getting the data and delivery them as WebsocketRequests
func (s *SocketHub) manageSocket(ws *SocketWorker) {
	defer func() {
		s.so.Lock()
		delete(s.sockets, ws)
		s.so.Unlock()
	}()
	for {
		select {
		case <-s.CloseNotify():
			return
		case <-ws.CloseNotify():
			return
		case data, ok := <-ws.Messages():
			if !ok {
				return
			}
			go s.handler(s, data.(*WebsocketMessage))
		}
	}
}

// WebsocketPort provides a websocket port,handling websocket connection
type WebsocketPort struct {
	flux.Reactor
	codec   SocketCodec
	upgrade *websocket.Upgrader
	headers http.Header
	handle  SocketHandler
}

// Handle handles the reception of http request and returns a HTTPRequest object
func (ws *WebsocketPort) Handle(res http.ResponseWriter, req *http.Request, params flux.Collector) {
	conn, err := ws.upgrade.Upgrade(res, req, ws.headers)

	if err != nil {
		return
	}

	ws.handle(NewSocketWorker(&Websocket{
		Conn:   conn,
		Req:    req,
		Res:    res,
		Params: params,
	}, ws.codec))
}

//SocketHandler provides an handler type without the port option
type SocketHandler func(*SocketWorker)

// NewWebsocketPort returns a new websocket port
func NewWebsocketPort(codec SocketCodec, upgrader *websocket.Upgrader, headers http.Header, hs SocketHandler) (ws *WebsocketPort) {
	ws = &WebsocketPort{
		codec:   codec,
		headers: headers,
		upgrade: upgrader,
		handle:  hs,
	}
	return
}
