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
	var lines []string
	var skpTargets []int

	currentLine := 0
	
	// Register all SKPs
	for i := 0; i < len(opCodes); i++ {
		op := opCodes[i]
		if op.Name == "SKP" {
			skpTargets = append(skpTargets, i+int(op.Args[1].RawValue))
		}
	}
	
	for i := 0; i < len(opCodes); i++ {
		op := opCodes[i]
		
		codeColor := "fg:white"
		numColor := "fg:gray"
		skpColor := "fg:cyan"
		if i == int(state.IP) { // Cursor line?
			codeColor = "fg:red,bg:white,mod:bold"
			numColor = "fg:black,bg:white,mod:bold"
			skpColor = "fg:cyan,bg:white,mod:bold"
			currentLine = i
		}

		if op.Name == "SKP" {
			skpTargets = append(skpTargets, i+int(op.Args[1].RawValue))
		}

		lineNoLabel := fmt.Sprintf("[ %3d](%s)", i, numColor)
		
		skipLabel := "    "
		hasSkipLabel := false
		for _, p := range skpTargets {
			if p == (i - 1) {
				skipLabel = fmt.Sprintf("\n[L%2d:](%s)", i, skpColor)				
				hasSkipLabel = true
				break
			}
		}

		label := lineNoLabel
		if hasSkipLabel {
			label = skipLabel
		}
		
		str := fmt.Sprintf("%s[ %s  ](%s)",
			label,
			disasm.OpCodeToString(op, i, false), codeColor)
		lines = append(lines, str)
	}

	threshold := uiState.terminalHeight - 8
	if currentLine > threshold {
		return strings.Join(lines[threshold:], "\n")		
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
