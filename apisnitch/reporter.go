package apisnitch

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"
)

type reporter struct {
	serviceURL   string
	apiSecret    string
	apiCallCache ApiCallCache
	wsClient     *websocket.Conn
	messageOut   chan []byte
}

type Reporter interface {
	Start() error
	CreateApiCall(requestID string, apiPath string, apiMethod string, request Request)
	FinalizeApiCall(reqID string, response Response)
}

func NewReporter(serviceURL string, apiSecret string, apiCallCache ApiCallCache) Reporter {
	return &reporter{
		serviceURL:   serviceURL,
		apiSecret:    apiSecret,
		apiCallCache: apiCallCache,
		wsClient:     nil,
	}
}

func (r *reporter) Start() error {
	slog.Info("API Whisperer Reporter connecting")

	r.messageOut = make(chan []byte, 100)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: r.serviceURL, Path: "/ws"}
	slog.Info("Connecting", "url", u)

	c, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		slog.Error("handshake failed", "error", err)
		return err
	}
	r.wsClient = c
	if resp != nil {
		slog.Debug("handshake response", "status", resp.Status)
	}
	slog.Debug("Connected to API Whisperer server")

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer r.wsClient.Close()
		// Listen for messages coming from the server
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				slog.Error("read:", "error", err)
				return
			}
			messageStr := string(message)
			slog.Debug("Received message on client side", "message", messageStr)

			if strings.HasPrefix(messageStr, "OK") {
				// This is a confirmation message
				parts := strings.Split(messageStr, ":")
				err = r.apiCallCache.DeleteApiCall(parts[1])
				if err != nil {
					slog.Error("Failed to delete api call from cache", "error", err)
				}
			} else if strings.HasPrefix(messageStr, "ERROR") {
				// This is an error message
				slog.Error("Received error message from server", "message", messageStr)
			} else {
				// Unknown message type
				slog.Info("Received unknown message from server", "message", messageStr)
			}
		}
	}()

	slog.Debug("Waiting for messages")
	for {
		select {
		case <-done:
			slog.Info("Connection closed")
			return nil
		case m := <-r.messageOut:
			go r.sendMessage(m)
		case <-interrupt:
			slog.Info("Interrupt received, closing connection")
			err := r.wsClient.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				slog.Error("write close:", "error", err)
				return err
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return nil
		}
	}
}

func (r *reporter) CreateApiCall(requestID string, apiPath string, apiMethod string, request Request) {
	r.apiCallCache.CreateApiCall(requestID, apiPath, apiMethod, request)
}

func (r *reporter) FinalizeApiCall(reqID string, response Response) {
	apiCall := r.apiCallCache.FinalizeApiCall(reqID, response)

	j, err := json.Marshal(apiCall)
	if err != nil {
		slog.Error("Failed to marshal api call", "error", err)
		return
	}
	r.messageOut <- j
	slog.Info("Sent API call to server")
}

func (r *reporter) sendMessage(message []byte) {
	slog.Debug("Sending message", "message", message)
	err := r.wsClient.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		slog.Error("Failed to send message", "error", err)
	}
}
