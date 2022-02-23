package dsp

import (
	"fmt"
	"strings"

	termui "github.com/gizak/termui/v3"
	ui "github.com/gizak/termui/v3"
	widgets "github.com/gizak/termui/v3/widgets"

	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/disasm"
)

var lastState *State
var lastOpCodes []base.Op
var lastSampleNum int
var showRegistersAsFloats bool = true

func UpdateDebuggerScreen(opCodes []base.Op, state *State, sampleNum int) {
	lastState = state
	lastOpCodes = opCodes
	lastSampleNum = sampleNum

	updateCodeView(opCodes, state)
	updateStateView(state, sampleNum)
	updateMetaInfoView(opCodes, state)

	width, height := termui.TerminalDimensions()
	helpLine := widgets.NewParagraph()
	helpLine.Text =
		"[ESC/q:](fg:black) Quit [|](bg:black) " +
			"[s/PgDn:](fg:black) +1 sample [|](bg:black) " +
			"[CTRL-s:](fg:black) +100 samples [|](bg:black) " +
			"[n/Down:](fg:black) Next op [|](bg:black) " +
			"[f:](fg:black) Floats/Ints "

	helpLine.Border = false
	helpLine.TextStyle.Fg = termui.ColorWhite
	helpLine.TextStyle.Bg = termui.ColorBlue
	helpLine.SetRect(0, height-1, width, height)
	ui.Render(helpLine)
}

// Prints the code with a highlighted current-op
func updateCodeView(opCodes []base.Op, state *State) {
	width, height := termui.TerminalDimensions()

	width = width / 2
	height = height - 5 - 1

	code := widgets.NewParagraph()
	code.Title = fmt.Sprintf("  Instructions (%d) ", len(opCodes))
	code.TitleStyle = termui.NewStyle(termui.ColorYellow, termui.ColorBlue)

	code.Text = generateCodeListing(opCodes, state, height)
	code.SetRect(0, 0, width, height)

	ui.Render(code)
}

func generateCodeListing(opCodes []base.Op, state *State, screenHeight int) string {
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
		color := "fg:white"
		if (lineNo + i) == int(state.IP) {
			color = "fg:red,bg:white,mod:bold"
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

		str := fmt.Sprintf("[%3d](fg:yellow)  [%s](%s)",
			lineNo+i+1,
			disasm.OpCodeToString(op, lineNo+i, false), color)
		lines = append(lines, str)

	}
	return strings.Join(lines, "\n")
}

func overflowColored(v float64, min float64, max float64) string {
	overflowed := false
	if v > max || v < min {
		overflowed = true
	}
	color := ""
	if overflowed {
		color = "fg:black,bg:red"
	}
	return fmt.Sprintf("[%f](%s)", v, color)
}

// Prints all values in the state
func updateStateView(state *State, sampleNum int) {
	twidth, theight := termui.TerminalDimensions()
	theight = theight - 1 // Save one for the keys line

	stateStr := fmt.Sprintf("[IP:](fg:yellow,mod:bold) %d, [ACC:](fg:yellow,mod:bold) %d (%s)\n"+
		"[PACC:](fg:yellow) %s, [LR:](fg:yellow) %s\n"+
		"[ADDR_PTR:](fg:yellow) %d, [DelayRAMPtr:](fg:yellow) %d, [RUN_FLAG:](fg:yellow) %t\n",
		state.IP,
		state.ACC.Value, overflowColored(state.ACC.ToFloat64(), -1.0, 1.0),
		overflowColored(state.PACC.ToFloat64(), -1.0, 1.0),
		overflowColored(state.LR.ToFloat64(), -1.0, 1.0),
		state.Registers[base.ADDR_PTR].Value,
		state.DelayRAMPtr,
		state.RUN_FLAG)

	ioStr := fmt.Sprintf("[ADCL:](fg:yellow) %f, [ADCR:](fg:yellow) %f\n"+
		"[DACL:](fg:green) %s, [DACR:](fg:green) %s\n",
		state.Registers[base.ADCL].ToFloat64(), state.Registers[base.ADCR].ToFloat64(),
		overflowColored(state.Registers[base.DACL].ToFloat64(), -1.0, 1.0),
		overflowColored(state.Registers[base.DACR].ToFloat64(), -1.0, 1.0))

	potStr := fmt.Sprintf("[POT0](fg:cyan)=%f, [POT1](fg:cyan)=%f, [POT2](fg:cyan)=%f\n",
		state.Registers[base.POT0].ToFloat64(),
		state.Registers[base.POT1].ToFloat64(),
		state.Registers[base.POT2].ToFloat64())

	lfoStr := fmt.Sprintf("[SIN0](fg:yellow) [Rate:](fg:cyan) %d (%f)\n     [Range:](fg:cyan) %d (%f)\n"+
		"[SIN1](fg:yellow) [Rate:](fg:cyan) %d (%f)\n     [Range:](fg:cyan) %d (%f)\n",
		state.Registers[base.SIN0_RATE].Value, state.Registers[base.SIN0_RATE].ToFloat64(),
		state.Registers[base.SIN0_RANGE].Value, state.Registers[base.SIN0_RANGE].ToFloat64(),
		state.Registers[base.SIN1_RATE].Value, state.Registers[base.SIN1_RATE].ToFloat64(),
		state.Registers[base.SIN1_RANGE].Value, state.Registers[base.SIN1_RANGE].ToFloat64())
	lfoStr += fmt.Sprintf("[SIN0](fg:yellow) [\u03B1:](fg:cyan) %f, "+
		"[SIN1](fg:yellow) [\u03B1:](fg:cyan) %f\n",
		GetLFOValue(0, state, false), GetLFOValue(1, state, false))
	lfoStr += fmt.Sprintf("[RAMP0](fg:yellow) [Rate:](fg:cyan) %d  [Range:](fg:cyan) %d\n"+
		"[RAMP1](fg:yellow) [Rate:](fg:cyan) %d  [Range:](fg:cyan) %d\n",
		state.Registers[base.RAMP0_RATE].Value, state.Registers[base.RAMP0_RANGE].Value,
		state.Registers[base.RAMP1_RATE].Value, state.Registers[base.RAMP1_RANGE].Value)
	lfoStr += fmt.Sprintf("[RAMP0](fg:yellow) [\u0394:](fg:cyan) %f, "+
		"[RAMP1](fg:yellow) [\u0394:](fg:cyan) %f\n",
		GetLFOValue(2, state, false), GetLFOValue(3, state, false))

	vPos := 0
	stateP := widgets.NewParagraph()
	stateP.Title = fmt.Sprintf("  State (sample #%d)  ", sampleNum)
	stateP.TitleStyle = termui.NewStyle(termui.ColorYellow, termui.ColorBlue)
	stateP.Text = stateStr + ioStr + potStr
	stateP.SetRect(twidth/2-1, vPos, twidth, vPos+8)
	vPos += 8

	lfoP := widgets.NewParagraph()
	lfoP.Title = "  LFOs  "
	lfoP.TitleStyle = termui.NewStyle(termui.ColorYellow, termui.ColorBlue)
	lfoP.Text = lfoStr
	lfoP.SetRect(twidth/2-1, vPos, twidth, vPos+10)
	vPos += 10

	regStr := ""
	for i := 0x20; i <= 0x3f; i += 2 {
		if showRegistersAsFloats {
			regStr += fmt.Sprintf("[Reg%2d:](fg:cyan) %f  ",
				i-0x20, state.Registers[i].ToFloat64())
			regStr += fmt.Sprintf("[Reg%2d:](fg:cyan) %f\n",
				i-0x20+1, state.Registers[i+1].ToFloat64())
		} else {
			regStr += fmt.Sprintf("[Reg%2d:](fg:cyan) %d  ",
				i-0x20, state.Registers[i].ToInt32())
			regStr += fmt.Sprintf("[Reg%2d:](fg:cyan) %d\n",
				i-0x20+1, state.Registers[i+1].ToInt32())
		}
	}

	regP := widgets.NewParagraph()
	regP.Title = "  Registers"
	if showRegistersAsFloats {
		regP.Title += " (as floats)  "
	} else {
		regP.Title += " (as integers)  "
	}
	regP.TitleStyle = termui.NewStyle(termui.ColorYellow, termui.ColorBlue)
	regP.Text = regStr
	regP.SetRect(twidth/2-1, vPos, twidth, theight-5)

	ui.Render(stateP)
	ui.Render(lfoP)
	ui.Render(regP)
}

// Prints misc info regarding current state and op
func updateMetaInfoView(opCodes []base.Op, state *State) {
	op := opCodes[state.IP]
	opDoc := disasm.OpDocs[op.Name]

	twidth, theight := termui.TerminalDimensions()
	theight = theight - 1 // Make one line free at the bottom for keys

	infoStr := fmt.Sprintf("[%s](fg:red): [%s](fg:yellow) (%s)\n[%s](fg:cyan)",
		op.Name, opDoc.Short, opDoc.Formulae, opDoc.Long)

	infoP := widgets.NewParagraph()
	infoP.Title = "  Info  "
	infoP.TitleStyle = termui.NewStyle(termui.ColorYellow, termui.ColorBlue)
	infoP.Text = infoStr
	infoP.SetRect(0, theight-5, twidth, theight)
	ui.Render(infoP)
}

/*
   Returns the Event.ID string for events which is relevant for others
   (quit, restart etc.)
*/
func WaitForDebuggerInput(state *State) string {
	for e := range ui.PollEvents() {
		switch e.ID {
		case "q", "<C-c>", "<Escape>":
			return "quit"
		case "n", "<Down>":
			return "next op"
		case "s", "<PageDown>":
			return "next sample"
		case "<C-s>":
			return "next 100 samples"
		case "f":
			showRegistersAsFloats = !showRegistersAsFloats
			UpdateDebuggerScreen(lastOpCodes, lastState, lastSampleNum)
		case "<Resize>":
			onTerminalResized()
		}
	}

	return ""
}

func onTerminalResized() {
	termui.Clear()
	UpdateDebuggerScreen(lastOpCodes, lastState, lastSampleNum)
}
