package main

import (
	"app/audio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx    context.Context
	lo     audio.Loopback
	ws     *websocket.Conn
	labels []string
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

	if err := a.fetchLabels(); err != nil {
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

			var sum float64
			for _, s := range samples {
				sum += float64(s * s)
			}
			rms := math.Sqrt(sum / float64(len(samples)))

			runtime.EventsEmit(a.ctx, "loudness", rms)

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

			switch payload["type"] {
			case "inference":
				rawScores, ok := payload["scores"].([]interface{})
				if !ok {
					log.Println("invalid scores format")
					continue
				}

				results := make([]map[string]interface{}, len(rawScores))

				for i, s := range rawScores {
					scoreFloat, _ := s.(float64)

					label := ""
					if i < len(a.labels) {
						label = a.labels[i]
					}

					results[i] = map[string]interface{}{
						"label": label,
						"score": scoreFloat,
					}
				}

				runtime.EventsEmit(a.ctx, "inference", results)
			}
			log.Printf("received JSON: %+v", payload)
		}
	}
}

type LabelsResponse struct {
	Model   string   `json:"model"`
	Version string   `json:"version"`
	Labels  []string `json:"labels"`
}

func (a *App) fetchLabels() error {
	resp, err := http.Get("http://localhost:8000/v1/labels")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("labels request failed: %s", string(body))
	}

	var lr LabelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return err
	}

	a.labels = lr.Labels

	log.Printf("Loaded labels: %+v", a.labels)
	return nil
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
