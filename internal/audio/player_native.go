//go:build nativeaudio

package audio

import (
	"bytes"
	"time"

	"github.com/ebitengine/oto/v3"
)

type audioPlayer struct {
	ctx   *oto.Context
	ready <-chan struct{}
}

func newAudioBackend() (backend, error) {
	ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: channels,
		Format:       oto.FormatFloat32LE,
	})
	if err != nil {
		return noopBackend{}, err
	}
	return &audioPlayer{ctx: ctx, ready: ready}, nil
}

func (p *audioPlayer) Play(buf []byte, volume float64) {
	if p == nil || p.ctx == nil || len(buf) == 0 || volume <= 0 {
		return
	}

	go func(data []byte, level float64) {
		select {
		case <-p.ready:
		case <-time.After(2 * time.Second):
			return
		}

		player := p.ctx.NewPlayer(bytes.NewReader(data))
		player.SetVolume(level)
		player.Play()

		deadline := time.Now().Add(2 * time.Second)
		for player.IsPlaying() && time.Now().Before(deadline) {
			time.Sleep(10 * time.Millisecond)
		}
		_ = player.Close()
	}(buf, volume)
}

func (p *audioPlayer) Close() error {
	if p == nil || p.ctx == nil {
		return nil
	}
	return p.ctx.Suspend()
}

func (p *audioPlayer) Available() bool {
	return p != nil && p.ctx != nil
}
