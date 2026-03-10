package audio

import "time"

type SoundType int

const (
	SoundYourTurn SoundType = iota
	SoundHoleCards
	SoundChip
	SoundOpponentPressure
	SoundCheck
	SoundFold
	SoundAllIn
	SoundFlop
	SoundTurnRiver
	SoundShowdown
	SoundPotWon
	SoundBlindIncrease
	SoundBustout
	SoundVictory
	SoundDefeat
	SoundInvalidAction

	soundCount
)

func generateSoundBank() map[SoundType][]byte {
	sounds := make(map[SoundType][]byte, soundCount)
	for sound := SoundType(0); sound < soundCount; sound++ {
		sounds[sound] = encodeFloat32LE(withSafetyPadding(soundPCM(sound)))
	}
	return sounds
}

func cooldownFor(sound SoundType) time.Duration {
	switch sound {
	case SoundYourTurn:
		return 800 * time.Millisecond
	case SoundHoleCards:
		return 500 * time.Millisecond
	case SoundChip:
		return 120 * time.Millisecond
	case SoundOpponentPressure:
		return 220 * time.Millisecond
	case SoundCheck:
		return 100 * time.Millisecond
	case SoundFold:
		return 120 * time.Millisecond
	case SoundAllIn:
		return 500 * time.Millisecond
	case SoundFlop:
		return 300 * time.Millisecond
	case SoundTurnRiver:
		return 220 * time.Millisecond
	case SoundShowdown:
		return 500 * time.Millisecond
	case SoundPotWon:
		return time.Second
	case SoundBlindIncrease:
		return 2 * time.Second
	case SoundBustout:
		return 1200 * time.Millisecond
	case SoundVictory, SoundDefeat:
		return 2 * time.Second
	case SoundInvalidAction:
		return 250 * time.Millisecond
	default:
		return 0
	}
}

func allSoundTypes() []SoundType {
	all := make([]SoundType, 0, soundCount)
	for sound := SoundType(0); sound < soundCount; sound++ {
		all = append(all, sound)
	}
	return all
}

func soundPCM(sound SoundType) []float32 {
	switch sound {
	case SoundYourTurn:
		return concatBuffers(
			fadedSine(523.25, 0.09, 0.28),
			silence(0.03),
			fadedSine(659.25, 0.11, 0.24),
		)
	case SoundHoleCards:
		return concatBuffers(
			cardFlick(940, 0.09),
			silence(0.016),
			cardFlick(860, 0.08),
		)
	case SoundChip:
		return noiseBurst(0.016, 0.18)
	case SoundOpponentPressure:
		return mixBuffers(
			noiseBurst(0.012, 0.08),
			applyEnvelope(fadedSine(293.66, 0.05, 0.06), 0.01, 0.10, 0.55, 0.50),
		)
	case SoundCheck:
		return applyEnvelope(fadedSine(620, 0.014, 0.10), 0.02, 0.08, 0.5, 0.50)
	case SoundFold:
		return applyEnvelope(sweep(659.25, 622.25, 0.08, 0.18), 0.02, 0.08, 0.7, 0.45)
	case SoundAllIn:
		return applyEnvelope(sweep(262, 1047, 0.15, 0.24), 0.01, 0.10, 0.9, 0.35)
	case SoundFlop:
		return mixBuffers(
			concatBuffers(
				cardFlick(940, 0.08),
				silence(0.010),
				cardFlick(860, 0.08),
				silence(0.010),
				cardFlick(790, 0.07),
			),
			delayBuffer(applyEnvelope(fadedSine(440, 0.10, 0.05), 0.02, 0.10, 0.65, 0.45), 0.008),
		)
	case SoundTurnRiver:
		return mixBuffers(
			cardFlick(820, 0.085),
			delayBuffer(applyEnvelope(fadedSine(659.25, 0.05, 0.035), 0.01, 0.10, 0.55, 0.45), 0.004),
		)
	case SoundShowdown:
		return mixBuffers(
			applyEnvelope(fadedSine(392.00, 0.16, 0.07), 0.02, 0.08, 0.75, 0.45),
			delayBuffer(applyEnvelope(fadedSine(440.00, 0.14, 0.05), 0.02, 0.08, 0.70, 0.45), 0.018),
		)
	case SoundPotWon:
		return mixBuffers(
			delayBuffer(fadedSine(523.25, 0.15, 0.18), 0),
			delayBuffer(fadedSine(659.25, 0.15, 0.17), 0.06),
			delayBuffer(fadedSine(783.99, 0.17, 0.16), 0.12),
		)
	case SoundBlindIncrease:
		return concatBuffers(
			fadedSine(880, 0.055, 0.18),
			silence(0.05),
			fadedSine(880, 0.055, 0.18),
		)
	case SoundBustout:
		return concatBuffers(
			applyEnvelope(fadedSine(392.00, 0.08, 0.09), 0.02, 0.08, 0.65, 0.40),
			silence(0.02),
			applyEnvelope(sweep(349.23, 196.00, 0.14, 0.10), 0.01, 0.08, 0.60, 0.55),
		)
	case SoundVictory:
		return applyEnvelope(mixBuffers(
			fadedSine(523.25, 0.33, 0.16),
			fadedSine(659.25, 0.33, 0.14),
			fadedSine(783.99, 0.33, 0.13),
		), 0.03, 0.10, 0.9, 0.45)
	case SoundDefeat:
		return concatBuffers(
			applyEnvelope(fadedSine(440.00, 0.08, 0.08), 0.02, 0.08, 0.6, 0.45),
			silence(0.02),
			applyEnvelope(fadedSine(349.23, 0.10, 0.07), 0.02, 0.08, 0.6, 0.45),
			silence(0.02),
			applyEnvelope(fadedSine(293.66, 0.14, 0.06), 0.02, 0.08, 0.55, 0.55),
		)
	case SoundInvalidAction:
		return applyEnvelope(sweep(520, 430, 0.045, 0.12), 0.01, 0.08, 0.5, 0.55)
	default:
		return nil
	}
}

func cardFlick(tailFreq, amplitude float64) []float32 {
	return mixBuffers(
		noiseBurst(0.010, amplitude),
		delayBuffer(applyEnvelope(fadedSine(tailFreq, 0.022, amplitude*0.35), 0.01, 0.08, 0.35, 0.55), 0.002),
	)
}

func withSafetyPadding(buf []float32) []float32 {
	if len(buf) == 0 {
		return nil
	}
	padded := make([]float32, 0, len(buf)+durationToFrames(safetyPadDuration)*channels*2)
	padded = append(padded, silence(safetyPadDuration)...)
	padded = append(padded, buf...)
	padded = append(padded, silence(safetyPadDuration)...)
	return padded
}

func delayBuffer(buf []float32, delay float64) []float32 {
	if len(buf) == 0 {
		return nil
	}
	if delay <= 0 {
		out := make([]float32, len(buf))
		copy(out, buf)
		return out
	}
	out := make([]float32, 0, len(buf)+durationToFrames(delay)*channels)
	out = append(out, silence(delay)...)
	out = append(out, buf...)
	return out
}

func concatBuffers(buffers ...[]float32) []float32 {
	totalLen := 0
	for _, buf := range buffers {
		totalLen += len(buf)
	}
	if totalLen == 0 {
		return nil
	}
	out := make([]float32, 0, totalLen)
	for _, buf := range buffers {
		out = append(out, buf...)
	}
	return out
}
