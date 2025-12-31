package main

import (
	"app/assets"
	"app/audio"
	"app/ort"
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yalue/onnxruntime_go"
)

type App struct {
	ctx context.Context
	lo  audio.Loopback

	session *onnxruntime_go.AdvancedSession

	inputTensor  *onnxruntime_go.Tensor[float32]
	maskTensor   *onnxruntime_go.Tensor[int64]
	outputTensor *onnxruntime_go.Tensor[float32]
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Audio
	lo := audio.NewLoopback()
	if err := lo.Start(); err != nil {
		panic("failed to start loopback: " + err.Error())
	}
	a.lo = lo

	// ORT runtime
	runtimePath, err := ort.ExtractEmbeddedOrt()
	if err != nil {
		panic(err)
	}
	onnxruntime_go.SetSharedLibraryPath(runtimePath)

	if err := onnxruntime_go.InitializeEnvironment(); err != nil {
		panic(err)
	}

	// Extract model
	data, err := assets.FS.ReadFile("models/tone.onnx")
	if err != nil {
		panic(err)
	}

	tmpFile := filepath.Join(os.TempDir(), "tone.onnx")
	if err := os.WriteFile(tmpFile, data, 0755); err != nil {
		panic(err)
	}

	// ----------------------------
	// Pre-create tensors
	// ----------------------------
	inputShape := onnxruntime_go.NewShape(1, WINDOW_SIZE)
	maskShape := onnxruntime_go.NewShape(1, WINDOW_SIZE)
	outputShape := onnxruntime_go.NewShape(1, 9)

	a.inputTensor, err = onnxruntime_go.NewTensor(inputShape, make([]float32, WINDOW_SIZE))
	if err != nil {
		panic(err)
	}

	a.maskTensor, err = onnxruntime_go.NewTensor(maskShape, make([]int64, WINDOW_SIZE))
	if err != nil {
		panic(err)
	}

	// attention mask = all ones
	for i := range a.maskTensor.GetData() {
		a.maskTensor.GetData()[i] = 1
	}

	a.outputTensor, err = onnxruntime_go.NewEmptyTensor[float32](outputShape)
	if err != nil {
		panic(err)
	}

	// ----------------------------
	// Create session WITH tensors
	// ----------------------------
	a.session, err = onnxruntime_go.NewAdvancedSession(
		tmpFile,
		[]string{"input_values", "attention_mask"},
		[]string{"preds"},
		[]onnxruntime_go.Value{a.inputTensor, a.maskTensor},
		[]onnxruntime_go.Value{a.outputTensor},
		nil,
	)
	if err != nil {
		panic(err)
	}

	go a.inferenceLoop(ctx)
}

func (a *App) inferenceLoop(ctx context.Context) {
	// Start EMPTY buffer (important)
	buf := make([]float32, 0, WINDOW_SIZE*2)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			samples, err := a.lo.Read()
			if err != nil {
				log.Printf("read error: %v", err)
				continue
			}

			buf = append(buf, samples...)

			for len(buf) >= WINDOW_SIZE {
				window := buf[:WINDOW_SIZE]

				// Copy window into pre-allocated tensor
				copy(a.inputTensor.GetData(), window)

				// Run inference
				if err := a.session.Run(); err != nil {
					log.Printf("inference error: %v", err)
					break
				}

				runtime.EventsEmit(ctx, "inference", a.outputTensor.GetData())

				// Slide window
				buf = buf[HOP_SIZE:]
			}
		}
	}
}

func (a *App) shutdown(ctx context.Context) {
	if a.lo != nil {
		a.lo.Stop()
		a.lo = nil
	}

	if a.session != nil {
		a.session.Destroy()
		a.session = nil
	}

	if a.inputTensor != nil {
		a.inputTensor.Destroy()
		a.maskTensor.Destroy()
		a.outputTensor.Destroy()
	}

	onnxruntime_go.DestroyEnvironment()
}
