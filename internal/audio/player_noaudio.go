//go:build !nativeaudio

package audio

func newAudioBackend() (backend, error) {
	return noopBackend{}, nil
}
