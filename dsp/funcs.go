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
		// FIXME: It seems like LFOs might be update once per sample period but at
		// different times (ref: // http://www.spinsemi.com/forum/viewtopic.php?p=5086#p5086)
		state.UpdateRampLFOs()
		state.UpdateSineLFOs()

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
		state.UpdateSineLFOs()
		state.UpdateRampLFOs()
		cycles += 1
	}

	return true // Lets continue!
}

/*
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

	lfo := lfoValue + 0.5
	for lfo > 1.0 {
		lfo -= 0.5
	}

	utils.Assert(lfo <= 1.0 && lfo >= 0.0, "LFO range outside [0 .. 1] (was %f)", lfo)
	return lfo
}

/*
Return a LFO value scaled with the amplitude value specified in the state

	NOTE: The scaled value is an integer, ie *NOT* <0 .. 1.0>
*/
func ScaleLFOValue(value float64, lfoType int, state *State) float64 {
	amp := 1.0
	switch lfoType {
	// FIXME: Why do we have to divide with 16? The amp scaling is way to
	// high... (20260121 handegar)
	case base.LFO_SIN0, base.LFO_COS0:
		amp = float64(state.GetRegister(base.SIN0_RANGE).Value) / 16.0
	case base.LFO_SIN1, base.LFO_COS1:
		amp = float64(state.GetRegister(base.SIN1_RANGE).Value) / 16.0

	case base.LFO_RMP0:
		amp = float64(state.GetRegister(base.RAMP0_RANGE).ToInt32())
	case base.LFO_RMP1:
		amp = float64(state.GetRegister(base.RAMP1_RANGE).ToInt32())
	}

	return value * amp
}

/*
Return the normalized LFO value
ie. a value from  <-1.0 .. 1.0>

Set the 'storeValue' parameter to TRUE when the "REG" keyword is
used for the "CHO RDA" instruction.

"REG" is set => fetch LFO value from system, store for later use. Return Value.
"REG" not set => Return whatever is stored last time REG was specified.
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
			utils.Assert(false, "Invalid LFO type: %d", lfoType)
		}
	}

	switch lfoType {
	case base.LFO_SIN0: // Sine
		lfo = state.Sin0Osc.GetSine()
		state.sin0LFOReg.SetFloat64(lfo)
	case base.LFO_SIN1:
		lfo = state.Sin1Osc.GetSine()
		state.sin1LFOReg.SetFloat64(lfo)
	case base.LFO_COS0: // Cosine
		lfo = state.Sin0Osc.GetCosine()
		state.sin0LFOReg.SetFloat64(lfo)
	case base.LFO_COS1:
		lfo = state.Sin1Osc.GetCosine()
		state.sin1LFOReg.SetFloat64(lfo)

	case base.LFO_RMP0: // Ramps
		lfo = float64(state.Ramp0Osc.GetValue())
		utils.Assert(lfo <= 1.0 && lfo >= 0.0, "LFO Ramp0 range outside [0 .. 1.0] (was %f)", lfo)
		state.ramp0LFOReg.SetFloat64(lfo)

	case base.LFO_RMP1:
		lfo = float64(state.Ramp1Osc.GetValue())
		utils.Assert(lfo <= 1.0 && lfo >= 0.0, "LFO Ramp1 range outside [0 .. 1.0] (was %f)", lfo)
		state.ramp1LFOReg.SetFloat64(lfo)

	default:
		utils.Assert(false, "Unknown LFO type: %d", lfoType)
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

// Outputs a [0 .. 1.0] range.
func GetXFadeFromLFO(lfo float64, typ int, state *State) float64 {
	if isSinLFO(typ) {
		panic("Cannot crossfade a SIN LFO")
	}

	val := 0.0
	if lfo < 0.5 {
		val = lfo * 1.0
	} else {
		val = (1 - lfo) * 1.0
	}

	state.DebugFlags.XFadeMax = math.Max(state.DebugFlags.XFadeMax, val)
	state.DebugFlags.XFadeMin = math.Min(state.DebugFlags.XFadeMin, val)

	return val
}

// Output normalized
func GetRampRange(typ int, state *State) float64 {
	if typ == base.LFO_RMP0 {
		return float64(state.Registers[base.RAMP0_RANGE].ToInt32()) / 4096.0
	} else if typ == base.LFO_RMP1 {
		return float64(state.Registers[base.RAMP1_RANGE].ToInt32()) / 4096.0
	}

	panic("Only RAMPx types allowed")
}

// Ensure the DelayRAM index is within bounds
func capDelayRAMIndex(in int, state *State) (int, error) {
	var err error = nil

	if in > DELAY_RAM_SIZE {
		err = errors.New(fmt.Sprintf("DelayRAM index out of bounds: %d (size=%d, IP=%d)",
			in, DELAY_RAM_SIZE, state.IP))
	}

	return in & (DELAY_RAM_SIZE - 1), err
}
