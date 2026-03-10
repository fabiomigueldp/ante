package audio

import (
	"math"
	"sync"
	"time"
)

type manager struct {
	factory    backendFactory
	backend    backend
	available  bool
	enabled    bool
	volume     float64
	initErr    error
	initOnce   sync.Once
	soundsOnce sync.Once

	mu        sync.RWMutex
	sounds    map[SoundType][]byte
	cooldowns map[SoundType]time.Time
}

var defaultManager = newManager(newAudioBackend)

func Init() error {
	return defaultManager.Init()
}

func Close() {
	defaultManager.Close()
}

func Play(sound SoundType) {
	defaultManager.Play(sound)
}

func SetVolume(v float64) {
	defaultManager.SetVolume(v)
}

func SetEnabled(on bool) {
	defaultManager.SetEnabled(on)
}

func IsAvailable() bool {
	return defaultManager.IsAvailable()
}

func newManager(factory backendFactory) *manager {
	return &manager{
		factory:    factory,
		backend:    noopBackend{},
		enabled:    true,
		volume:     defaultVolume,
		cooldowns:  make(map[SoundType]time.Time),
		available:  false,
		initErr:    nil,
		sounds:     nil,
		soundsOnce: sync.Once{},
	}
}

func (m *manager) Init() error {
	m.ensureSounds()
	m.initOnce.Do(func() {
		m.mu.Lock()
		m.available = false
		m.mu.Unlock()
		if m.factory == nil {
			return
		}
		backend, err := m.factory()
		if err != nil {
			m.initErr = err
			return
		}
		m.mu.Lock()
		m.backend = backend
		m.available = backend.Available()
		m.mu.Unlock()
	})
	return m.initErr
}

func (m *manager) Close() {
	m.mu.RLock()
	backend := m.backend
	m.mu.RUnlock()
	_ = backend.Close()
}

func (m *manager) Play(sound SoundType) {
	m.ensureSounds()

	m.mu.Lock()
	if !m.enabled || !m.available {
		m.mu.Unlock()
		return
	}
	buf, ok := m.sounds[sound]
	if !ok || len(buf) == 0 {
		m.mu.Unlock()
		return
	}
	now := time.Now()
	if cooldown := cooldownFor(sound); cooldown > 0 {
		if next, ok := m.cooldowns[sound]; ok && now.Before(next) {
			m.mu.Unlock()
			return
		}
		m.cooldowns[sound] = now.Add(cooldown)
	}
	backend := m.backend
	volume := perceptualVolume(m.volume)
	m.mu.Unlock()

	backend.Play(buf, volume)
}

func (m *manager) SetEnabled(on bool) {
	m.mu.Lock()
	m.enabled = on
	m.mu.Unlock()
}

func (m *manager) SetVolume(v float64) {
	m.mu.Lock()
	m.volume = clamp01(v)
	m.mu.Unlock()
}

func (m *manager) IsAvailable() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.available
}

func (m *manager) ensureSounds() {
	m.soundsOnce.Do(func() {
		m.mu.Lock()
		m.sounds = generateSoundBank()
		m.mu.Unlock()
	})
}

func perceptualVolume(v float64) float64 {
	return math.Pow(clamp01(v), 1.6)
}
