package debugger

import (
	"fmt"
	"strings"

	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/disasm"
	"github.com/handegar/fv1emu/dsp"
)

var lastState *dsp.State
var lastOpCodes []base.Op
var lastSampleNum int

var previousStates map[uint]*dsp.State = make(map[uint]*dsp.State)

func Reset() {
	previousStates = make(map[uint]*dsp.State)
}

func RegisterState(state *dsp.State) {
	if previousStates[state.IP] == nil {
		previousStates[state.IP] = state.Duplicate()
	}
}

func GetRegisteredState(ip uint) (*dsp.State, bool) {
	prevState, found := previousStates[ip]
	return prevState, found
}

func generateCodeListing(opCodes []base.Op, state *dsp.State) string {
	screenHeight := uiState.terminalHeight
	var lines []string
	var skpTargets []int

	lineNo := 0
	if int(state.IP) > (screenHeight / 2) {
		lineNo = int(state.IP) - (screenHeight / 2)
	}

	for i := 0; i < screenHeight; i++ {
		if (lineNo + i) > (len(opCodes) - 1) {
			break
		}

		op := opCodes[lineNo+i]
		codeColor := "fg:white"
		numColor := "fg:yellow"
		if (lineNo + i) == int(state.IP) { // Cursor line?
			codeColor = "fg:red,bg:white,mod:bold"
			numColor = "fg:black,bg:white,mod:bold"
		}

		if op.Name == "SKP" {
			skpTargets = append(skpTargets, lineNo+i+int(op.Args[1].RawValue))
		}

		for _, p := range skpTargets {
			if p == ((lineNo + i) - 1) {
				skpLine := fmt.Sprintf("[addr_%d:](fg:cyan)", lineNo+i)
				lines = append(lines, skpLine)
				break
			}
		}

		str := fmt.Sprintf("[%3d](%s)[  %s  ](%s)",
			lineNo+i, numColor,
			disasm.OpCodeToString(op, lineNo, false), codeColor)
		lines = append(lines, str)

	}
	return strings.Join(lines, "\n")
}

// Make string RED if above of below a specified threshold
func overflowColored(v float64, min float64, max float64) string {
	overflowed := false
	if v > max || v < min {
		overflowed = true
	}
	color := ""
	if overflowed {
		color = "fg:black,bg:red"
	} else {
		if v > 0.0 {
			color = "bg:black"
		}
	}
	return fmt.Sprintf("[%f](%s)", v, color)
}
