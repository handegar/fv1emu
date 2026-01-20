package dsp

import (
	"math"

	"github.com/handegar/fv1emu/settings"
	"github.com/handegar/fv1emu/utils"
)

//
// Sine/Cosine oscillator (LFO)
//

type SineOscillator struct {
	value float64
	freq  float64 // normalized 0..1
	amp   float64 // normalized 0..1
}

func (s *SineOscillator) Update() {
	factor := (2.0 * math.Pi) / settings.ClockFrequency
	f0 := 20.0 * s.freq / 512.0
	utils.Assert(f0 >= 0.0 && f0 <= 20.0, "Sin0 rate out of range [0..20hz]: %f", f0)
	s.value += f0 * factor
}

func (s *SineOscillator) SetFreq(freq int32) {
	s.freq = (float64(freq) / 512.0) * 20.0
}
func (s *SineOscillator) SetAmp(amp int32) {
	s.amp = float64(amp) / 32767.0
}

func (s *SineOscillator) GetSine() float64 {
	return math.Sin(s.value)
}

func (s *SineOscillator) GetCosine() float64 {
	return math.Cos(s.value)
}

//
// Ramp oscillator (LFO)
//

type RampOscillator struct {
	value float64
	freq  int32
	amp   int16
}

func (r *RampOscillator) Update() {
	utils.Assert(r.freq >= -16384 && r.freq <= 32767,
		"Ramp0 rate out of range [-16384 .. 38767]: %d", r.freq)

	delta := (float64(r.freq) / 16384.0) / settings.ClockFrequency
	if r.amp < 0.0 {
		delta = -delta
	}

	r.value -= delta

	// New cycle?
	for r.value < RAMP_START {
		r.value = RAMP_END
	}
	for r.value > RAMP_END {
		r.value = RAMP_START
	}
}

func (r *RampOscillator) SetFreq(freq int32) {
	r.freq = freq
}

func (r *RampOscillator) SetAmp(amp int16) {
	r.amp = amp
}

func (r *RampOscillator) GetValue() float64 {
	return r.value
}

func (r *RampOscillator) Reset() {
	r.value = RAMP_START
}
