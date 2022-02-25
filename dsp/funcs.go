package dsp

import (
	"errors"
	"fmt"
	"math"

	ui "github.com/gizak/termui/v3"

	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/settings"
)

var skipNumSamples int = 0

func ProcessSample(opCodes []base.Op, state *State, sampleNum int) bool {
	state.IP = 0

	// FIXME: Update LFO each sample for now to keep thinks
	// simple. (20220222 handegar)
	updateSinLFO(state)
	updateRampLFO(state)

	// Used to keep previous states within one sample for
	// "rewinding" when debugging.
	previousStates := make([]*State, len(opCodes))

	for state.IP < uint(len(opCodes)) {
		if int(state.IP) > len(opCodes) { // The program has ended.
			break
		}

		op := opCodes[state.IP]

		// FIXME: We should update the LFO at every cycle, not
		// just every sample. This is how the FV-1 does
		// it. (20220222 handegar)
		//updateLFOStates(state)

		// Print pre-op state
		if settings.Debugger && skipNumSamples <= 0 {
			// Append to the state-history for this sample
			if previousStates[state.IP] == nil {
				previousStates[state.IP] = state.Duplicate()
			}
			UpdateDebuggerScreen(opCodes, state, sampleNum)
		}

		err := applyOp(op, state)
		if err != nil {
			if settings.Debugger {
				ui.Close()
			}
			fmt.Printf("An error occured (IP=%d, Sample=%d):\n",
				state.IP, sampleNum)
			fmt.Println(err)
			state.DebugFlags.Print()
			panic(false)
		}

		if settings.Debugger && skipNumSamples <= 0 {
			e := WaitForDebuggerInput(state)
			switch e {
			case "quit":
				return false
			case "next op":
				break
			case "previous op":
				if state.IP > 0 {
					state.Copy(previousStates[state.IP-1])
				}
				continue
			case "next sample":
				skipNumSamples = 1
				break
			case "next 100 samples":
				skipNumSamples = 100
				break
			case "next 1000 samples":
				skipNumSamples = 1000
				break
			case "next 10000 samples":
				skipNumSamples = 10000
				break
			case "next 100000 samples":
				skipNumSamples = 100000
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
	if skipNumSamples > 0 {
		skipNumSamples -= 1
	}

	//
	// To ensure the all samples uses 128 cycles we'll have to
	// update the LFOs for remaining cycles to ensure that we stay
	// close to how the FV-1 operates (and sounds).
	//
	// FIXME: This could probably be done with one operation
	// instead of a loop. (20220225 handegar)
	// FIXME: This does not work until I get the "LFO update each
	// cycle" working. Now it will just make the LFO go way faster
	// than expected. (20220225 handegar)
	/*
		remainingCycles := 128 - len(opCodes)
		for i := 0; i < remainingCycles; i++ {
			updateSinLFO(state)
			updateRampLFO(state)
		}
	*/
	return true // Lets continue!
}

func updateSinLFO(state *State) {
	// Sin LFO range is from 0.5Hz to 20Hz
	sin0Freq := 0.5 + (state.Registers[base.SIN0_RATE].ToFloat64() * (20 - 0.5))
	sin1Freq := 0.5 + (state.Registers[base.SIN1_RATE].ToFloat64() * (20 - 0.5))

	// Update Sine-LFOs
	sin0delta := (2 * math.Pi * (sin0Freq - 0.5)) / settings.SampleRate
	state.Sin0State.Angle += sin0delta

	sin1delta := (2 * math.Pi * (sin1Freq - 0.5)) / settings.SampleRate
	state.Sin1State.Angle += sin1delta
}

//
// NB: This generates a sawtooth of [0 .. 1.0]. FV-1 uses [0 .. 0.5]
// so this value needs to be scaled when needed.
//
func updateRampLFO(state *State) {
	rate0 := float64(state.Registers[base.RAMP0_RATE].ToInt32())
	range0 := float64(state.Registers[base.RAMP0_RANGE].ToInt32())
	rate1 := float64(state.Registers[base.RAMP1_RATE].ToInt32())
	range1 := float64(state.Registers[base.RAMP1_RANGE].ToInt32())

	if rate0 != 0 {
		fcount := float64(state.Ramp0State.count)
		state.Ramp0State.Value = ((fcount * (range0 / rate0)) / (range0 + 1))
	}

	if rate1 != 0 {
		fcount := float64(state.Ramp1State.count)
		state.Ramp1State.Value = ((fcount * (range1 / rate1)) / (range1 + 1))
	}

	for state.Ramp0State.Value > 1.0 {
		state.Ramp0State.Value -= 1.0
	}

	for state.Ramp1State.Value > 1.0 {
		state.Ramp1State.Value -= 1.0
	}

	state.Ramp0State.count += 1
	state.Ramp1State.count += 1
	if float64(state.Ramp0State.count) > rate0 {
		state.Ramp0State.count = 0
	}
	if float64(state.Ramp1State.count) > rate1 {
		state.Ramp1State.count = 0
	}
}

/**
  Returns the LFO value, but 1/2 further into the cycle.
  NB: This is only valid for RAMP LFOs.
*/
func GetLFOValuePlusHalfCycle(lfoType int, state *State) float64 {
	if isSinLFO(lfoType) {
		panic("Cannot call GetLFOValuePlusHalfCycle() for SIN LFOs")
	}

	// Save the original RAMP state
	rmp0count := state.Ramp0State.count
	rmp1count := state.Ramp1State.count
	rmp0Value := state.Ramp0State.Value
	rmp1Value := state.Ramp1State.Value
	lfo := 0.0

	if lfoType == base.LFO_RMP0 {
		rate0 := state.Registers[base.RAMP0_RATE].ToInt32()
		state.Ramp0State.count += int(rate0 >> 1)
		updateRampLFO(state)
		lfo = state.Ramp0State.Value
	} else {
		rate1 := state.Registers[base.RAMP1_RATE].ToInt32()
		state.Ramp1State.count += int(rate1 >> 1)
		updateRampLFO(state)
		lfo = state.Ramp1State.Value
	}

	// Restore the state
	state.Ramp0State.count = rmp0count
	state.Ramp1State.count = rmp1count
	state.Ramp0State.Value = rmp0Value
	state.Ramp1State.Value = rmp1Value

	assert(lfo < 1.0 && lfo >= 0.0, "LFO range outside [0 .. 1]")
	return lfo
}

/**
  Return a LFO value scaled with the amplitude value specified in the state
*/
func ScaleLFOValue(value float64, lfoType int, state *State) float64 {
	amp := 1.0
	switch lfoType {
	case base.LFO_SIN0, base.LFO_COS0:
		amp = float64(state.GetRegister(base.SIN0_RANGE).ToInt32()) / (1 << 23)
	case base.LFO_SIN1, base.LFO_COS1:
		amp = float64(state.GetRegister(base.SIN1_RANGE).ToInt32()) / (1 << 23)
	case base.LFO_RMP0:
		amp = float64(state.GetRegister(base.RAMP0_RANGE).ToInt32()) / (1 << 10)
	case base.LFO_RMP1:
		amp = float64(state.GetRegister(base.RAMP1_RANGE).ToInt32()) / (1 << 10)
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
		case base.LFO_SIN0, base.LFO_SIN1, base.LFO_COS0, base.LFO_COS1:
			return state.sinLFOReg.ToFloat64()
		case base.LFO_RMP0, base.LFO_RMP1:
			return state.rampLFOReg.ToFloat64()
		default:
			panic("Invalid LFO index")
		}
	}

	switch lfoType {
	case base.LFO_SIN0:
		lfo = math.Sin(state.Sin0State.Angle)
		state.sinLFOReg.SetFloat64(lfo)
	case base.LFO_SIN1:
		lfo = math.Sin(state.Sin1State.Angle)
		state.sinLFOReg.SetFloat64(lfo)
	case base.LFO_RMP0:
		lfo = state.Ramp0State.Value
		state.rampLFOReg.SetFloat64(lfo)
	case base.LFO_RMP1:
		lfo = state.Ramp1State.Value
		state.rampLFOReg.SetFloat64(lfo)
	case base.LFO_COS0:
		lfo = math.Cos(state.Sin0State.Angle)
		state.sinLFOReg.SetFloat64(lfo)
	case base.LFO_COS1:
		lfo = math.Cos(state.Sin1State.Angle)
		state.sinLFOReg.SetFloat64(lfo)
	default:
		panic("Unknown LFO type")
	}

	assert(lfo <= 1.0 && lfo >= -1.0, "LFO range outside [-1 .. 1]")
	return lfo
}

//
// This expects an lfo-input of [0 .. 1.0], ie. not scaled according to RANGE.
// Outputs a [0 .. 1.0] range.
func GetXFadeFromLFO(lfo float64, typ int, state *State) float64 {
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

	phaseWidth := (1.0 / 5.0)
	startPhase := phaseWidth
	risePhase := (2 * phaseWidth)
	restPhase := (3 * phaseWidth)
	sinkPhase := (4 * phaseWidth)

	val := 0.0
	if lfo > 0.0 && lfo < startPhase {
		val = 0.0
	} else if lfo > startPhase && lfo < risePhase {
		val = (lfo - startPhase) * 5.0
	} else if lfo > risePhase && lfo < restPhase {
		val = 1.0
	} else if lfo > restPhase && lfo < sinkPhase {
		val = 1.0 - (lfo-restPhase)*5.0
	} else { // End phase
		val = 0.0
	}

	return val
}

// Ensure the DelayRAM index is within bounds
func capDelayRAMIndex(in int, state *State) (int, error) {
	var err error = nil
	if in > DELAY_RAM_SIZE {
		err = errors.New(fmt.Sprintf("DelayRAM index out of bounds: %d (len=%d)",
			in, DELAY_RAM_SIZE))
	}

	return in & 0x7fff, err
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
