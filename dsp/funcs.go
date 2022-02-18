package dsp

import (
	"fmt"
	"math"
	"syscall"

	"github.com/eiannone/keyboard"
	"github.com/fatih/color"

	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/disasm"
	"github.com/handegar/fv1emu/settings"
)

const debugPrompt = "< (N)ext op | Next (s)ample | (V)iew state | (P)rint op | (Q)uit >"

func ProcessSample(opCodes []base.Op, state *State) {
	state.IP = 0
	skipToNextSample := false

	clockDelta := settings.ClockFrequency / settings.SampleRate

	for state.IP = 0; state.IP < uint(len(opCodes)); {
		op := opCodes[state.IP]

		// Print pre-op state
		if settings.StepDebug && skipToNextSample != true {
			color.Blue("IP=%d (of %d), ACC=%d(%f), ADDR_PTR=%d, DelayRAMPtr=%d",
				state.IP, len(opCodes), state.ACC.Value, state.ACC.ToFloat64(),
				state.Registers[base.ADDR_PTR].Value, state.DelayRAMPtr)
			color.Cyan(disasm.OpCodeToString(op))
		}

		updateLFOStates(state, clockDelta)
		applyOp(op, state)

		if settings.StepDebug && skipToNextSample != true {
			// Print post-op state
			color.White("  => ACC=%d(%f), ADDR_PTR=%d",
				state.ACC.Value, state.ACC.ToFloat64(), state.Registers[base.ADDR_PTR].Value)

			fmt.Println()
			color.Yellow(debugPrompt)
			for skipToNextSample != true {
				char, _, err := keyboard.GetKey()

				if err != nil {
					fmt.Printf("ERROR: %d\n", err)
					return
				}

				if char == 'q' {
					_ = keyboard.Close()
					syscall.Exit(1)
				} else if char == 'p' {
					color.Cyan(disasm.OpCodeToString(op))
					color.Yellow(debugPrompt)
				} else if char == 'v' {
					state.Print()
					color.Yellow(debugPrompt)
				} else if char == 'n' {
					break
				} else if char == 's' {
					skipToNextSample = true
					color.Red("Skipping to next sample")
					break
				}
			}
		}

		state.IP += 1
	}

	state.RUN_FLAG = true
	state.PACC = state.ACC
	state.ACC.Clear()

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
}

func updateLFOStates(state *State, clockDelta float64) {
	// FIXME: This is measured by hand but I am not sure if it is
	// correct. Investigate. (20220207 handegar)

	x := 1024.0

	// Update Sine-LFO
	sin0delta := state.Registers[base.SIN0_RATE].ToFloat64() * clockDelta
	sin1delta := state.Registers[base.SIN1_RATE].ToFloat64() * clockDelta
	state.Sin0State.Angle += sin0delta / x
	state.Sin1State.Angle += sin1delta / x

	// Update Ramp-LFO
	ramp0delta := state.Registers[base.RAMP0_RATE].ToFloat64() * clockDelta
	ramp1delta := state.Registers[base.RAMP1_RATE].ToFloat64() * clockDelta

	// NOTE: Ramp-values are always positive according to the FV-1 spec.
	state.Ramp0State.Value += ramp0delta
	if state.Ramp0State.Value > state.Registers[base.RAMP0_RANGE].ToFloat64() {
		state.Ramp0State.Value = 0
	}

	state.Ramp1State.Value += ramp1delta
	if state.Ramp1State.Value > state.Registers[base.RAMP1_RANGE].ToFloat64() {
		state.Ramp1State.Value = 0
	}
}

// Ret -1.0 .. 1.0
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
			return 0
		}
	}

	switch lfoType {
	case 0:
		lfo = (math.Sin(state.Sin0State.Angle) *
			float64(state.GetRegister(base.SIN0_RANGE).ToInt32())) / 2.0
		state.sinLFOReg.SetFloat64(lfo)
	case 1:
		lfo = (math.Sin(state.Sin1State.Angle) *
			float64(state.GetRegister(base.SIN1_RANGE).Value>>4)) / 2.0
		state.sinLFOReg.SetFloat64(lfo)
	case 2:
		lfo = state.Ramp0State.Value *
			state.GetRegister(base.RAMP0_RANGE).ToFloat64()
		state.rampLFOReg.SetFloat64(lfo)
	case 3:
		lfo = state.Ramp1State.Value *
			state.GetRegister(base.RAMP1_RANGE).ToFloat64()
		state.rampLFOReg.SetFloat64(lfo)
	case 4:
		lfo = (math.Cos(state.Sin0State.Angle) *
			state.GetRegister(base.SIN0_RANGE).ToFloat64())
		state.sinLFOReg.SetFloat64(lfo)
	case 5:
		lfo = (math.Cos(state.Sin1State.Angle) *
			state.GetRegister(base.SIN1_RANGE).ToFloat64())
		state.sinLFOReg.SetFloat64(lfo)
	}
	return lfo
}

func GetLFOMaximum(lfoType int, state *State) float64 {
	// FIXME: This is measured by hand but I am not sure if it is
	// correct. Investigate. (20220207 handegar)
	magic := (2.0 * math.Pi) / settings.SampleRate

	switch lfoType {
	case 0:
		return state.Registers[base.SIN0_RANGE].ToFloat64() * magic
	case 1:
		return state.Registers[base.SIN1_RANGE].ToFloat64() * magic
	case 2:
		return 2.0 * state.Registers[base.RAMP0_RANGE].ToFloat64()
	case 3:
		return 2.0 * state.Registers[base.RAMP1_RANGE].ToFloat64()
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
