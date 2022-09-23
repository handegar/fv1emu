package dsp

import (
	"errors"
	"fmt"
	"math"

	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/settings"
	"github.com/handegar/fv1emu/utils"
)

var skipNumSamples int = 0

// Return ENUMS for the debug callbacks
const (
	Ok int = iota
	NextInstruction
	Quit
	Fatal
)

type DebugCallback func(opCodes []base.Op, state *State, sampleNum int) int

func SkipNumSamples(num int) {
	skipNumSamples = num
}

func GetSkipNumSamples() int {
	return skipNumSamples
}

func ProcessSample(opCodes []base.Op, state *State, sampleNum int,
	debugPre DebugCallback, debugPost DebugCallback) bool {
	state.IP = 0

	cycles := 0

	for state.IP < uint(len(opCodes)) {
		cycles += 1
		// FIXME: The LFO should probably be updated in sync
		// with an external clock, not per
		// instruction. (20220227 handegar)
		updateSinLFO(state, 0)
		updateSinLFO(state, 1)
		updateRampLFO(state, 0)
		updateRampLFO(state, 1)

		op := opCodes[state.IP]

		if skipNumSamples < 1 {
			debugPre(opCodes, state, sampleNum)
		}

		err := applyOp(op, state)
		if err != nil {
			fmt.Printf("An error occured (IP=%d, Sample=%d):\n",
				state.IP, sampleNum)
			fmt.Println(err)
			state.DebugFlags.Print()
			return false
		}

		if skipNumSamples <= 0 {
			status := debugPost(opCodes, state, sampleNum)
			if status == Fatal || status == Quit {
				return false
			} else if status == NextInstruction {
				continue
			}
		}

		state.IP += 1

		// Did we clip? Update debug info accordingly
		state.CheckForOverflows()
	}

	state.RUN_FLAG = true

	state.DelayRAMPtr -= 1
	if state.DelayRAMPtr <= -32768 {
		state.DelayRAMPtr = 0
	}

	// Decrease debug-skipping count
	if skipNumSamples > 0 {
		skipNumSamples -= 1
	}

	//
	// To ensure the all samples uses 128 cycles we'll have to
	// update the LFOs for remaining cycles to ensure that we stay
	// close to how the FV-1 operates (and sounds).
	//
	for cycles < settings.InstructionsPerSample {
		updateSinLFO(state, 0)
		updateSinLFO(state, 1)
		updateRampLFO(state, 0)
		updateRampLFO(state, 1)
		cycles += 1
	}

	return true // Lets continue!
}

func updateSinLFO(state *State, num int) {
	//
	// Frequency formula (from the datasheet):
	//   f = (RATE*CLOCKFREQ)/(2*PI*(2^17)
	// Allowed values for RATE is 1..512 => 0..20Hz
	//

	// This is a magic constant I have figured out by calibrating the
	// emulated result against a real FV-1 chip. I assume this value is
	// due to some internal mechanism in the chip.
	calibrationFactor := 1 / 4.0

	if num == 0 {
		f0 := float64(state.Registers[base.SIN0_RATE].ToInt32()) //* settings.ClockFrequency
		f0 = f0 / (2.0 * math.Pi * (2 << 11))

		utils.Assert(f0 >= 0.0 && f0 <= 20.0, "Sin0 rate out of range [0..20hz]: %f", f0)

		sin0delta := f0 / float64(settings.InstructionsPerSample)
		state.Sin0State.Angle += sin0delta * calibrationFactor
	} else {
		f1 := float64(state.Registers[base.SIN1_RATE].ToInt32()) //* settings.ClockFrequency
		f1 = f1 / (2.0 * math.Pi * (2 << 11))

		utils.Assert(f1 >= 0.0 && f1 <= 20.0, "Sin1 rate out of range [0..20hz]: %f", f1)

		sin1delta := f1 / float64(settings.InstructionsPerSample)
		state.Sin1State.Angle += sin1delta * calibrationFactor
	}
}

/**
  This generates a sawtooth of [0 .. 0.5].
*/
func updateRampLFO(state *State, num int) {
	//
	// Frequency formula (from the datasheet):
	//   f=RATE/512
	// Allowed values for RATE (16bit) is 0..65565 -> 0..128 hz
	//

	// This is a magic constant I have figured out by calibrating the
	// emulated result against a real FV-1 chip. I assume this value is
	// due to some internal mechanism in the chip.
	calibrationFactor := 8.0

	if num == 0 {
		range0 := base.RampAmpValues[state.Registers[base.RAMP0_RANGE].ToInt32()]

		if range0 > 0.0 {
			rate := state.Registers[base.RAMP0_RATE].ToInt32()
			utils.Assert(rate >= -64*512 && rate <= 128*512,
				"Ramp0 rate out of range [0..128hz]: %d", rate/512)
			delta := (float64(rate) / float64(range0)) / settings.ClockFrequency
			delta = delta / float64(settings.InstructionsPerSample)
			state.Ramp0State.Value -= delta * calibrationFactor
		}

		// New cycle?
		for state.Ramp0State.Value < 0 {
			state.Ramp0State.Value = 0.5
		}
		for state.Ramp1State.Value > 0.5 {
			state.Ramp1State.Value -= 0.5
		}

	} else {
		range1 := base.RampAmpValues[state.Registers[base.RAMP1_RANGE].ToInt32()]
		if range1 > 0.0 {
			rate := state.Registers[base.RAMP1_RATE].ToInt32()
			utils.Assert(rate >= -64*512 && rate <= 128*512,
				"Ramp1 rate out of range1 [0..128hz]: %d", rate/512)

			delta := (float64(rate) / float64(range1)) / settings.ClockFrequency
			delta = delta / float64(settings.InstructionsPerSample)
			state.Ramp1State.Value -= delta * calibrationFactor
		}

		// New cycle?
		for state.Ramp1State.Value < 0 {
			state.Ramp1State.Value += 0.5
		}

		for state.Ramp1State.Value > 0.5 {
			state.Ramp1State.Value -= 0.5
		}

	}

}

/**
  Returns the LFO value, but 1/2 further into the cycle.
  NB: This is only valid for RAMP LFOs.

  // FIXME: This func needs to be re-checked to see if it is
  // correct. We should also get rid of the daft while/for loop by
  // using some bitmask'ing on integers instead.
  // (20220917 handegar)
*/
func GetLFOValuePlusHalfCycle(lfoType int, lfoValue float64) float64 {
	if isSinLFO(lfoType) {
		panic("Cannot call GetLFOValuePlusHalfCycle() for SIN LFOs")
	}

	lfo := lfoValue - 0.5
	for lfo < 0 {
		lfo += 0.5
	}

	utils.Assert(lfo <= 0.5 && lfo >= 0.0, "LFO range outside [0 .. 0.5] (was %f)", lfo)
	return lfo
}

/**
  Return a LFO value scaled with the amplitude value specified in the state
*/
func ScaleLFOValue(value float64, lfoType int, state *State) float64 {

	amp := 1.0
	switch lfoType {
	case base.LFO_SIN0, base.LFO_COS0:
		amp = float64(state.GetRegister(base.SIN0_RANGE).ToInt32())
	case base.LFO_SIN1, base.LFO_COS1:
		amp = float64(state.GetRegister(base.SIN1_RANGE).ToInt32())
	case base.LFO_RMP0:
		amp = float64(base.RampAmpValues[state.GetRegister(base.RAMP0_RANGE).ToInt32()])
	case base.LFO_RMP1:
		amp = float64(base.RampAmpValues[state.GetRegister(base.RAMP1_RANGE).ToInt32()])
	}

	result := value * amp

	/*
		utils.Assert(result < 1.0 && result >= -1.0,
			"ScaleLFOValue(%f, %d, state) generated value out of range [-1.0, 0.999] (%f)",
				value, lfoType, result)
	*/
	return result
}

/**
  Will "normalize" the LFO value to a number between [-1.0 .. 0.999]
*/
func NormalizeLFOValue(value float64, lfoType int, state *State) float64 {
	sinFactor := 1.0 / float64((1<<15)-1)
	rmpFactor := 1.0 / float64((1<<12)-1)

	ret := 0.0
	switch lfoType {
	case base.LFO_SIN0, base.LFO_COS0:
		ret = value * sinFactor
	case base.LFO_SIN1, base.LFO_COS1:
		ret = value * sinFactor
	case base.LFO_RMP0:
		ret = value * rmpFactor
	case base.LFO_RMP1:
		ret = value * rmpFactor
	}

	utils.Assert(ret < 1.0 && ret >= -1.0,
		"NormalizeLFOValue(%f, %d, state) generated value out of range [-1.0, 0.999] (%f)",
		value, lfoType, ret)

	return ret
}

/**
  Return the normalized LFO value
  ie. a value from  <-1.0 .. 1.0>

  Set the 'storeValue' parameter to TRUE when the "REG" keyword is
  used for the "CHO RDA" instruction.
*/
func GetLFOValue(lfoType int, state *State, storeValue bool) float64 {
	lfo := 0.0

	if !storeValue {
		switch lfoType {
		case base.LFO_SIN0, base.LFO_COS0:
			return state.sin0LFOReg.ToFloat64()
		case base.LFO_SIN1, base.LFO_COS1:
			return state.sin1LFOReg.ToFloat64()
		case base.LFO_RMP0:
			return state.ramp0LFOReg.ToFloat64()
		case base.LFO_RMP1:
			return state.ramp1LFOReg.ToFloat64()
		default:
			panic("Invalid LFO index")
		}
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

	switch lfoType {
	case base.LFO_SIN0: // Sine
		lfo = math.Sin(state.Sin0State.Angle)
		state.sin0LFOReg.SetFloat64(lfo)
	case base.LFO_SIN1:
		lfo = math.Sin(state.Sin1State.Angle)
		state.sin1LFOReg.SetFloat64(lfo)

	case base.LFO_COS0: // Cosine
		lfo = math.Cos(state.Sin0State.Angle)
		state.sin0LFOReg.SetFloat64(lfo)
	case base.LFO_COS1:
		lfo = math.Cos(state.Sin1State.Angle)
		state.sin1LFOReg.SetFloat64(lfo)

	case base.LFO_RMP0: // Ramps
		lfo = state.Ramp0State.Value
		utils.Assert(lfo <= 0.5 && lfo >= 0.0, "LFO Ramp0 range outside [0 .. 0.5] (was %f)", lfo)
		state.ramp0LFOReg.SetFloat64(lfo)
	case base.LFO_RMP1:
		lfo = state.Ramp1State.Value
		utils.Assert(lfo <= 0.5 && lfo >= 0.0, "LFO Ramp1 range outside [0 .. 0.5] (was %f)", lfo)
		state.ramp1LFOReg.SetFloat64(lfo)

	default:
		panic("Unknown LFO type")
	}

	// Debugging
	if lfoType == base.LFO_SIN0 {
		state.DebugFlags.Sin0Max = math.Max(state.DebugFlags.Sin0Max, lfo)
		state.DebugFlags.Sin0Min = math.Min(state.DebugFlags.Sin0Min, lfo)

	} else if lfoType == base.LFO_SIN1 {
		state.DebugFlags.Sin1Max = math.Max(state.DebugFlags.Sin1Max, lfo)
		state.DebugFlags.Sin1Min = math.Min(state.DebugFlags.Sin1Min, lfo)

	} else if lfoType == base.LFO_SIN1 {
		state.DebugFlags.Ramp0Max = math.Max(state.DebugFlags.Ramp0Max, lfo)
		state.DebugFlags.Ramp0Min = math.Min(state.DebugFlags.Ramp0Min, lfo)

	} else if lfoType == base.LFO_SIN1 {
		state.DebugFlags.Ramp1Max = math.Max(state.DebugFlags.Ramp1Max, lfo)
		state.DebugFlags.Ramp1Min = math.Min(state.DebugFlags.Ramp1Min, lfo)
	}

	return lfo
}

//
// This expects an lfo-input of [0 .. 1.0], ie. not scaled according to RANGE.
// Outputs a [0 .. 1.0] range.
//
func GetXFadeFromLFO__OLD(lfo float64, typ int, state *State) float64 {
	if isSinLFO(typ) {
		panic("Cannot crossfade a SIN LFO")
	}

	/**
		   We want the ramp to look like this over the LFO
	           period (0 .. 0.5):

		                  _______
		                 /       \
	  	          ____/         \____
	           Phase   0   1   2   3   4

		   We'll divide it into 5 phases: Start, rise, rest, sink, end
	*/

	/*
	   UPDATE:
	   It seems to me that the AN-0001 illustration of an xfade envelope is abit incorrect.

	*/

	phaseWidth := (1.0 / 5.0) / 2.0
	startPhase := phaseWidth / 2.0
	risePhase := startPhase + phaseWidth
	restPhase := risePhase + phaseWidth
	sinkPhase := restPhase + phaseWidth

	x := 10.0
	val := 0.0
	if lfo > 0.0 && lfo < startPhase {
		val = 0.0
	} else if lfo > startPhase && lfo < risePhase {
		val = (lfo - startPhase) * x
	} else if lfo > risePhase && lfo < restPhase {
		val = 0.9999
	} else if lfo > restPhase && lfo < sinkPhase {
		val = 0.9999 - (lfo-restPhase)*x
	} else { // End phase
		val = 0.0
	}

	state.DebugFlags.XFadeMax = math.Max(state.DebugFlags.XFadeMax, val)
	state.DebugFlags.XFadeMin = math.Min(state.DebugFlags.XFadeMin, val)

	return val
}

//
// This expects an lfo-input of [0 .. 1.0], ie. not scaled according to RANGE.
// Outputs a [0 .. 0.5] ranged "pyramid".
//
func GetXFadeFromLFO(lfo float64, typ int, state *State) float64 {
	if isSinLFO(typ) {
		panic("Cannot crossfade a SIN LFO")
	}

	amp := 0.0
	if typ == 0 {
		amp = float64(base.RampAmpValues[state.Registers[base.RAMP0_RANGE].ToInt32()])
	} else {
		amp = float64(base.RampAmpValues[state.Registers[base.RAMP1_RANGE].ToInt32()])
	}

	xfade := 0.25 - lfo
	if lfo > 0.25 {
		xfade = lfo - 0.25
	}

	val := (xfade * amp) / 4096.0

	utils.Assert(val < 1.0 && val >= -1.0, "XFadeValue out of range (was %f)", val)
	return val
}

func GetRampRange(typ int, state *State) float64 {
	if typ == base.LFO_RMP0 {
		return float64(state.Registers[base.RAMP0_RANGE].ToInt32())
	} else if typ == base.LFO_RMP1 {
		return float64(state.Registers[base.RAMP1_RANGE].ToInt32())
	}

	panic("Only RAMPx types allowed")
	return 0.0
}

// Ensure the DelayRAM index is within bounds
func capDelayRAMIndex(in int, state *State) (int, error) {
	var err error = nil

	if in > DELAY_RAM_SIZE {
		err = errors.New(fmt.Sprintf("DelayRAM index out of bounds: %d (len=%d)",
			in, DELAY_RAM_SIZE))
	}

	return in & (DELAY_RAM_SIZE - 1), err
}
