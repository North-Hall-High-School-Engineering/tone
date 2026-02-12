//go:build windows
// +build windows

package audio

/*
#cgo LDFLAGS: -lole32 -luuid -loleaut32
#include "lo_windows.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"math"
	"unsafe"

	"github.com/gopxl/beep/v2"
)

// windowsLoopback implements Loopback for Windows WASAPI
type windowsLoopback struct {
	lb *C.Loopback
}

func newLoopback() *windowsLoopback {
	return &windowsLoopback{}
}

func (w *windowsLoopback) Start() error {
	if w.lb != nil {
		return errors.New("loopback already started")
	}

	w.lb = C.loopback_start()
	if w.lb == nil {
		return errors.New("failed to start loopback")
	}

	return nil
}

func (w *windowsLoopback) Stop() error {
	if w.lb == nil {
		return errors.New("loopback not started")
	}
	C.loopback_stop(w.lb)
	w.lb = nil
	return nil
}

func (w *windowsLoopback) Read() ([]float32, error) {
	if w.lb == nil {
		return nil, errors.New("loopback not started")
	}

	pwfx := C.loopback_get_waveformat(w.lb)
	channels := int(pwfx.nChannels)
	bits := int(pwfx.wBitsPerSample)
	inRate := int(pwfx.nSamplesPerSec)

	if bits != 32 {
		return nil, fmt.Errorf("unsupported bits per sample: %d", bits)
	}

	var data *C.BYTE
	var numFrames C.UINT32

	ret := C.loopback_read(w.lb, &data, &numFrames)
	if ret <= 0 || data == nil {
		return nil, nil
	}
	defer C.free(unsafe.Pointer(data))

	totalSamples := int(numFrames) * channels
	src := unsafe.Slice((*float32)(unsafe.Pointer(data)), totalSamples)

	audio := make([]float32, int(numFrames))
	maxVal := float32(0)

	for i := 0; i < int(numFrames); i++ {
		sum := float32(0)
		for c := 0; c < channels; c++ {
			v := src[i*channels+c]
			sum += v
			if abs := float32(math.Abs(float64(v))); abs > maxVal {
				maxVal = abs
			}
		}
		audio[i] = sum / float32(channels)
	}

	streamer := &floatStreamer{samples: audio, pos: 0}
	resampled := beep.Resample(4, beep.SampleRate(inRate), beep.SampleRate(16000), streamer)

	var out []float32
	for {
		buf := make([][2]float64, 1024) // <-- must be [2]float64, not [1]
		n, ok := resampled.Stream(buf)
		for i := 0; i < n; i++ {
			// pick left channel (both are identical since mono)
			out = append(out, float32(buf[i][0]))
		}
		if !ok {
			break
		}
	}

	return out, nil
}

type floatStreamer struct {
	samples []float32
	pos     int
}

func (s *floatStreamer) Stream(buf [][2]float64) (n int, ok bool) {
	for i := 0; i < len(buf) && s.pos < len(s.samples); i++ {
		v := float64(s.samples[s.pos])
		buf[i][0] = v
		buf[i][1] = v
		s.pos++
		n++
	}
	return n, s.pos < len(s.samples)
}
func (s *floatStreamer) Err() error {
	return nil
}
