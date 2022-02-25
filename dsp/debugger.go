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
var showHelpScreen bool = false

func UpdateDebuggerScreen(opCodes []base.Op, state *State, sampleNum int) {
	lastState = state
	lastOpCodes = opCodes
	lastSampleNum = sampleNum

	if showHelpScreen {
		renderHelpScreen()
	} else {
		updateCodeView(opCodes, state)
		updateStateView(state, sampleNum)
		updateMetaInfoView(opCodes, state)

		width, height := termui.TerminalDimensions()
		helpLine := widgets.NewParagraph()
		// FIXME: Can we center this line? (20220225 handegar)
		helpLine.Text =
			"[ESC/q:](fg:black) Quit [|](bg:black) " +
				"[F1/h:](fg:black) Help [|](bg:black) " +
				"[s/PgDn:](fg:black) Next sample [|](bg:black) " +
				"[n/Down:](fg:black) Next op "

		helpLine.Border = false
		helpLine.TextStyle = termui.NewStyle(termui.ColorWhite, termui.ColorBlue)
		helpLine.SetRect(0, height-1, width, height)
		ui.Render(helpLine)
	}
}

func renderHelpScreen() {
	width, height := termui.TerminalDimensions()
	help := widgets.NewParagraph()
	help.Title = "  Help / Keys  "
	help.TitleStyle = termui.NewStyle(termui.ColorYellow, termui.ColorBlue)

	keys := widgets.NewList()
	keys.Border = false
	keys.TextStyle = termui.NewStyle(termui.ColorYellow)
	keys.SetRect(1, 1, width-1, height-1)
	keys.SelectedRowStyle = termui.NewStyle(termui.ColorCyan)

	keys.Rows = append(keys.Rows, "Keys")
	keys.Rows = append(keys.Rows, " ESC, q, CTRL-C:    [Quit debugger / exit help](fg:white)")
	keys.Rows = append(keys.Rows, " s:                 [Next sample](fg:white)")
	keys.Rows = append(keys.Rows, " SHIFT-s:           [Skip 100 samples](fg:white)")
	keys.Rows = append(keys.Rows, " CTRL-s:            [Skip 1000 samples](fg:white)")
	keys.Rows = append(keys.Rows, " n, DownKey, PgDn:  [Next instruction](fg:white)")
	keys.Rows = append(keys.Rows, " f:                 [Display register values as float or integers](fg:white)")

	help.SetRect(0, 0, width, height)
	ui.Render(help)
	ui.Render(keys)
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

	potStr := fmt.Sprintf("[POT0](fg:cyan)=%.4f, [POT1](fg:cyan)=%.4f, [POT2](fg:cyan)=%.4f\n",
		state.Registers[base.POT0].ToFloat64(),
		state.Registers[base.POT1].ToFloat64(),
		state.Registers[base.POT2].ToFloat64())

	lfoStr := fmt.Sprintf("[SIN0](fg:yellow)  [Rate:](fg:cyan) %d (%f)  [\u03B10:](fg:cyan) %f\n"+
		"      [Range:](fg:cyan) %d (%f)\n"+
		"[SIN1](fg:yellow)  [Rate:](fg:cyan) %d (%f)  [\u03B11:](fg:cyan) %f\n"+
		"      [Range:](fg:cyan) %d (%f)\n",
		state.Registers[base.SIN0_RATE].Value, state.Registers[base.SIN0_RATE].ToFloat64(),
		GetLFOValue(0, state, false),
		state.Registers[base.SIN0_RANGE].Value, state.Registers[base.SIN0_RANGE].ToFloat64(),
		state.Registers[base.SIN1_RATE].Value, state.Registers[base.SIN1_RATE].ToFloat64(),
		GetLFOValue(1, state, false),
		state.Registers[base.SIN1_RANGE].Value, state.Registers[base.SIN1_RANGE].ToFloat64())
	lfoStr += fmt.Sprintf("[RAMP0](fg:yellow) [Rate:](fg:cyan) %d (%f)  [\u03940:](fg:cyan) %f\n"+
		"      [Range:](fg:cyan) %d (%f)\n"+
		"[RAMP1](fg:yellow) [Rate:](fg:cyan) %d (%f)  [\u03941:](fg:cyan) %f\n"+
		"      [Range:](fg:cyan) %d (%f)",
		state.Registers[base.RAMP0_RATE].Value, state.Registers[base.RAMP0_RATE].ToFloat64(),
		GetLFOValue(2, state, false),
		state.Registers[base.RAMP0_RANGE].Value, state.Registers[base.RAMP0_RANGE].ToFloat64(),
		state.Registers[base.RAMP1_RATE].Value, state.Registers[base.RAMP1_RATE].ToFloat64(),
		GetLFOValue(3, state, false),
		state.Registers[base.RAMP1_RANGE].Value, state.Registers[base.RAMP1_RANGE].ToFloat64())

	vPos := 0
	stateP := widgets.NewParagraph()
	stateP.Title = fmt.Sprintf("  State (sample #%d)  ", sampleNum)
	stateP.TitleStyle = termui.NewStyle(termui.ColorYellow, termui.ColorBlue)
	stateP.BorderStyle = termui.NewStyle(termui.ColorGreen)
	stateP.Text = stateStr + ioStr + potStr
	stateP.SetRect(twidth/2-1, vPos, twidth, vPos+8)
	vPos += 8

	lfoP := widgets.NewParagraph()
	lfoP.Title = "  LFOs  "
	lfoP.TitleStyle = termui.NewStyle(termui.ColorYellow, termui.ColorBlue)
	lfoP.BorderStyle = termui.NewStyle(termui.ColorGreen)
	lfoP.Text = lfoStr
	lfoP.SetRect(twidth/2-1, vPos, twidth, vPos+10)
	vPos += 10

	regStr := ""
	for i := 0x20; i <= 0x3f; i += 2 {
		if showRegistersAsFloats {
			regStr += fmt.Sprintf("[Reg%2d:](fg:cyan) %s  ",
				i-0x20, overflowColored(state.Registers[i].ToFloat64(), 0, 1.0))
			regStr += fmt.Sprintf("[Reg%2d:](fg:cyan) %s\n",
				i-0x20+1, overflowColored(state.Registers[i+1].ToFloat64(), 0, 1.0))
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
	regP.BorderStyle = termui.NewStyle(termui.ColorGreen)
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
			if showHelpScreen {
				showHelpScreen = false
				onTerminalResized()
			} else {
				return "quit"
			}
		case "n", "<Down>":
			return "next op"
		case "s", "<PageDown>":
			return "next sample"
		case "S":
			return "next 100 samples"
		case "<C-s>":
			return "next 1000 samples"
		case "h", "<F1>":
			showHelpScreen = !showHelpScreen
			UpdateDebuggerScreen(lastOpCodes, lastState, lastSampleNum)
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
