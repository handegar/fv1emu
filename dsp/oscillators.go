package dsp

import (
	//"fmt"

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
	value int32
	freq  int32
	amp   int32 // 512, 1024, 2048 or 4096
}

func (r *RampOscillator) Update() {
	utils.Assert(r.freq >= -16384 && r.freq <= 32767,
		"Ramp0 rate out of range [-16384 .. 38767]: %d", r.freq)
	r.value -= r.freq >> 2
	if r.value < 0 {
		r.value = 0x7FFFFFFF
	}
}

func (r *RampOscillator) SetFreq(freq int32) {
	utils.Assert(freq < 32769 && freq >= -16384, "Ramp freq out of range: %d", freq)
	r.freq = freq
}

// Input: 0, 1, 2 or 3
func (r *RampOscillator) SetAmpIdx(ampIdx int8) {
	utils.Assert(ampIdx >= 0 && ampIdx <= 3, "Invalid AmpIdx value: %d", ampIdx)
	r.amp = 512 << ampIdx
}

// Will choose a matching ampidx.
// Input: 512, 1024, 2048 or 4096
func (r *RampOscillator) SetAmp(amp int32) {
	utils.Assert(amp == 512 || amp == 1024 || amp == 2048 || amp == 2096,
		"Invalid Ramp amp value")
	r.amp = amp
}

// Always returns a value between 0 and 0.5
// FIXME: Return <0 .. 1.0> or <0 .. 0.5>? (20260202 handegar)
func (r *RampOscillator) GetValue() float64 {
	// FIXME: Here we could do a quick bitshift instead I think (20260202 handegar)
	//fmt.Printf("r.val=%d\n", r.value)
	return float64(r.value) / float64(0x7FFFFFFF)
}

func (r *RampOscillator) Reset() {
	r.value = 0
}
