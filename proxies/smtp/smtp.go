package smtp

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/smtp"
	"net/textproto"
	"strings"
	"sync/atomic"
	"time"

	"code.google.com/p/go-uuid/uuid"

	"github.com/influx6/flux"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("flux")

type (

	//Envelope entails the email address and attachments
	Envelope struct {
		ID          uuid.UUID
		Sender      string
		Receivers   []string
		AuthData    string
		AuthType    string
		AuthRequest string
		Data        []byte
		Peer        SMTPPeer
	}

	//Service defines a service provider
	Service interface {
		Serve()
	}

	//SMTPService provides a servicer for smtp deployment
	SMTPService struct {
		listener  net.Listener
		tlsconfig *tls.Config
		config    SMTPConfig
	}

	//DeferConn provides a means of proxying
	DeferConn struct {
		net.Conn
		TargetConn net.Conn
		config     *tls.Config
	}

	//SMTPProxyConn provides a low-level smtp processor of a net.Conn
	SMTPProxyConn struct {
		*DeferConn
		c SMTPConfig
	}

	//DeferListener provides a custom listener generator
	DeferListener struct {
		net.Listener
		targetAddr string
		network    string
		config     *tls.Config
	}

	//SMTPRecordMaker provides a means of providing a record handler
	SMTPRecordMaker func([]byte) SMTPRecord

	//SMTPRecordHandler provides a means of providing a record handler
	SMTPRecordHandler func(*SMTPRecord) error

	//SMTPPeer provides peer data
	SMTPPeer struct {
		HeloName   string
		Username   string
		Password   string
		ServerName string
		Protocol   Protocol
		Addr       net.Addr
		TLS        tls.ConnectionState `yaml:"-" json:"-"`
	}

	//MetaConfig provides a config struct
	MetaConfig struct {
		Hostname       string
		WelcomeMessage string
		MaxSize        int //10240000
		Peer           SMTPPeer
		Features       *Capability
		Config         *tls.Config
		Timeouts       *Fsop
	}

	//SMTPConfig provides a config struct
	SMTPConfig struct {
		MetaConfig
		Conn net.Conn
	}

	//SMTPSession provides a base level content handler for smtp sessions from the server
	SMTPSession struct {
		base     *SMTPConfig
		target   net.Conn
		reader   *bufio.Reader
		writer   *bufio.Writer
		envelope *Envelope

		scanner  *bufio.Scanner
		Timeouts *Fsop
		closenow int64
		tlsd     bool
	}

	//SMTPRecordChannel represents a channel of smtp record
	SMTPRecordChannel chan SMTPRecord

	//SMTPRecord provides a record struct
	SMTPRecord struct {
		Peer     SMTPPeer
		Packet   []byte
		Cmd      SMTPCommand
		Session  *SMTPSession
		Finished chan struct{}
	}

	//Fsop provides time details for read/write deadlines
	Fsop struct {
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
		DataTimeout  time.Duration
		CloseTimeout time.Duration
	}

	//Protocol defines a connection protocol type
	Protocol string

	//Capability provide a means of decided session extensions
	Capability struct {
		TLS            bool
		Authentication bool
		XClient        bool
	}

	//SMTPError represents a stmp error
	SMTPError struct {
		Code    int
		Message string
	}

	//SMTPCommand represents a current command
	SMTPCommand struct {
		Packet string
		Action string
		Fields []string
		Params []string
		From   string
		To     string
	}
)

//Error returns the string message of the error
func (e SMTPError) Error() string {
	return fmt.Sprintf("%d %s", e.Code, e.Message)
}

const (
	//SMTP protocol format
	SMTP Protocol = "SMTP"
	//ESMTP protocol format
	ESMTP Protocol = "ESMTP"
)

var (
	// notready        = false
	defaultFeatures = &Capability{false, false, false}

	defaultFsop = &Fsop{
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		DataTimeout:  1 * time.Minute,
		CloseTimeout: 200 * time.Millisecond,
	}

	//ErrShortResponse represents a bad smtp connection
	ErrShortResponse = errors.New("Response Too Short")
	//ErrBadSMTP represents a bad smtp connection
	ErrBadSMTP = errors.New("SMTP Unavailable")
	//ErrBadSMTPRequest represents a bad smtp connection
	ErrBadSMTPRequest = errors.New("SMTP Request Failed")
)

//Serve handles the operations of the servicer
func (s *SMTPService) Serve() {
	defer s.listener.Close()
	for {

		con, err := s.listener.Accept()

		if err != nil {
			log.Error("SMTPService Listener Error", err)
			continue
		}

		dcon, ok := con.(*DeferConn)

		if !ok {
			log.Error("SMTPService ConnType", con.Close())
			continue
		}

		flux.GoDefer("SMTPDeferConOp", func() {

			conf := s.config
			conf.Config = s.tlsconfig

			log.Debug("Creating SMTPProxyConnector for Remote:%s Local:%s", dcon.RemoteAddr().String(), dcon.LocalAddr().String())

			err = (SMTPProxyConn{
				dcon,
				conf,
			}).Proxy()

			if err != nil {
				log.Error("Proxy Error for %s", con.RemoteAddr().String(), err)
			}
		})
	}
}

//NewSMTP returns a new smtp service provder (SMTPService)
func NewSMTP(meta MetaConfig, target, host string, config *tls.Config) (so *SMTPService, err error) {
	var ls *DeferListener

	ls, err = NewListener("tcp", target, host, config)

	if err != nil {
		return
	}

	// addr, _, _ := net.SplitHostPort(ls.Addr().String())

	so = &SMTPService{
		listener:  ls,
		tlsconfig: config,
		config:    SMTPConfig{MetaConfig: meta},
	}

	return
}

//Accept handles the Accept operations of the listener
func (l *DeferListener) Accept() (net.Conn, error) {
	con, err := l.Listener.Accept()

	if err != nil {
		return nil, err
	}

	con2, err := BuildConn(l.network, l.targetAddr, l.config)

	if err != nil {
		log.Error("Failed to connect target %s by %s", l.targetAddr, con.LocalAddr().String(), con.Close())
		return nil, err
	}

	return NewConn(con, con2, l.config), nil
}

//NewConn returns a new deferConn
func NewConn(host, target net.Conn, utls *tls.Config) *DeferConn {
	return &DeferConn{host, target, utls}
}

//BuildConn returns a new net.Conn
func BuildConn(network, addr string, conf *tls.Config) (con net.Conn, err error) {
	if conf != nil {
		con, err = tls.Dial(network, addr, conf)
	} else {
		con, err = net.Dial(network, addr)
	}
	return
}

//BuildListener returns a new net.Listener
func BuildListener(netc, addr string, config *tls.Config) (host net.Listener, err error) {
	if config == nil {
		host, err = net.Listen(netc, addr)
	} else {
		host, err = tls.Listen(netc, addr, config)
	}

	if err != nil {
		return nil, err
	}

	return
}

//NewListener returns a new DeferListener
func NewListener(nc, targetAddr, hostAddr string, config *tls.Config) (*DeferListener, error) {
	ls, err := BuildListener(nc, hostAddr, config)

	if err != nil {
		return nil, err
	}

	return &DeferListener{
		Listener:   ls,
		targetAddr: targetAddr,
		network:    nc,
		config:     config,
	}, nil
}

//Proxy handles a proxy operation
func (s SMTPProxyConn) Proxy( /*c *ProxyConfig*/ ) error {
	// if notready {
	// 	return ErrBadSMTP
	// }

	addr, _, _ := net.SplitHostPort(s.TargetConn.RemoteAddr().String())
	client, err := smtp.NewClient(s.TargetConn, addr)

	if err != nil {
		return err
	}

	_ = client

	config := s.c
	config.Conn = s.Conn

	session := NewSMTPSession(&config, s.TargetConn)

	//close the session
	defer session.Close()

	//lets start proxy the commands between the clients
	for rec := range session.Serve() {
		// _ = cmd

		cmd := rec.Cmd
		log.Debug("Cmd: %s, Fields: %s Data: %s Line: %s", cmd.Action, cmd.Fields, cmd.Params, cmd.Packet)

		// the plan is to recieve our Commands(cmd) then take the data
		//and send that to our the client that delivers that to the smtp Server
		//on the container ,take the response and record then send that back to the
		//requester

		err := smtpHandlers(&rec, client, session)
		log.Error("Handling SMTP Request %s", cmd.Action, err)

		if err != nil {
			session.Reply(502, fmt.Sprintf("Request %s Failed for proxy", cmd.Action))
		}

		close(rec.Finished)
	}

	return nil
}

//NewSMTPSession returns a new smtp session instance
func NewSMTPSession(conf *SMTPConfig, target net.Conn) (s *SMTPSession) {
	// conf := &config

	if conf.Timeouts == nil {
		conf.Timeouts = defaultFsop
	}

	if conf.MaxSize == 0 {
		conf.MaxSize = 1024000
	}

	if conf.Hostname == "" {
		conf.Hostname = "localhost.localdomain"
	}

	if conf.WelcomeMessage == "" {
		proc := conf.Peer.Protocol
		if proc == "" {
			proc = ESMTP
		}
		conf.WelcomeMessage = fmt.Sprintf("%s %s Ready!", conf.Hostname, proc)
	}

	s = &SMTPSession{
		base:   conf,
		target: target,
		reader: bufio.NewReader(conf.Conn),
		writer: bufio.NewWriter(conf.Conn),
		envelope: &Envelope{
			ID:   uuid.NewUUID(),
			Peer: conf.Peer,
		},
	}

	s.scanner = bufio.NewScanner(s.reader)

	// go s.serve()
	return
}

//Welcome sends a welcome message for the session
func (s *SMTPSession) Welcome() {
	s.Reply(220, s.base.WelcomeMessage)
}

//Reject sends a rejecrt message for the session
func (s *SMTPSession) Reject() {
	s.Reply(421, "Too busy.Try again Later")
}

//RejectClose sends a rejecrt message for the session and closes the session
func (s *SMTPSession) RejectClose() {
	s.Reject()
	s.Close()
}

//Close the session
func (s *SMTPSession) Close() error {
	s.writer.Flush()
	time.Sleep(s.base.Timeouts.CloseTimeout)
	atomic.StoreInt64(&s.closenow, 1)
	s.Reset()
	s.envelope = nil
	return s.base.Conn.Close()
}

//Reset resets the session
func (s *SMTPSession) Reset() {
	s.envelope = &Envelope{
		ID:   uuid.NewUUID(),
		Peer: s.base.Peer,
	}
}

func (s *SMTPSession) flush() error {
	s.base.Conn.SetWriteDeadline(time.Now().Add(s.base.Timeouts.WriteTimeout))
	err := s.writer.Flush()
	s.base.Conn.SetReadDeadline(time.Now().Add(s.base.Timeouts.ReadTimeout))
	return err
}

func (s *SMTPSession) readyData() {
	s.base.Conn.SetReadDeadline(time.Now().Add(s.base.Timeouts.DataTimeout))
}

func setDeadline(c net.Conn, d time.Duration) {
	c.SetDeadline(time.Now().Add(d))
}

func setReadDeadline(c net.Conn, d time.Duration) {
	c.SetReadDeadline(time.Now().Add(d))
}

func setWriteDeadline(c net.Conn, d time.Duration) {
	c.SetWriteDeadline(time.Now().Add(d))
}

func (s *SMTPSession) readyDataWrite() {
	s.base.Conn.SetWriteDeadline(time.Now().Add(s.base.Timeouts.DataTimeout))
}

func (s *SMTPSession) readyRead() {
	s.base.Conn.SetReadDeadline(time.Now().Add(s.base.Timeouts.ReadTimeout))
}

//Extensions returns the extensions of this session
func (s *SMTPSession) Extensions() (ext []string) {

	if s.base.Features.XClient {
		ext = append(ext, "XCLIENT")
	}

	if s.base.Features.Authentication {
		ext = append(ext, "AUTH PLAIN LOGIN")
	}

	if s.base.Features.TLS {
		ext = append(ext, "STARTTLS")
	}

	return
}

//Reply emits a response to the connection
func (s *SMTPSession) Reply(code int, message string) {
	fmt.Fprintf(s.writer, "%d %s\r\n", code, message)
	s.flush()
}

//Error emits a response to the connection
func (s *SMTPSession) Error(err error) {
	if se, ok := err.(SMTPError); ok {
		s.Reply(se.Code, se.Message)
		return
	}
	s.Reply(502, fmt.Sprintf("%s", err))
}

//Serve the session connection,returns a channel that responds
func (s *SMTPSession) Serve() SMTPRecordChannel {
	records := make(SMTPRecordChannel)

	go func() {
		// defer s.Close()
		defer close(records)

		log.Debug("Serving Session for %s", s.base.Hostname)
		s.Welcome()
		log.Debug("Sending Welcome for %s", s.base.Hostname)

		for {
			if atomic.LoadInt64(&s.closenow) > 0 {
				break
			}

			for s.scanner.Scan() {
				pack := s.scanner.Bytes()
				rec := SMTPRecord{
					Peer:     s.base.Peer,
					Packet:   pack,
					Session:  s,
					Cmd:      parseCommand(pack),
					Finished: make(chan struct{}),
				}

				records <- rec

				<-rec.Finished
			}

			err := s.scanner.Err()

			if err == bufio.ErrTooLong {
				s.Reply(500, "Line Too Long")

				//advance reader to next line
				s.reader.ReadString('\n')
				s.scanner = bufio.NewScanner(s.reader)

				//reset
				// s.reset()

				continue
			}

			break
		}
	}()

	return records
}

func validLine(line string) bool {
	if len(line) < 4 || line[3] != ' ' && line[3] != '-' {
		return false
	}
	return true
}

func collectLines(c *textproto.Conn) ([]string, error) {
	var ex error
	var line string
	var msg []string
	var ended bool

	line, ex = c.ReadLine()

	if ex != nil {
		return nil, ex
	}

	if !validLine(line) {
		return nil, ErrShortResponse
	}

	msg = append(msg, line)

	ended = line[3] != '-'

	for ex == nil && !ended {
		line, ex = c.ReadLine()

		if ex != nil {
			return nil, ex
		}

		msg = append(msg, line)
		ended = line[3] != '-'
	}

	return msg, nil
}

func repairLines(msg []string) string {
	return fmt.Sprintf("%s\r\n", strings.Join(msg, "\n"))
}

func collectClientLines(c *textproto.Conn) (string, error) {
	msg, err := collectLines(c)

	if err != nil {
		return "", err
	}

	return repairLines(msg), nil
}

//SendCmd sends a command to a client
func SendCmd(cl *smtp.Client) (int, string, error) {
	var code int
	var err error
	var msg string

	return code, msg, err
}

func smtpHandlers(rec *SMTPRecord, c *smtp.Client, proxy *SMTPSession) error {

	switch rec.Cmd.Action {

	case "HELO":
		return handleHELO(rec, c, proxy)

	case "EHLO":
		return handleEHLO(rec, c, proxy)

	case "MAIL":
		return handleMAIL(rec, c, proxy)

	case "RCPT":
		return handleRCPT(rec, c, proxy)

	case "STARTTLS":
		return handleSTARTTLS(rec, c, proxy)

	case "DATA":
		return handleDATA(rec, c, proxy)

	case "RSET":
		return handleRSET(rec, c, proxy)

	case "NOOP":
		return handleNOOP(rec, c, proxy)

	case "QUIT":
		return handleQUIT(rec, c, proxy)

	case "AUTH":
		return handleAUTH(rec, c, proxy)

	case "XCLIENT":
		return handleXCLIENT(rec, c, proxy)

	}

	rec.Session.Reply(502, "Unsupported command.")
	return ErrBadSMTPRequest
}

func handleDATA(rec *SMTPRecord, c *smtp.Client, proxy *SMTPSession) error {

	proxy.Reply(354, "Go Ahead.")
	setDeadline(proxy.base.Conn, proxy.base.Timeouts.DataTimeout)

	buf := &bytes.Buffer{}

	reader := textproto.NewReader(proxy.reader).DotReader()

	_, err := io.CopyN(buf, reader, int64(proxy.base.MaxSize))

	if err == io.EOF || err == nil {

		proxy.Reply(250, "Thank you.")

		bu := buf.Bytes()

		datawriter, err := c.Data()

		if err != nil {
			return err
		}

		n, err := datawriter.Write(bu)
		log.Debug("Data Submitted total written %d!", n)

		if err != nil {
			return err
		}

		err = datawriter.Close()

		if err != nil {
			return err
		}

		proxy.Reset()
		return nil
	}

	if err != nil {
		proxy.Reply(254, "Unable to recieve data.")
		return err
	}

	_, err = io.Copy(ioutil.Discard, reader)

	proxy.Reply(552, fmt.Sprintf("Message exceeded max message size of %d bytes", proxy.base.MaxSize))

	if err != nil {
		return err
	}

	return nil
}

func handleAUTH(rec *SMTPRecord, c *smtp.Client, proxy *SMTPSession) error {

	id, err := c.Text.Cmd(rec.Cmd.Packet)

	if err != nil {
		return err
	}

	c.Text.StartResponse(id)
	defer c.Text.EndResponse(id)

	msg, err := collectClientLines(c.Text)

	if err != nil {
		return err
	}

	// mech := rec.Cmd.Fields[1]

	if proxy.envelope != nil {
		proxy.envelope.AuthType = rec.Cmd.Fields[1]
		proxy.envelope.AuthData = msg
		proxy.envelope.AuthRequest = rec.Cmd.Packet
	}

	fmt.Fprint(proxy.base.Conn, msg)

	return nil
}

func handleQUIT(rec *SMTPRecord, c *smtp.Client, proxy *SMTPSession) error {
	id, err := c.Text.Cmd(rec.Cmd.Packet)

	if err != nil {
		return err
	}

	c.Text.StartResponse(id)
	defer c.Text.EndResponse(id)

	msg, err := collectClientLines(c.Text)

	if err != nil {
		return err
	}

	fmt.Fprint(proxy.base.Conn, msg)

	c.Close()
	proxy.Close()

	return nil
}

func handleSTARTTLS(rec *SMTPRecord, c *smtp.Client, proxy *SMTPSession) error {

	if proxy.tlsd {
		proxy.Reply(502, "TLS Already Active")
	}

	if proxy.base.Config == nil {
		proxy.Reply(502, "TLS Not Supported")
	}

	id, err := c.Text.Cmd(rec.Cmd.Packet)

	if err != nil {
		return err
	}

	c.Text.StartResponse(id)
	defer c.Text.EndResponse(id)

	code, msg, err := c.Text.ReadResponse(220)

	if err != nil {
		proxy.Reply(502, "Not Supported")
		return err
	}

	tlsconn := tls.Server(proxy.base.Conn, proxy.base.Config)
	proxy.Reply(code, msg)

	if err := tlsconn.Handshake(); err != nil {
		proxy.Reply(550, "Handshake Error")
		return err
	}

	// proxy.Reply(220, "Go Ahead")
	proxy.Reset()

	proxy.base.Conn.SetDeadline(time.Time{})

	proxy.base.Conn = tlsconn
	proxy.reader = bufio.NewReader(tlsconn)
	proxy.writer = bufio.NewWriter(tlsconn)
	proxy.scanner = bufio.NewScanner(proxy.reader)
	proxy.tlsd = true

	state := tlsconn.ConnectionState()
	proxy.base.Peer.TLS = state

	proxy.flush()

	return nil
}

func handleXCLIENT(rec *SMTPRecord, c *smtp.Client, proxy *SMTPSession) error {
	id, err := c.Text.Cmd(rec.Cmd.Packet)

	if err != nil {
		return err
	}

	c.Text.StartResponse(id)
	defer c.Text.EndResponse(id)

	msg, err := collectClientLines(c.Text)

	if err != nil {
		return err
	}

	fmt.Fprint(proxy.base.Conn, msg)

	return nil
}

func handleNOOP(rec *SMTPRecord, c *smtp.Client, proxy *SMTPSession) error {
	id, err := c.Text.Cmd(rec.Cmd.Packet)

	if err != nil {
		return err
	}

	c.Text.StartResponse(id)
	defer c.Text.EndResponse(id)

	msg, err := collectClientLines(c.Text)

	if err != nil {
		return err
	}

	fmt.Fprint(proxy.base.Conn, msg)

	return nil
}

func handleHELO(rec *SMTPRecord, c *smtp.Client, proxy *SMTPSession) error {
	id, err := c.Text.Cmd(rec.Cmd.Packet)

	if err != nil {
		return err
	}

	c.Text.StartResponse(id)
	defer c.Text.EndResponse(id)

	msg, err := collectClientLines(c.Text)

	if err != nil {
		return err
	}

	fmt.Fprint(proxy.base.Conn, msg)

	return nil
}

func handleEHLO(rec *SMTPRecord, c *smtp.Client, proxy *SMTPSession) error {
	id, err := c.Text.Cmd(rec.Cmd.Packet)

	if err != nil {
		return err
	}

	c.Text.StartResponse(id)
	defer c.Text.EndResponse(id)

	msg, err := collectClientLines(c.Text)

	if err != nil {
		return err
	}

	log.Debug("EHLO Reply:", msg, err)
	fmt.Fprint(proxy.base.Conn, msg)
	return nil
}

func handleMAIL(rec *SMTPRecord, c *smtp.Client, proxy *SMTPSession) error {
	id, err := c.Text.Cmd(rec.Cmd.Packet)

	if err != nil {
		return err
	}

	if proxy.envelope != nil {
		proxy.envelope.Sender = rec.Cmd.Params[1]
	}

	c.Text.StartResponse(id)
	defer c.Text.EndResponse(id)

	msg, err := collectClientLines(c.Text)

	if err != nil {
		return err
	}

	fmt.Fprint(proxy.base.Conn, msg)
	return nil
}

func handleRCPT(rec *SMTPRecord, c *smtp.Client, proxy *SMTPSession) error {
	id, err := c.Text.Cmd(rec.Cmd.Packet)

	if err != nil {
		return err
	}

	c.Text.StartResponse(id)
	defer c.Text.EndResponse(id)

	if proxy.envelope != nil {
		proxy.envelope.Receivers = append(proxy.envelope.Receivers, rec.Cmd.Params[1])
	}

	msg, err := collectClientLines(c.Text)

	if err != nil {
		return err
	}

	fmt.Fprint(proxy.base.Conn, msg)
	return nil
}

func handleRSET(rec *SMTPRecord, c *smtp.Client, proxy *SMTPSession) error {
	id, err := c.Text.Cmd(rec.Cmd.Packet)

	if err != nil {
		return err
	}

	c.Text.StartResponse(id)
	defer c.Text.EndResponse(id)

	msg, err := collectClientLines(c.Text)

	if err != nil {
		return err
	}

	proxy.Reset()
	fmt.Fprint(proxy.base.Conn, msg)
	return nil
}

func handleGeneric(rec *SMTPRecord, c *smtp.Client, proxy *SMTPSession) error {
	id, err := c.Text.Cmd(rec.Cmd.Packet)

	if err != nil {
		return err
	}

	c.Text.StartResponse(id)
	defer c.Text.EndResponse(id)

	msg, err := collectClientLines(c.Text)

	if err != nil {
		return err
	}

	fmt.Fprint(proxy.base.Conn, msg)
	return nil
}
