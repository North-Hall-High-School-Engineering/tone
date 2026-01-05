//go:build windows
// +build windows

package audio

/*
#cgo LDFLAGS: -lole32 -luuid -loleaut32
#include "loopback_windows.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"math"
	"unsafe"
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

	// Convert multi-channel -> mono and find max abs value for normalization
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

	// Normalize to [-1, 1] if maxVal > 0
	// if maxVal > 0 {
	// 	for i := range audio {
	// 		audio[i] /= maxVal
	// 	}
	// }

	// // Compute RMS
	// var sumSquares float32
	// for _, v := range audio {
	// 	sumSquares += v * v
	// }

	return ResampleTo16kHz(audio, int(pwfx.nSamplesPerSec)), nil
}

func ResampleTo16kHz(input []float32, inRate int) []float32 {
	outRate := 16000
	outLen := len(input) * outRate / inRate
	out := make([]float32, outLen)
	ratio := float64(len(input)-1) / float64(outLen-1)
	for i := 0; i < outLen; i++ {
		pos := ratio * float64(i)
		idx := int(pos)
		frac := float32(pos - float64(idx))
		if idx+1 < len(input) {
			out[i] = input[idx]*(1-frac) + input[idx+1]*frac
		} else {
			out[i] = input[idx]
		}
	}
	return out
}
