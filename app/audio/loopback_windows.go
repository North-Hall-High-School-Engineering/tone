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

	var data *C.BYTE
	var numFrames C.UINT32

	ret := C.loopback_read(w.lb, &data, &numFrames)
	if ret <= 0 || data == nil {
		return nil, nil
	}
	defer C.free(unsafe.Pointer(data)) // free buffer

	length := int(numFrames) * 2 // if stereo
	audio := make([]float32, length)

	// Convert raw 32-bit float PCM
	src := (*[1 << 28]C.float)(unsafe.Pointer(data))[:length:length]
	for i := 0; i < length; i++ {
		audio[i] = float32(src[i])
	}

	return audio, nil
}
