package main

const SAMPLING_RATE = 16000           // 16kHz
const WINDOW_SIZE = 2 * SAMPLING_RATE // 2 sec
const HOP_SIZE = SAMPLING_RATE / 2

type RingBuffer struct {
	data []float32
	pos  int
	full bool
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data: make([]float32, size),
	}
}

func (rb *RingBuffer) Add(samples []float32) {
	for _, s := range samples {
		rb.data[rb.pos] = s
		rb.pos++
		if rb.pos >= len(rb.data) {
			rb.pos = 0
			rb.full = true
		}
	}
}

func (rb *RingBuffer) GetWindow() []float32 {
	if rb.full {
		out := make([]float32, len(rb.data))
		copy(out, rb.data[rb.pos:])
		copy(out[len(rb.data)-rb.pos:], rb.data[:rb.pos])
		return out
	}
	return rb.data[:rb.pos]
}
