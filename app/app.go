package main

import (
	"app/audio"
	"context"
	"encoding/json"
	"log"
	"math"
	"time"

	"github.com/gorilla/websocket"
)

type App struct {
	ctx context.Context
	lo  audio.Loopback
	ws  *websocket.Conn
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.lo = audio.NewLoopback()

	if err := a.lo.Start(); err != nil {
		panic(err)
	}

	var err error
	a.ws, _, err = websocket.DefaultDialer.Dial("ws://localhost:8000/v1/ws", nil)
	if err != nil {
		panic(err)
	}

	go a.send(ctx)
	go a.recv(ctx)
}

func (a *App) send(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			samples, err := a.lo.Read()
			if err != nil {
				log.Printf("audio read error: %v", err)
				continue
			}
			if len(samples) == 0 {
				time.Sleep(1 * time.Millisecond)
				continue
			}

			data := float32SliceToBytes(samples)
			err = a.ws.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				log.Printf("websocket send error: %v", err)
				return
			}

		}
	}
}

func (a *App) recv(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, msg, err := a.ws.ReadMessage()
			if err != nil {
				log.Printf("websocket read error: %v", err)
				return
			}

			var payload map[string]interface{}
			if err := json.Unmarshal(msg, &payload); err != nil {
				log.Printf("invalid JSON received: %v", err)
				continue
			}

			log.Printf("received JSON: %+v", payload)
		}
	}
}

func (a *App) shutdown(ctx context.Context) {
	a.lo.Stop()
}

func float32SliceToBytes(f []float32) []byte {
	buf := make([]byte, len(f)*4)
	for i, v := range f {
		bits := math.Float32bits(v)
		buf[i*4] = byte(bits)
		buf[i*4+1] = byte(bits >> 8)
		buf[i*4+2] = byte(bits >> 16)
		buf[i*4+3] = byte(bits >> 24)
	}
	return buf
}
