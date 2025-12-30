package audio

type Loopback interface {
	// Initalize captue
	Start() error

	// Stop capture
	Stop() error

	// Next chunk of audio as f32 PCM
	Read() ([]float32, error)
}

func NewLoopback() Loopback {
	return &windowsLoopback{}
}
