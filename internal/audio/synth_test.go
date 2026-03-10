package audio

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestSineWaveLength(t *testing.T) {
	buf := sineWave(440, 0.1, 0.5)
	want := durationToFrames(0.1) * channels
	if len(buf) != want {
		t.Fatalf("len(buf) = %d, want %d", len(buf), want)
	}
}

func TestSineWaveAmplitude(t *testing.T) {
	buf := sineWave(440, 0.1, 0.25)
	max := 0.0
	for _, sample := range buf {
		max = math.Max(max, math.Abs(float64(sample)))
	}
	if max > 0.251 {
		t.Fatalf("max amplitude = %.4f, want <= 0.251", max)
	}
}

func TestFadedSineNoClicks(t *testing.T) {
	buf := fadedSine(440, 0.12, 0.3)
	for i := 0; i < 10*channels; i++ {
		if math.Abs(float64(buf[i])) > 0.05 {
			t.Fatalf("start sample[%d] = %.4f, want near zero", i, buf[i])
		}
	}
	for i := len(buf) - 10*channels; i < len(buf); i++ {
		if math.Abs(float64(buf[i])) > 0.02 {
			t.Fatalf("end sample[%d] = %.4f, want near zero", i, buf[i])
		}
	}
}

func TestSweepFrequencyRangeProducesFiniteValues(t *testing.T) {
	buf := sweep(220, 880, 0.2, 0.5)
	for i, sample := range buf {
		if math.IsNaN(float64(sample)) || math.IsInf(float64(sample), 0) {
			t.Fatalf("sample[%d] is invalid: %v", i, sample)
		}
	}
}

func TestNoiseBurstDeterministic(t *testing.T) {
	a := noiseBurst(0.02, 0.2)
	b := noiseBurst(0.02, 0.2)
	if len(a) != len(b) {
		t.Fatalf("length mismatch %d != %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("noise burst differs at %d", i)
		}
	}
}

func TestMixBuffersNormalization(t *testing.T) {
	a := sineWave(440, 0.1, 0.9)
	b := sineWave(660, 0.1, 0.9)
	mixed := mixBuffers(a, b)
	max := 0.0
	for _, sample := range mixed {
		max = math.Max(max, math.Abs(float64(sample)))
	}
	if max > 1.0001 {
		t.Fatalf("mixed peak = %.4f, want <= 1.0", max)
	}
}

func TestMixBuffersDifferentLengths(t *testing.T) {
	short := sineWave(440, 0.05, 0.2)
	long := sineWave(440, 0.1, 0.2)
	mixed := mixBuffers(short, long)
	if len(mixed) != len(long) {
		t.Fatalf("len(mixed) = %d, want %d", len(mixed), len(long))
	}
}

func TestSilenceIsZero(t *testing.T) {
	buf := silence(0.05)
	for i, sample := range buf {
		if sample != 0 {
			t.Fatalf("sample[%d] = %v, want 0", i, sample)
		}
	}
}

func TestEnvelopeShape(t *testing.T) {
	buf := make([]float32, durationToFrames(0.1)*channels)
	for i := range buf {
		buf[i] = 1
	}
	shaped := applyEnvelope(buf, 0.2, 0.2, 0.5, 0.2)
	if shaped[0] <= 0 || shaped[0] >= 1 {
		t.Fatalf("expected attack shaping, got %v", shaped[0])
	}
	mid := shaped[len(shaped)/2]
	if mid < 0.45 || mid > 0.55 {
		t.Fatalf("mid sustain = %v, want near 0.5", mid)
	}
}

func TestEncodeFloat32LE(t *testing.T) {
	buf := []float32{0, 1, -1}
	encoded := encodeFloat32LE(buf)
	if len(encoded) != len(buf)*bytesPerSample {
		t.Fatalf("len(encoded) = %d", len(encoded))
	}
	if got := math.Float32frombits(binary.LittleEndian.Uint32(encoded[4:8])); got != 1 {
		t.Fatalf("decoded sample = %v, want 1", got)
	}
}
