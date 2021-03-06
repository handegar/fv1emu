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

	for state.IP < uint(len(opCodes)) {
		// FIXME: The LFO should probably be updated in sync
		// with an external clock, not per
		// instruction. (20220227 handegar)
		updateSinLFO(state)
		updateRampLFO(state)

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
	// FIXME: Should we clear ACC? (20220222 handegar)
	//state.ACC.Clear()

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
	remainingCycles := settings.InstructionsPerSample - len(opCodes)
	for i := 0; i < remainingCycles; i++ {
		updateSinLFO(state)
		updateRampLFO(state)
	}

	return true // Lets continue!
}

func updateSinLFO(state *State) {
	// Sin LFO range is from 0.5Hz to 20Hz
	sin0Freq := 0.5 + (state.Registers[base.SIN0_RATE].ToFloat64() * (20 - 0.5))
	sin1Freq := 0.5 + (state.Registers[base.SIN1_RATE].ToFloat64() * (20 - 0.5))

	// FIXME: What is the correct value here? (20220310 handegar)
	// 168.0: Tuned "sin-lfo.spn" to 100Hz for POT1=0.0
	// 8.0: Tuned "misc/tremolo-1.spn" to get a credible max/min-rate
	x := 8.0

	// Update Sine-LFOs
	sin0delta := ((2 * math.Pi * (sin0Freq - 0.5)) / settings.SampleRate) * x
	state.Sin0State.Angle += sin0delta

	sin1delta := ((2 * math.Pi * (sin1Freq - 0.5)) / settings.SampleRate) * x
	state.Sin1State.Angle += sin1delta
}

//
// This generates a sawtooth of [0 .. 0.5].
//
func updateRampLFO(state *State) {
	rate0 := float64(state.Registers[base.RAMP0_RATE].ToInt32() + 1)
	rate1 := float64(state.Registers[base.RAMP1_RATE].ToInt32() + 1)

	// FIXME: What is the correct value here? (20220310 handegar)
	// - 56.0: To get the "ramp-lfo.spn" to get 100Hz with POT1=0.0
	// - 16385.0: Handtuning OEM1_6.bin (flanger) (half clockspeed?)
	//x := 16385.0
	x := 1000.0

	state.Ramp0State.Value += (1.0 / rate0) / x
	state.Ramp1State.Value += (1.0 / rate1) / x

	for state.Ramp0State.Value > 0.5 {
		state.Ramp0State.Value -= 0.5
	}

	for state.Ramp1State.Value > 0.5 {
		state.Ramp1State.Value -= 0.5
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
	rmp0value := state.Ramp0State.Value
	rmp1value := state.Ramp1State.Value
	lfo := 0.0

	if lfoType == base.LFO_RMP0 {
		state.Ramp0State.Value += (0.5 / 2.0)
		updateRampLFO(state)
		lfo = state.Ramp0State.Value
	} else {
		state.Ramp1State.Value += (0.5 / 2.0)
		updateRampLFO(state)
		lfo = state.Ramp1State.Value
	}

	// Restore the state
	state.Ramp0State.Value = rmp0value
	state.Ramp1State.Value = rmp1value

	utils.Assert(lfo <= 0.5 && lfo >= 0.0, "LFO range outside [0 .. 0.5]")
	return lfo
}

/**
  Return a LFO value scaled with the amplitude value specified in the state
*/
func ScaleLFOValue(value float64, lfoType int, state *State) float64 {
	// FIXME: Get this right (20220311 handegar)
	// 2.0: The tri-lfo.spn program.
	// 32.0: misc/tremolo-1.spn (SpinCAD, virker rart)
	// 2.0: vibrato-2.spn
	sinX := 2.0

	// FIXME: Get this right (20220311 handegar)
	// 2.0: oem-1/OEM1_6.spn (flanger)
	rmpX := 2.0

	amp := 1.0
	switch lfoType {
	case base.LFO_SIN0, base.LFO_COS0:
		amp = float64(state.GetRegister(base.SIN0_RANGE).ToInt32()) / sinX
	case base.LFO_SIN1, base.LFO_COS1:
		amp = float64(state.GetRegister(base.SIN1_RANGE).ToInt32()) / sinX
	case base.LFO_RMP0:
		amp = float64(state.GetRegister(base.RAMP0_RANGE).ToInt32()) / rmpX
	case base.LFO_RMP1:
		amp = float64(state.GetRegister(base.RAMP1_RANGE).ToInt32()) / rmpX
	}

	return value * amp
}

/**
  Will "normalize" the LFO value to a number between [-1 .. 1]
*/
func NormalizeLFOValue(value float64, lfoType int, state *State) float64 {
	sinFactor := float64((1 << 9) - 1)

	// Which value is correct?
	// 1<<11: Handtuned "calibrate/ramp-lfo.spn"
	rampFactor := float64((1 << 11) - 1)

	ret := 0.0
	switch lfoType {
	case base.LFO_SIN0, base.LFO_COS0:
		ret = value / sinFactor
	case base.LFO_SIN1, base.LFO_COS1:
		ret = value / sinFactor
	case base.LFO_RMP0:
		ret = value / rampFactor
	case base.LFO_RMP1:
		ret = value / rampFactor
	}

	return ret
}

/**
  Return the normalized LFO value
  ie. a value from  <-1.0 .. 1.0>
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

	switch lfoType {
	case base.LFO_SIN0:
		lfo = math.Sin(state.Sin0State.Angle)
		state.sin0LFOReg.SetFloat64(lfo)
	case base.LFO_SIN1:
		lfo = math.Sin(state.Sin1State.Angle)
		state.sin1LFOReg.SetFloat64(lfo)
	case base.LFO_COS0:
		lfo = math.Cos(state.Sin0State.Angle)
		state.sin0LFOReg.SetFloat64(lfo)
	case base.LFO_COS1:
		lfo = math.Cos(state.Sin1State.Angle)
		state.sin1LFOReg.SetFloat64(lfo)
	case base.LFO_RMP0:
		lfo = state.Ramp0State.Value
		utils.AssertFloat64(lfo <= 0.5 && lfo >= 0.0, lfo,
			"LFO Ramp0 range outside [0 .. 0.5]")
		state.ramp0LFOReg.SetFloat64(lfo)
	case base.LFO_RMP1:
		lfo = state.Ramp1State.Value
		utils.AssertFloat64(lfo <= 0.5 && lfo >= 0.0, lfo,
			"LFO Ramp1 range outside [0 .. 0.5]")
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
// Expects a [min .. max] input. Returns a [0 .. 1.0] output.
//
func GetLFOComplement(lfo float64, min float64, max float64) float64 {
	upShift := (0.0 - min) / 2.0
	ret := (lfo / (max - min)) + upShift

	utils.Assert(ret >= 0.0 && ret <= 1.0, "Complement value is < 0 || > 1.0")
	return 1.0 - ret
}

//
// This expects an lfo-input of [0 .. 1.0], ie. not scaled according to RANGE.
// Outputs a [0 .. 1.0] range.
//
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
		val = 1.0
	} else if lfo > restPhase && lfo < sinkPhase {
		val = 1.0 - (lfo-restPhase)*x
	} else { // End phase
		val = 0.0
	}

	/*
		fmt.Printf("XFade: lfo=%f -> val=%f (%f, %f, %f, %f)\n",
			lfo, val, startPhase, risePhase, restPhase, sinkPhase)
	*/

	state.DebugFlags.XFadeMax = math.Max(state.DebugFlags.XFadeMax, val)
	state.DebugFlags.XFadeMin = math.Min(state.DebugFlags.XFadeMin, val)

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
