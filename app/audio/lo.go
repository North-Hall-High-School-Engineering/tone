package audio

type Loopback interface {
	Start() error

	Stop() error

	Read() ([]float32, error)
}

func NewLoopback() Loopback {
	return newLoopback()
}
