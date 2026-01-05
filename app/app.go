package main

import (
	"app/assets"
	"app/audio"
	"context"
	"log"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yalue/onnxruntime_go"
)

const (
	SAMPLING_RATE = 16000

	WINDOW_SIZE = 2 * SAMPLING_RATE
	HOP_SIZE    = SAMPLING_RATE / 2

	VAD_WINDOW_SIZE  = 512
	VAD_CONTEXT_SIZE = 64
	VAD_INPUT_SIZE   = VAD_WINDOW_SIZE + VAD_CONTEXT_SIZE // 576
	VAD_STATE_SIZE   = 2 * 1 * 128
	VAD_THRESHOLD    = 0.7 // set high threshold for fast speakers or edited content (inclusive)

	MAX_UTTERANCE_LEN = WINDOW_SIZE * 3 // 6 seconds
)

type App struct {
	ctx context.Context
	lo  audio.Loopback

	serSession *onnxruntime_go.AdvancedSession
	serInput   *onnxruntime_go.Tensor[float32]
	serMask    *onnxruntime_go.Tensor[int64]
	serOutput  *onnxruntime_go.Tensor[float32]

	vadSession *onnxruntime_go.AdvancedSession
	vadInput   *onnxruntime_go.Tensor[float32]
	vadState   *onnxruntime_go.Tensor[float32]
	vadStateN  *onnxruntime_go.Tensor[float32]
	vadSR      *onnxruntime_go.Tensor[int64]
	vadOutput  *onnxruntime_go.Tensor[float32]

	vadContext []float32
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	lo := audio.NewLoopback()
	if err := lo.Start(); err != nil {
		panic(err)
	}
	a.lo = lo

	libPath, serModelPath, vadModelPath, err := assets.ExtractEmbeddedFiles()
	if err != nil {
		panic(err)
	}

	onnxruntime_go.SetSharedLibraryPath(libPath)
	if err := onnxruntime_go.InitializeEnvironment(); err != nil {
		panic(err)
	}

	a.serInput, _ = onnxruntime_go.NewTensor(
		onnxruntime_go.NewShape(1, WINDOW_SIZE),
		make([]float32, WINDOW_SIZE),
	)

	a.serMask, _ = onnxruntime_go.NewTensor(
		onnxruntime_go.NewShape(1, WINDOW_SIZE),
		make([]int64, WINDOW_SIZE),
	)
	for i := range a.serMask.GetData() {
		a.serMask.GetData()[i] = 1
	}

	a.serOutput, _ = onnxruntime_go.NewEmptyTensor[float32](
		onnxruntime_go.NewShape(1, 9),
	)

	a.serSession, err = onnxruntime_go.NewAdvancedSession(
		serModelPath,
		[]string{"input_values", "attention_mask"},
		[]string{"preds"},
		[]onnxruntime_go.Value{a.serInput, a.serMask},
		[]onnxruntime_go.Value{a.serOutput},
		nil,
	)
	if err != nil {
		panic(err)
	}

	a.vadContext = make([]float32, VAD_CONTEXT_SIZE)

	a.vadInput, _ = onnxruntime_go.NewTensor(
		onnxruntime_go.NewShape(1, VAD_INPUT_SIZE),
		make([]float32, VAD_INPUT_SIZE),
	)

	a.vadState, _ = onnxruntime_go.NewTensor(
		onnxruntime_go.NewShape(2, 1, 128),
		make([]float32, VAD_STATE_SIZE),
	)

	a.vadStateN, _ = onnxruntime_go.NewEmptyTensor[float32](
		onnxruntime_go.NewShape(2, 1, 128),
	)

	a.vadSR, _ = onnxruntime_go.NewTensor(
		onnxruntime_go.NewShape(1),
		[]int64{SAMPLING_RATE},
	)

	a.vadOutput, _ = onnxruntime_go.NewEmptyTensor[float32](
		onnxruntime_go.NewShape(1, 1),
	)

	a.vadSession, err = onnxruntime_go.NewAdvancedSession(
		vadModelPath,
		[]string{"input", "state", "sr"},
		[]string{"output", "stateN"},
		[]onnxruntime_go.Value{
			a.vadInput,
			a.vadState,
			a.vadSR,
		},
		[]onnxruntime_go.Value{
			a.vadOutput,
			a.vadStateN,
		},
		nil,
	)
	if err != nil {
		panic(err)
	}

	go a.inferenceLoop(ctx)
}

func (a *App) inferenceLoop(ctx context.Context) {
	vadBuf := make([]float32, 0, VAD_WINDOW_SIZE*4)
	utterance := make([]float32, 0, MAX_UTTERANCE_LEN)
	inUtterance := false
	utteranceID := 0

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
				continue
			}

			vadBuf = append(vadBuf, samples...)

			for len(vadBuf) >= VAD_WINDOW_SIZE {
				chunk := vadBuf[:VAD_WINDOW_SIZE]
				vadBuf = vadBuf[VAD_WINDOW_SIZE:]

				// Prepare VAD input
				copy(a.vadInput.GetData()[:VAD_CONTEXT_SIZE], a.vadContext)
				copy(a.vadInput.GetData()[VAD_CONTEXT_SIZE:], chunk)

				if err := a.vadSession.Run(); err != nil {
					log.Printf("VAD inference error: %v", err)
					continue
				}

				prob := a.vadOutput.GetData()[0]
				runtime.EventsEmit(a.ctx, "vad", prob)

				copy(a.vadState.GetData(), a.vadStateN.GetData())
				copy(a.vadContext, a.vadInput.GetData()[VAD_INPUT_SIZE-VAD_CONTEXT_SIZE:])

				if prob >= VAD_THRESHOLD {
					if !inUtterance {
						inUtterance = true
						utteranceID++
						utterance = utterance[:0]
						runtime.EventsEmit(a.ctx, "utterance_start", utteranceID)
					}

					utterance = append(utterance, chunk...)

					// Emit and reset if utterance too long
					for len(utterance) >= MAX_UTTERANCE_LEN {
						part := utterance[:MAX_UTTERANCE_LEN]

						// Copy for SER
						copy(a.serInput.GetData(), part)
						go func() {
							for i := range a.serMask.GetData() {
								if i < len(utterance) {
									a.serMask.GetData()[i] = 1
								} else {
									a.serMask.GetData()[i] = 0
								}
							}

							if err := a.serSession.Run(); err != nil {
								log.Printf("SER inference error: %v", err)
							}

							runtime.EventsEmit(ctx, "inference", a.serOutput.GetData())
						}()

						utterance = utterance[MAX_UTTERANCE_LEN:]
					}

				} else if inUtterance {
					// VAD dropped -> finalize utterance
					inUtterance = false
					runtime.EventsEmit(a.ctx, "utterance_end", utteranceID)

					if len(utterance) > 0 {
						for i := range a.serMask.GetData() {
							if i < len(utterance) {
								a.serMask.GetData()[i] = 1
							} else {
								a.serMask.GetData()[i] = 0
							}
						}

						copy(a.serInput.GetData(), utterance)
						go func() {
							if err := a.serSession.Run(); err != nil {
								log.Printf("SER inference error: %v", err)
							}

							runtime.EventsEmit(ctx, "inference", a.serOutput.GetData())
						}()
						utterance = utterance[:0]
					}
				}
			}
		}
	}
}

func (a *App) shutdown(ctx context.Context) {
	if a.lo != nil {
		a.lo.Stop()
	}

	if a.serSession != nil {
		a.serSession.Destroy()
	}

	if a.vadSession != nil {
		a.vadSession.Destroy()
	}

	onnxruntime_go.DestroyEnvironment()
}
