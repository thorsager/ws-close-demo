package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/coder/websocket"
)

// Custom WebSocket close codes for application-level errors.
// The 4000-4999 range is reserved for application use (per RFC 6455).
const (
	CloseCodeInsufficientCredits = 4001
	CloseCodeRateLimited         = 4002
	CloseCodeAccountSuspended    = 4003
)

// wsHandler upgrades the HTTP request to a WebSocket connection, then
// performs business-logic checks. If a check fails, the connection is
// closed with a custom code (4000-4999) and a JSON reason string that
// the browser WebSocket API exposes in the onclose event.
func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Printf("upgrade error: %v", err)
		return
	}
	defer conn.Close(websocket.StatusInternalError, "server closing")

	log.Printf("client connected: %s", r.RemoteAddr)

	// Business-logic check: simulate a failure via ?status=NNN
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		if code, err := strconv.Atoi(statusStr); err == nil {
			msg := r.URL.Query().Get("message")
			if msg == "" {
				msg = http.StatusText(code)
			}

			wsCloseCode := websocket.StatusCode(http.StatusBadGateway)
			switch code {
			case 402:
				wsCloseCode = CloseCodeInsufficientCredits
			case 429:
				wsCloseCode = CloseCodeRateLimited
			case 403:
				wsCloseCode = CloseCodeAccountSuspended
			}

			reason, _ := json.Marshal(map[string]any{
				"httpStatus": code,
				"message":    msg,
			})

			log.Printf("closing %s: code=%d reason=%s", r.RemoteAddr, wsCloseCode, reason)
			conn.Close(wsCloseCode, string(reason))
			return
		}
	}

	// Normal echo loop
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	for {
		mt, msg, err := conn.Read(ctx)
		if err != nil {
			log.Printf("read error: %v", err)
			break
		}
		log.Printf("recv from %s: %s", r.RemoteAddr, msg)
		if err := conn.Write(ctx, mt, msg); err != nil {
			log.Printf("write error: %v", err)
			break
		}
	}
	log.Printf("client disconnected: %s", r.RemoteAddr)
}

func main() {
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	addr := ":8080"
	log.Printf("listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
