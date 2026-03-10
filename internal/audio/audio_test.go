package audio

import (
	"sync"
	"testing"
	"time"
)

type fakeBackend struct {
	mu      sync.Mutex
	plays   int
	volumes []float64
	closed  bool
}

func (f *fakeBackend) Play(_ []byte, volume float64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.plays++
	f.volumes = append(f.volumes, volume)
}

func (f *fakeBackend) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	return nil
}

func (f *fakeBackend) Available() bool { return true }

func TestAllSoundsGenerate(t *testing.T) {
	bank := generateSoundBank()
	for _, sound := range allSoundTypes() {
		buf := bank[sound]
		if len(buf) == 0 {
			t.Fatalf("sound %v generated empty buffer", sound)
		}
	}
}

func TestSoundDurations(t *testing.T) {
	for _, sound := range allSoundTypes() {
		frames := len(withSafetyPadding(soundPCM(sound))) / channels
		duration := float64(frames) / sampleRate
		if duration > 0.5 {
			t.Fatalf("sound %v duration %.3fs exceeds limit", sound, duration)
		}
	}
}

func TestPlayWhenDisabled(t *testing.T) {
	fb := &fakeBackend{}
	mgr := newManager(func() (backend, error) { return fb, nil })
	_ = mgr.Init()
	mgr.SetEnabled(false)
	mgr.Play(SoundChip)
	if fb.plays != 0 {
		t.Fatalf("plays = %d, want 0", fb.plays)
	}
}

func TestPlayWhenUnavailable(t *testing.T) {
	mgr := newManager(func() (backend, error) { return noopBackend{}, nil })
	_ = mgr.Init()
	mgr.Play(SoundChip)
}

func TestSetVolumeClamps(t *testing.T) {
	mgr := newManager(nil)
	mgr.SetVolume(2)
	if mgr.volume != 1 {
		t.Fatalf("volume = %v, want 1", mgr.volume)
	}
	mgr.SetVolume(-1)
	if mgr.volume != 0 {
		t.Fatalf("volume = %v, want 0", mgr.volume)
	}
}

func TestConcurrentPlay(t *testing.T) {
	fb := &fakeBackend{}
	mgr := newManager(func() (backend, error) { return fb, nil })
	_ = mgr.Init()

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mgr.Play(SoundVictory)
		}()
	}
	wg.Wait()
	if fb.plays == 0 {
		t.Fatal("expected at least one play")
	}
}

func TestInitIdempotent(t *testing.T) {
	fb := &fakeBackend{}
	count := 0
	mgr := newManager(func() (backend, error) {
		count++
		return fb, nil
	})
	_ = mgr.Init()
	_ = mgr.Init()
	if count != 1 {
		t.Fatalf("factory count = %d, want 1", count)
	}
}

func TestCooldownBlocksRapidReplay(t *testing.T) {
	fb := &fakeBackend{}
	mgr := newManager(func() (backend, error) { return fb, nil })
	_ = mgr.Init()
	mgr.Play(SoundChip)
	mgr.Play(SoundChip)
	if fb.plays != 1 {
		t.Fatalf("plays = %d, want 1", fb.plays)
	}
	time.Sleep(cooldownFor(SoundChip))
	mgr.Play(SoundChip)
	if fb.plays != 2 {
		t.Fatalf("plays = %d, want 2", fb.plays)
	}
}

func TestPerceptualVolumeIsMonotonic(t *testing.T) {
	low := perceptualVolume(0.25)
	mid := perceptualVolume(0.5)
	high := perceptualVolume(0.75)
	if !(low < mid && mid < high) {
		t.Fatalf("expected monotonic volume curve, got low=%f mid=%f high=%f", low, mid, high)
	}
}

func TestCloseCallsBackend(t *testing.T) {
	fb := &fakeBackend{}
	mgr := newManager(func() (backend, error) { return fb, nil })
	_ = mgr.Init()
	mgr.Close()
	if !fb.closed {
		t.Fatal("expected backend to be closed")
	}
}
