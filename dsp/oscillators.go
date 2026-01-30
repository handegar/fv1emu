package dsp

import (
	//"fmt"
	"github.com/handegar/fv1emu/settings"
	"github.com/handegar/fv1emu/utils"
	"math"
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
	factor := ((2.0 * math.Pi) / settings.ClockFrequency)

	// Calibrated so that max freq -> sine of 20 hz
	f0 := 4.0 * s.freq / 512.0
	utils.Assert(f0 >= 0.0 && f0 <= 20.0, "Sin0 rate out of range [0..20hz]: %f", f0)
	s.value += f0 * factor
}

func (s *SineOscillator) SetFreq(freq int32) {
	s.freq = (float64(freq) / 512.0) * 20.0
}
func (s *SineOscillator) SetAmp(amp int32) {
	s.amp = float64(amp) / 32767.0
}

//
// A possible optimization could be to roll our own SIN/COS
// function like this:
//   https://news.ycombinator.com/item?id=30844872
//
// The hardware sine-function in the FV-1 is not 64-bit as
// GOLang's "math.sin()" so there is no need for this kind of
// super-precision. A simple 24-bit will be more than enough
// and probably too precise for anyone to notice.
//

//
// NOTE: One thing I have seen while calibrating against a real FV-1
// is that the chip's sine function is quite uneven and more
// flattend on the bottom than the top. Maybe I should experiment with
// approximations to match the asymetric look on the real deal.
//

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
	value  float64
	freq   int32
	ampIdx int8 // 0..3 => 4k, 2k, 1k, 512
}

func (r *RampOscillator) Update() {
	utils.Assert(r.freq >= -16384 && r.freq <= 32767,
		"Ramp0 rate out of range [-16384 .. 38767]: %d", r.freq)

	// Calibrated so that max freq -> saw of 20 hz
	//delta := (float64(r.freq) / settings.ClockFrequency) / float64(settings.InstructionsPerSample) / 1024.0

	delta := ((float64(r.freq) / settings.ClockFrequency) / 4096.0) / float64(settings.InstructionsPerSample)

	//delta := float64(r.freq) / 4096.0

	r.value -= delta
	//fmt.Printf("val=%f, delta=%f, freq=%d\n", r.value, delta, r.freq)

	if r.value >= 1.0 {
		r.value -= 1.0
	}
	if r.value < 0 {
		r.value += 1.0
	}
}

func (r *RampOscillator) SetFreq(freq int32) {
	utils.Assert(freq < 32769 && freq > -16384, "Ramp freq out of range: %d", freq)
	r.freq = freq
}

// Input: 0, 1, 2 or 3
func (r *RampOscillator) SetAmpIdx(amp int8) {
	utils.Assert(amp >= 0 && amp <= 3, "Invalid amp value: %d", amp)
	r.ampIdx = amp
}

// Will choose a matching ampidx.
// Input 0 .. 1.0
func (r *RampOscillator) SetAmp(amp float64) {
	r.ampIdx = int8(amp * 4)
}

// Always returns a value between 0 and 0.5
func (r *RampOscillator) GetValue() float64 {
	return r.value
}

func (r *RampOscillator) Reset() {
	r.value = 0
}
