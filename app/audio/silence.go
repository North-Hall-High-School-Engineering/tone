package audio

func IsSilent(samples []float32) bool {
	maxAmp := float32(0)

	for _, v := range samples {
		abs := v
		if abs < 0 {
			abs = -abs
		}
		if abs > maxAmp {
			maxAmp = abs
		}
	}

	return maxAmp == 0
}
