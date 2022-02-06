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

	for state.IP = 0; state.IP < uint(len(opCodes)); {
		op := opCodes[state.IP]

		// Print pre-op state
		if settings.StepDebug && skipToNextSample != true {
			color.Blue("IP=%d (of %d), ACC=%f, ADDR_PTR=%d, DelayRAMPtr=%d",
				state.IP, len(opCodes), state.ACC,
				state.Registers[base.ADDR_PTR].(int), state.DelayRAMPtr)
			color.Cyan(disasm.OpCodeToString(op))
		}

		applyOp(op, state)

		if settings.StepDebug && skipToNextSample != true {

			// Print post-op state
			color.White("  => ACC=%f, ADDR_PTR=%d",
				state.ACC, state.Registers[base.ADDR_PTR].(int))

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

	state.RUN_FLAG = false
	state.PACC = state.ACC
	state.ACC = 0.0
	state.DelayRAMPtr -= 1
	if state.DelayRAMPtr <= -32768 {
		state.DelayRAMPtr = 0
	}
	skipToNextSample = false

}

func GetLFOValue(lfoType int, state *State) float64 {
	lfo := 0.0
	switch lfoType {
	case 0:
		lfo = math.Sin(state.Sin0State.Angle) * float64(state.Registers[base.SIN0_RANGE].(int))
	case 1:
		lfo = math.Sin(state.Sin1State.Angle) * float64(state.Registers[base.SIN1_RANGE].(int))
	case 2:
		lfo = state.Ramp0State.Value * float64(state.Registers[base.RAMP0_RANGE].(int))
	case 3:
		lfo = state.Ramp1State.Value * float64(state.Registers[base.RAMP1_RANGE].(int))
	}
	return lfo
}

// Ensure the DelayRAM index is within bounds
func capDelayRAMIndex(in int) int {
	if in > DELAY_RAM_SIZE {
		fmt.Printf("ERROR: DelayRAM index out of bounds: %d (len=%d)\n",
			in, DELAY_RAM_SIZE)
		panic(false)
	}

	return in & 0x7fff
}
