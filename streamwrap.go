package corekit

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/tochka/core-kit/apierror"
)

type StreamAPIHandler func(req *http.Request) (receiver chan []byte, cancel chan struct{}, err error)

var defaultUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { // skip check origin header
		return true
	},
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024
)

func streamWrapAPIHandler(log func(format string, args ...interface{})) func(handler StreamAPIHandler) http.Handler {
	return func(handler StreamAPIHandler) http.Handler {
		wrap := func(w http.ResponseWriter, r *http.Request) {
			var ok bool
			w.Header().Set("Content-Type", "application/json")

			receiver, cancel, err := handler(r)
			if err != nil {
				var apiErr apierror.APIError

				innerErr := errors.Cause(err)
				if apiErr, ok = innerErr.(apierror.APIError); !ok {
					log("[ERROR] API wrapper: %+v", err)
					apiErr = apierror.InternalServerErr
				}
				w.WriteHeader(apiErr.StatusCode)

				if apiErr.StatusCode != http.StatusBadRequest {
					return
				}

				b, _ := json.Marshal(apiErr)
				w.Write(b)
				return
			}

			wsConn, err := defaultUpgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}

			wsConn.SetReadLimit(maxMessageSize)
			wsConn.SetReadDeadline(time.Now().Add(pongWait))
			wsConn.SetPongHandler(func(string) error { wsConn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

			ticker := time.NewTicker(pingPeriod)
			chWSClosed := make(chan struct{}, 2)
			for {
				select {
				case data := <-receiver:
					wsConn.SetWriteDeadline(time.Now().Add(writeWait))
					err = wsConn.WriteMessage(websocket.BinaryMessage, data)
					if err != nil {
						chWSClosed <- struct{}{}
					}
				case <-ticker.C:
					if err := wsConn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
						chWSClosed <- struct{}{}
					}
				case <-chWSClosed:
					cancel <- struct{}{}
					wsConn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(writeWait))
					wsConn.Close()
					close(chWSClosed)

					go func(r chan []byte) { // read all rest messages to /dev/null
						for _ = range r {
						}
					}(receiver)
					return
				}
			}
		}

		return http.HandlerFunc(wrap)
	}
}
