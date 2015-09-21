package relay

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
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
	*websocket.Conn
	Ctx *Context
}

// WebsocketMessage provides small abstraction for processing a message
type WebsocketMessage struct {
	*Websocket
	payload []byte
	mtype   int
	Worker  *SocketWorker
}

// Message returns the data of the socket
func (m *WebsocketMessage) Message() []byte {
	return m.payload
}

// MessageType returns the type of the message
func (m *WebsocketMessage) MessageType() int {
	return m.mtype
}

// SocketWorker provides a workpool for socket connections
type SocketWorker struct {
	data    chan interface{}
	mesgs   chan *WebsocketMessage
	closer  chan bool
	queue   *flux.Queue
	wo      *Websocket
	ro, rd  sync.Mutex
	closed  bool
	writing bool
}

// NewSocketWorker returns a new socketworker instance
func NewSocketWorker(wo *Websocket) *SocketWorker {
	data := make(chan interface{})

	sw := SocketWorker{
		wo:     wo,
		closer: make(chan bool),
		data:   data,
		queue:  flux.NewQueue(data),
	}

	go sw.manage()
	return &sw
}

// Messages returns a receive only channel for socket messages
func (s *SocketWorker) Messages() <-chan *WebsocketMessage {
	if s.writing {
		return s.mesgs
	}

	flux.GoDefer("Socket:Message:Receiver", func() {
		for dag := range s.data {
			if mg, ok := dag.(*WebsocketMessage); ok {
				s.mesgs <- mg
			}
		}
	})

	return s.mesgs
}

// Socket returns the internal socket for the worker
func (s *SocketWorker) Socket() *Websocket {
	return s.wo
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
				Websocket: s.wo,
				payload:   do,
				mtype:     tp,
				Worker:    s,
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
	so      sync.RWMutex
	sockets SocketStore
	handler SocketHubHandler
	closer  chan bool
	closed  bool
}

// NewSocketHub returns a new SocketHub instance,allows the passing of a codec for encoding and decoding data
func NewSocketHub(fx SocketHubHandler) (sh *SocketHub) {
	sh = &SocketHub{
		sockets: make(SocketStore),
		handler: fx,
	}
	return
}

// Close closes the hub
func (s *SocketHub) Close() {
	if s.closed {
		return
	}
	s.closed = true
	close(s.closer)
}

// CloseNotify provides a means of checking the close state of the hub
func (s *SocketHub) CloseNotify() <-chan bool {
	return s.closer
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
			go s.handler(s, data)
		}
	}
}

//SocketHandler provides an handler type without the port option
type SocketHandler func(*SocketWorker)

// NewSockets returns a new websocket port
func NewSockets(upgrader *websocket.Upgrader, headers http.Header, hs SocketHandler) FlatChains {
	return NewFlatChain(func(c *Context, nx NextHandler) {

		if headers != nil {
			origin, ok := c.Req.Header["Origin"]

			if ok {
				headers.Set("Access-Control-Allow-Credentials", "true")
				headers.Set("Access-Control-Allow-Origin", strings.Join(origin, ";"))
			} else {
				headers.Set("Access-Control-Allow-Origin", "*")
			}
		}

		conn, err := upgrader.Upgrade(c.Res, c.Req, headers)

		if err != nil {
			return
		}

		flux.GoDefer(fmt.Sprintf("WebSocketPort.Handler"), func() {
			hs(NewSocketWorker(&Websocket{
				Conn: conn,
				Ctx:  c,
			}))
			nx(c)
		})
	})
}
