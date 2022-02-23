package dsp

import (
	"fmt"
	"math"

	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/settings"
)

func ProcessSample(opCodes []base.Op, state *State, sampleNum int) bool {
	state.IP = 0
	skipToNextSample := false

	clockDelta := settings.ClockFrequency / settings.SampleRate

	// FIXME: Update LFO each sample for now to keep thinks
	// simple. (20220222 handegar)
	updateLFOStates(state, clockDelta)

	for state.IP < uint(len(opCodes)) {
		if int(state.IP) > len(opCodes) { // The program has ended.
			break
		}

		op := opCodes[state.IP]

		// FIXME: We should update the LFO at every cycle, not
		// just every sample. This is how the FV-1 does
		// it. (20220222 handegar)
		//updateLFOStates(state, clockDelta)

		// Print pre-op state
		if settings.Debugger && skipToNextSample != true {
			UpdateDebuggerScreen(opCodes, state, sampleNum)
		}

		applyOp(op, state)

		if settings.Debugger && skipToNextSample != true {
			e := WaitForDebuggerInput(state)
			switch e {
			case "quit":
				return false
			case "next op":
				break
			case "next sample":
				skipToNextSample = true
				break
			}
		}

		state.IP += 1

		// Did we clip? Update debug info accordingly
		state.CheckForOverflows()
	}

	state.RUN_FLAG = true
	state.PACC.Copy(state.ACC)

	// FIXME: Should we clear ACC? (20220222 handegar)
	//state.ACC.Clear()

	state.DelayRAMPtr -= 1
	if state.DelayRAMPtr <= -32768 {
		state.DelayRAMPtr = 0
	}

	// Stop debug-skipping
	skipToNextSample = false

	// FIXME: Shall we wait/NOP for the remaining operations so
	// that we always execute 128 instructions each sample? To
	// ensure the LFOs are correct? This might be audible on
	// really short programs vs. larger programs. (20220217
	// handegar)

	return true // Lets continue!
}

func updateLFOStates(state *State, clockDelta float64) {
	// Sin LFO range is from 0.5Hz to 20Hz
	sin0Freq := 0.5 + (state.Registers[base.SIN0_RATE].ToFloat64() * (20 - 0.5))
	sin1Freq := 0.5 + (state.Registers[base.SIN1_RATE].ToFloat64() * (20 - 0.5))

	// Update Sine-LFOs
	sin0delta := (2 * math.Pi * sin0Freq) / settings.SampleRate
	state.Sin0State.Angle += sin0delta

	sin1delta := (2 * math.Pi * sin1Freq) / settings.SampleRate
	state.Sin1State.Angle += sin1delta

	// Ramp LFO range is -16384 to 32767 (-0.5, 0.99)
	ramp0delta := state.Registers[base.RAMP0_RATE].ToFloat64() / settings.SampleRate
	ramp1delta := state.Registers[base.RAMP1_RATE].ToFloat64() / settings.SampleRate

	// NOTE: Ramp-values are always positive according to the FV-1 spec.
	state.Ramp0State.Value += ramp0delta / (1 << 8)
	if state.Ramp0State.Value > 0.5 {
		state.Ramp0State.Value = 0.0
	}

	state.Ramp1State.Value += (ramp1delta / 32768.0)
	if state.Ramp1State.Value > 0.5 {
		state.Ramp1State.Value = 0.0
	}

}

/**
  Return a LFO value scaled with the amplitude value specified in the state
*/
func ScaleLFOValue(value float64, lfoType int, state *State) float64 {
	amp := 1.0
	switch lfoType {
	case 0, 4:
		amp = float64(state.GetRegister(base.SIN0_RANGE).ToInt32()) / (1 << 23)
	case 1, 5:
		amp = float64(state.GetRegister(base.SIN1_RANGE).ToInt32()) / (1 << 23)
	case 2:
		amp = float64(state.GetRegister(base.RAMP0_RANGE).ToInt32())
	case 3:
		amp = float64(state.GetRegister(base.RAMP1_RANGE).ToInt32())
	}

	return value * amp
}

/**
  Return the normalized LFO value
  ie. a value from  <-1.0 .. 1.0>
*/
func GetLFOValue(lfoType int, state *State, retrieveOnly bool) float64 {
	lfo := 0.0

	if retrieveOnly {
		switch lfoType {
		case 0, 1, 4, 5:
			return state.sinLFOReg.ToFloat64()
		case 2, 3:
			return state.rampLFOReg.ToFloat64()
		default:
			panic("Invalid LFO index")
		}
	}

	switch lfoType {
	case 0:
		lfo = math.Sin(state.Sin0State.Angle)
		state.sinLFOReg.SetFloat64(lfo)
	case 1:
		lfo = math.Sin(state.Sin1State.Angle)
		state.sinLFOReg.SetFloat64(lfo)
	case 2:
		lfo = state.Ramp0State.Value
		state.rampLFOReg.SetFloat64(lfo)
	case 3:
		lfo = state.Ramp1State.Value
		state.rampLFOReg.SetFloat64(lfo)
	case 4:
		lfo = math.Cos(state.Sin0State.Angle)
		state.sinLFOReg.SetFloat64(lfo)
	case 5:
		lfo = math.Cos(state.Sin1State.Angle)
		state.sinLFOReg.SetFloat64(lfo)
	default:
		panic("Unknown LFO type")
	}
	return lfo
}

func GetLFOMaximum(lfoType int, state *State) float64 {
	switch lfoType {
	case 0:
		return float64(state.Registers[base.SIN0_RANGE].ToInt32())
	case 1:
		return float64(state.Registers[base.SIN1_RANGE].ToInt32())
	case 2:
		return float64(state.Registers[base.RAMP0_RANGE].ToInt32())
	case 3:
		return float64(state.Registers[base.RAMP1_RANGE].ToInt32())
	}
	return 0.0
}

// Ensure the DelayRAM index is within bounds
func capDelayRAMIndex(in int) int {
	if in > DELAY_RAM_SIZE {
		fmt.Printf("ERROR: DelayRAM index out of bounds: %d (len=%d)\n",
			in, DELAY_RAM_SIZE)
		//panic(false)
	}

	return in & 0x7fff
}

func clampInteger(in int) int {
	if in > 0x7fffff {
		return 0x7fffff
	}
	if in < -0x800000 {
		return -0x800000
	}
	return in
}
