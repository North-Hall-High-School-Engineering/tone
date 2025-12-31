package main

import (
	"app/audio"
	"context"
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
}

func (a *App) shutdown(ctx context.Context) {
	if a.lo != nil {
		a.lo.Stop()
		a.lo = nil
	}
}
