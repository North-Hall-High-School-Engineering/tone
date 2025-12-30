package main

import (
	"app/audio"
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
	lo  audio.Loopback
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	lo := audio.NewLoopback()
	err := lo.Start()
	if err != nil {
		panic("failed to start loopback: " + err.Error())
	}

	a.lo = lo
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				samples, err := a.lo.Read()
				if err != nil {
					println("read error:", err.Error())
					continue
				}
				if len(samples) > 0 {
					// prepare debug info
					debug := map[string]any{
						"numSamples": len(samples),
						"maxValue":   maxFloat32(samples),
						"minValue":   minFloat32(samples),
					}

					// send to frontend
					runtime.EventsEmit(a.ctx, "audioDebug", debug)
				}
			}
		}
	}()
}

func maxFloat32(s []float32) float32 {
	if len(s) == 0 {
		return 0
	}
	max := s[0]
	for _, v := range s {
		if v > max {
			max = v
		}
	}
	return max
}

func minFloat32(s []float32) float32 {
	if len(s) == 0 {
		return 0
	}
	min := s[0]
	for _, v := range s {
		if v < min {
			min = v
		}
	}
	return min
}

func (a *App) shutdown(ctx context.Context) {
	if a.lo != nil {
		a.lo.Stop()
		a.lo = nil
	}
}
