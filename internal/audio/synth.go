package audio

import (
	"encoding/binary"
	"math"
	"math/rand"
)

const (
	sampleRate         = 44100
	channels           = 2
	safetyPadDuration  = 0.002
	bytesPerSample     = 4
	defaultVolume      = 0.7
	defaultNoiseSeed   = 1337
	defaultFadeOutExpo = 2.4
)

func sineWave(freq, duration, amplitude float64) []float32 {
	frames := durationToFrames(duration)
	if frames == 0 {
		return nil
	}

	amp := clamp01(amplitude)
	buf := make([]float32, frames*channels)
	for i := 0; i < frames; i++ {
		sample := float32(math.Sin(2*math.Pi*freq*float64(i)/sampleRate) * amp)
		base := i * channels
		buf[base] = sample
		buf[base+1] = sample
	}
	return buf
}

func fadedSine(freq, duration, amplitude float64) []float32 {
	buf := sineWave(freq, duration, amplitude)
	frames := len(buf) / channels
	if frames == 0 {
		return nil
	}

	fadeInFrames := maxInt(1, int(math.Round(float64(frames)*0.05)))
	fadeOutFrames := maxInt(1, int(math.Round(float64(frames)*0.40)))
	out := make([]float32, len(buf))
	copy(out, buf)

	for frame := 0; frame < frames; frame++ {
		gain := 1.0
		if frame < fadeInFrames {
			gain *= float64(frame+1) / float64(fadeInFrames)
		}
		if frame >= frames-fadeOutFrames {
			progress := float64(frame-(frames-fadeOutFrames)+1) / float64(fadeOutFrames)
			gain *= math.Pow(maxFloat(0, 1-progress), defaultFadeOutExpo)
		}
		base := frame * channels
		out[base] *= float32(gain)
		out[base+1] *= float32(gain)
	}

	return out
}

func sweep(freqStart, freqEnd, duration, amplitude float64) []float32 {
	frames := durationToFrames(duration)
	if frames == 0 {
		return nil
	}

	amp := clamp01(amplitude)
	buf := make([]float32, frames*channels)
	phase := 0.0
	for i := 0; i < frames; i++ {
		progress := 0.0
		if frames > 1 {
			progress = float64(i) / float64(frames-1)
		}
		freq := freqStart + (freqEnd-freqStart)*progress
		phase += 2 * math.Pi * freq / sampleRate
		sample := float32(math.Sin(phase) * amp)
		base := i * channels
		buf[base] = sample
		buf[base+1] = sample
	}
	return buf
}

func noiseBurst(duration, amplitude float64) []float32 {
	frames := durationToFrames(duration)
	if frames == 0 {
		return nil
	}

	rng := rand.New(rand.NewSource(defaultNoiseSeed))
	amp := clamp01(amplitude)
	buf := make([]float32, frames*channels)
	prev := 0.0
	for i := 0; i < frames; i++ {
		current := (rng.Float64()*2 - 1) * amp
		filtered := clampSample(current - prev*0.82)
		prev = current
		base := i * channels
		buf[base] = float32(filtered)
		buf[base+1] = float32(filtered)
	}
	return applyEnvelope(buf, 0.02, 0.08, 0.35, 0.55)
}

func mixBuffers(buffers ...[]float32) []float32 {
	maxLen := 0
	for _, buf := range buffers {
		if len(buf) > maxLen {
			maxLen = len(buf)
		}
	}
	if maxLen == 0 {
		return nil
	}

	mixed := make([]float32, maxLen)
	peak := 0.0
	for _, buf := range buffers {
		for i, sample := range buf {
			mixed[i] += sample
			abs := math.Abs(float64(mixed[i]))
			if abs > peak {
				peak = abs
			}
		}
	}

	if peak > 1.0 {
		scale := 1.0 / peak
		for i := range mixed {
			mixed[i] *= float32(scale)
		}
	}

	return mixed
}

func silence(duration float64) []float32 {
	frames := durationToFrames(duration)
	if frames == 0 {
		return nil
	}
	return make([]float32, frames*channels)
}

func applyEnvelope(buf []float32, attack, decay, sustain, release float64) []float32 {
	frames := len(buf) / channels
	if frames == 0 {
		return nil
	}

	attack = maxFloat(0, attack)
	decay = maxFloat(0, decay)
	release = maxFloat(0, release)
	sustain = clamp01(sustain)

	total := attack + decay + release
	if total > 1 {
		scale := 1 / total
		attack *= scale
		decay *= scale
		release *= scale
	}

	attackFrames := int(math.Round(float64(frames) * attack))
	decayFrames := int(math.Round(float64(frames) * decay))
	releaseFrames := int(math.Round(float64(frames) * release))
	if attackFrames+decayFrames+releaseFrames > frames {
		overflow := attackFrames + decayFrames + releaseFrames - frames
		releaseFrames = maxInt(0, releaseFrames-overflow)
	}
	sustainFrames := frames - attackFrames - decayFrames - releaseFrames

	out := make([]float32, len(buf))
	for frame := 0; frame < frames; frame++ {
		gain := envelopeGain(frame, frames, attackFrames, decayFrames, sustainFrames, releaseFrames, sustain)
		base := frame * channels
		out[base] = buf[base] * float32(gain)
		out[base+1] = buf[base+1] * float32(gain)
	}
	return out
}

func encodeFloat32LE(buf []float32) []byte {
	if len(buf) == 0 {
		return nil
	}
	out := make([]byte, len(buf)*bytesPerSample)
	for i, sample := range buf {
		binary.LittleEndian.PutUint32(out[i*bytesPerSample:], math.Float32bits(sample))
	}
	return out
}

func durationToFrames(duration float64) int {
	if duration <= 0 {
		return 0
	}
	return int(math.Round(duration * sampleRate))
}

func envelopeGain(frame, totalFrames, attackFrames, decayFrames, sustainFrames, releaseFrames int, sustainLevel float64) float64 {
	switch {
	case attackFrames > 0 && frame < attackFrames:
		return float64(frame+1) / float64(attackFrames)
	case decayFrames > 0 && frame < attackFrames+decayFrames:
		progress := float64(frame-attackFrames+1) / float64(decayFrames)
		return 1 - (1-sustainLevel)*progress
	case frame < attackFrames+decayFrames+sustainFrames:
		return sustainLevel
	case releaseFrames > 0 && frame < totalFrames:
		progress := float64(frame-(totalFrames-releaseFrames)+1) / float64(releaseFrames)
		return maxFloat(0, sustainLevel*(1-progress))
	default:
		return sustainLevel
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func clampSample(v float64) float64 {
	if v < -1 {
		return -1
	}
	if v > 1 {
		return 1
	}
	return v
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
