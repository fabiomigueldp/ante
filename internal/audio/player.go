package audio

type backend interface {
	Play(buf []byte, volume float64)
	Close() error
	Available() bool
}

type backendFactory func() (backend, error)

type noopBackend struct{}

func (noopBackend) Play([]byte, float64) {}

func (noopBackend) Close() error { return nil }

func (noopBackend) Available() bool { return false }
