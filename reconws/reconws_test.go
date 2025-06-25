package reconws_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"github.com/thienhaole92/uframework/reconws"
)

func TestWebsocketIntegration(t *testing.T) {
	t.Parallel()

	// Create a cancelable context to control server lifecycle
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure cleanup

	// Start the WebSocket server
	ready := make(chan struct{})
	go StartWebsocketTestServer(ctx, ready)

	// Wait for the server to be ready
	<-ready
	time.Sleep(500 * time.Millisecond) // Ensure the server is fully ready

	// Create the WebSocket client
	wsc := &reconws.Websocket{
		ReconnectIntervalMin: 2 * time.Second,
		ReconnectIntervalMax: 10 * time.Second,
		Verbose:              true,
		Logger:               &log.Logger,
		OnConnect: func(_ *reconws.Websocket) {
			log.Info().Msg("connected to websocket server successfully")
		},
		ID:                      0,
		Name:                    "",
		Meta:                    nil,
		Errors:                  nil,
		Reconnect:               false,
		ReconnectIntervalFactor: 0,
		HandshakeTimeout:        0,
		OnDisconnect:            nil,
		OnConnectError:          nil,
		OnDisconnectError:       nil,
		OnReadError:             nil,
		OnWriteError:            nil,
		OnPreReconnect:          nil,
		Conn:                    nil,
	}

	// Connect to the WebSocket server
	err := wsc.Dial("ws://0.0.0.0:8888/ws", nil)
	require.NoError(t, err)

	// Ensure connection is active before sending messages
	require.True(t, wsc.IsConnected())

	// Send a message
	err = wsc.WriteJSON(map[string]any{"name": "trading"})
	require.NoError(t, err)

	// Receive a message
	messageType, message, err := wsc.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, websocket.TextMessage, messageType)
	require.Equal(t, []byte("Hello, Client!"), message)

	// Close the WebSocket connection
	wsc.Close()

	// Stop the server after the test is done
	cancel()
}

func StartWebsocketTestServer(ctx context.Context, ready chan struct{}) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(_ *http.Request) bool {
			return true // Allow all origins
		},
		HandshakeTimeout:  0,
		ReadBufferSize:    0,
		WriteBufferSize:   0,
		WriteBufferPool:   nil,
		Subprotocols:      nil,
		Error:             nil,
		EnableCompression: false,
	}

	server := &http.Server{
		Addr:                         "0.0.0.0:8888",
		ReadTimeout:                  10 * time.Second,
		WriteTimeout:                 10 * time.Second,
		Handler:                      nil,
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    nil,
		ReadHeaderTimeout:            0,
		IdleTimeout:                  0,
		MaxHeaderBytes:               0,
		TLSNextProto:                 nil,
		ConnState:                    nil,
		ErrorLog:                     nil,
		BaseContext:                  nil,
		ConnContext:                  nil,
	}

	http.HandleFunc("/ws", handleConnection(ctx, upgrader))

	// Signal that the server is ready
	close(ready)

	// Start the server and listen for shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Logger.Err(err).Msg("websocket server error")
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	log.Logger.Info().Msg("server shutting down")

	// Gracefully shutdown the server
	_ = server.Shutdown(ctx)
}

func handleConnection(ctx context.Context, upgrader websocket.Upgrader) func(http.ResponseWriter, *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Logger.Err(err).Msg("failed to upgrade connection")

			return
		}

		log.Logger.Info().Msg("websocket client connected")

		defer conn.Close()

		// Handle incoming messages
		for {
			select {
			case <-ctx.Done():
				log.Logger.Info().Msg("stopping websocket server")

				// Gracefully close the WebSocket connection
				err := conn.WriteMessage(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(
						websocket.CloseNormalClosure,
						"Server shutting down",
					),
				)
				if err != nil {
					log.Logger.Err(err).Msg("failed to send close message")
				}

				time.Sleep(500 * time.Millisecond) // Give time for client to process close

				return
			default:
				// Read a message from the client
				_, message, err := conn.ReadMessage()
				if err != nil {
					log.Logger.Err(err).Msg("client disconnected")

					return
				}

				log.Logger.Info().Any("payload", string(message)).Msg("received message from client")

				// Respond to the client
				err = conn.WriteMessage(websocket.TextMessage, []byte("Hello, Client!"))
				if err != nil {
					log.Logger.Err(err).Msg("failed to send response")

					return
				}
			}
		}
	}

	return handler
}
