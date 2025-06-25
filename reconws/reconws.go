package reconws

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jpillora/backoff"
	"github.com/rs/zerolog"
)

const (
	DefaultMinReconnectInterval = 2 * time.Second
	DefaultMaxReconnectInterval = 30 * time.Second
	DefaultHandshakeTimeout     = 2 * time.Second
)

var (
	ErrNotConnected          = errors.New("websocket not connected")
	ErrURLEmpty              = errors.New("url can not be empty")
	ErrURLWrongScheme        = errors.New("websocket uri must start with ws or wss scheme")
	ErrURLNamePassNotAllowed = errors.New("user name and password are not allowed in websocket uri")
)

type WsOpts func(dl *websocket.Dialer)

type Logger interface {
	Error() *zerolog.Event
	Info() *zerolog.Event
}

type Websocket struct {
	ID   uint64         // Websocket ID
	Name string         // Websocket Name
	Meta map[string]any // Websocket Meta

	Logger Logger
	Errors chan<- error

	Reconnect bool

	// Reconnect intervals and factors
	ReconnectIntervalMin    time.Duration
	ReconnectIntervalMax    time.Duration
	ReconnectIntervalFactor float64
	HandshakeTimeout        time.Duration
	Verbose                 bool

	// Callback functions
	OnConnect         func(ws *Websocket)
	OnDisconnect      func(ws *Websocket)
	OnConnectError    func(ws *Websocket, err error)
	OnDisconnectError func(ws *Websocket, err error)
	OnReadError       func(ws *Websocket, err error)
	OnWriteError      func(ws *Websocket, err error)
	OnPreReconnect    func() error

	dialer        *websocket.Dialer
	url           string
	requestHeader http.Header
	httpResponse  *http.Response
	mu            sync.Mutex
	wmu           sync.Mutex // Write mutex
	dialErr       error
	isConnected   bool
	*websocket.Conn
}

func (ws *Websocket) WriteJSON(val any) error {
	err := ErrNotConnected

	if ws.IsConnected() {
		ws.wmu.Lock()

		err = ws.Conn.WriteJSON(val)
		if err != nil {
			if ws.OnWriteError != nil {
				ws.OnWriteError(ws, err)
			}

			ws.closeAndReconnect()
		}

		ws.wmu.Unlock()
	}

	return err
}

func (ws *Websocket) WriteMessage(messageType int, data []byte) error {
	err := ErrNotConnected

	if ws.IsConnected() {
		ws.wmu.Lock()

		err = ws.Conn.WriteMessage(messageType, data)
		if err != nil {
			if ws.OnWriteError != nil {
				ws.OnWriteError(ws, err)
			}

			ws.closeAndReconnect()
		}

		ws.wmu.Unlock()
	}

	return err
}

func (ws *Websocket) ReadMessage() (int, []byte, error) {
	if !ws.IsConnected() {
		return 0, nil, ErrNotConnected
	}

	messageType, message, err := ws.Conn.ReadMessage()
	if err != nil {
		if ws.OnReadError != nil {
			ws.OnReadError(ws, err)
		}

		ws.closeAndReconnect()
	}

	return messageType, message, err
}

func (ws *Websocket) Close() {
	ws.mu.Lock()
	if ws.Conn != nil {
		err := ws.Conn.Close()
		if err == nil && ws.isConnected && ws.OnDisconnect != nil {
			ws.OnDisconnect(ws)
		}

		if err != nil && ws.OnDisconnectError != nil {
			ws.OnDisconnectError(ws, err)
		}
	}

	ws.isConnected = false
	ws.mu.Unlock()
}

func (ws *Websocket) closeAndReconnect() {
	ws.Close()

	if ws.OnPreReconnect != nil {
		if err := ws.OnPreReconnect(); err != nil {
			ws.logError(fmt.Sprintf("can not reconnect to %s, error: %s", ws.url, err.Error()))
		}
	}

	ws.Connect()
}

func (ws *Websocket) Dial(urlStr string, reqHeader http.Header, opts ...WsOpts) error {
	if _, err := parseURL(urlStr); err != nil {
		return err
	}

	ws.url = urlStr
	ws.requestHeader = reqHeader
	ws.setDefaults()

	ws.dialer = &websocket.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		HandshakeTimeout:  ws.HandshakeTimeout,
		NetDial:           nil,
		NetDialContext:    nil,
		NetDialTLSContext: nil,
		TLSClientConfig:   nil,
		ReadBufferSize:    0,
		WriteBufferSize:   0,
		WriteBufferPool:   nil,
		Subprotocols:      nil,
		EnableCompression: false,
		Jar:               nil,
	}

	for _, opt := range opts {
		opt(ws.dialer)
	}

	go ws.Connect()

	time.Sleep(ws.HandshakeTimeout) // Initial handshake timeout wait

	return nil
}

func (ws *Websocket) Connect() {
	bkf := &backoff.Backoff{
		Min:    ws.ReconnectIntervalMin,
		Max:    ws.ReconnectIntervalMax,
		Factor: ws.ReconnectIntervalFactor,
		Jitter: true,
	}

	for {
		nextInterval := bkf.Duration()

		wsConn, httpResp, err := ws.dialer.Dial(ws.url, ws.requestHeader)
		defer func() {
			if httpResp != nil {
				httpResp.Body.Close() // Close the response body when done
			}
		}()

		ws.mu.Lock()
		ws.Conn = wsConn
		ws.dialErr = err
		ws.isConnected = err == nil
		ws.httpResponse = httpResp
		ws.mu.Unlock()

		if err == nil {
			ws.logSuccess(fmt.Sprint("successfully connected to ", ws.url))

			if ws.OnConnect != nil {
				ws.OnConnect(ws)
			}

			return
		}

		ws.logError(fmt.Sprintf("can not connect to %s, retrying in %v", ws.url, nextInterval))

		if ws.OnConnectError != nil {
			ws.OnConnectError(ws, err)
		}

		time.Sleep(nextInterval)
	}
}

func (ws *Websocket) GetHTTPResponse() *http.Response {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	return ws.httpResponse
}

func (ws *Websocket) GetDialError() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	return ws.dialErr
}

func (ws *Websocket) IsConnected() bool {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	return ws.isConnected
}

func (ws *Websocket) setDefaults() {
	if ws.ReconnectIntervalMin == 0 {
		ws.ReconnectIntervalMin = DefaultMinReconnectInterval
	}

	if ws.ReconnectIntervalMax == 0 {
		ws.ReconnectIntervalMax = DefaultMaxReconnectInterval
	}

	if ws.ReconnectIntervalFactor == 0 {
		ws.ReconnectIntervalFactor = 1.5
	}

	if ws.HandshakeTimeout == 0 {
		ws.HandshakeTimeout = DefaultHandshakeTimeout
	}
}

func parseURL(urlStr string) (*url.URL, error) {
	if urlStr == "" {
		return nil, ErrURLEmpty
	}

	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	if url.Scheme != "ws" && url.Scheme != "wss" {
		return nil, ErrURLWrongScheme
	}

	if url.User != nil {
		return nil, ErrURLNamePassNotAllowed
	}

	return url, nil
}

func (ws *Websocket) logError(message string) {
	if ws.Verbose && ws.Logger != nil {
		ws.Logger.Error().Msg(message)
	}
}

func (ws *Websocket) logSuccess(message string) {
	if ws.Verbose && ws.Logger != nil {
		ws.Logger.Info().Msg(message)
	}
}
