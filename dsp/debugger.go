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

func UpdateDebuggerScreen(opCodes []base.Op, state *State) {
	lastState = state
	lastOpCodes = opCodes

	updateCodeView(opCodes, state)
	updateStateView(state)
	updateMetaInfoView(opCodes, state)

	width, height := termui.TerminalDimensions()
	helpLine := widgets.NewParagraph()
	helpLine.Text =
		"[ESC/q/CTRL-C:](fg:black) Quit [|](bg:black) " +
			"[s/PgDn:](fg:black) Next sample [|](bg:black) " +
			"[n/Down:](fg:black) Next operation "

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
	code.Title = "  Assembly  "
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
			skpTargets = append(skpTargets, lineNo+int(op.Args[1].RawValue))
		}

		for _, p := range skpTargets {
			if p == ((lineNo + i) - 1) {
				skpLine := fmt.Sprintf("[addr_%d:](fg:cyan)", lineNo+i)
				lines = append(lines, skpLine)
			}
		}

		str := fmt.Sprintf("[%3d](fg:yellow)  [%s](%s)", lineNo+i, disasm.OpCodeToString(op, false), color)
		lines = append(lines, str)

	}
	return strings.Join(lines, "\n")
}

// Prints all values in the state
func updateStateView(state *State) {
	twidth, theight := termui.TerminalDimensions()
	theight = theight - 1 // Save one for the keys line

	accColor := "yellow"
	if state.ACC.ToFloat64() > 1.0 {
		accColor = "red"
	}

	stateStr := fmt.Sprintf("[IP:](fg:yellow,mod:bold) %d, [ACC:](fg:%s,mod:bold) %d (%f)\n"+
		"[PACC:](fg:yellow) %f, [LR:](fg:yellow) %f\n"+
		"[ADDR_PTR:](fg:yellow) %d, [DelayRAMPtr:](fg:yellow) %d, [RUN_FLAG:](fg:yellow) %t\n",
		state.IP,
		accColor, state.ACC.Value, state.ACC.ToFloat64(),
		state.PACC.ToFloat64(), state.LR.ToFloat64(),
		state.Registers[base.ADDR_PTR].Value,
		state.DelayRAMPtr,
		state.RUN_FLAG)

	ioStr := fmt.Sprintf("[ADCL:](fg:yellow) %f, [ADCR:](fg:yellow) %f\n"+
		"[DACL:](fg:green) %f, [DACR:](fg:green) %f\n",
		state.Registers[base.ADCL].ToFloat64(), state.Registers[base.ADCR].ToFloat64(),
		state.Registers[base.DACL].ToFloat64(), state.Registers[base.DACR].ToFloat64())

	potStr := fmt.Sprintf("[POT0](fg:cyan)=%f, [POT1](fg:cyan)=%f, [POT2](fg:cyan)=%f\n",
		state.Registers[base.POT0].ToFloat64(),
		state.Registers[base.POT1].ToFloat64(),
		state.Registers[base.POT2].ToFloat64())

	lfoStr := fmt.Sprintf("[SIN0](fg:yellow) [rate:](fg:cyan) %d (%f)\n     [amplitude:](fg:cyan) %d (%f)\n"+
		"[SIN1](fg:yellow) [rate:](fg:cyan) %d (%f)\n     [amplitude:](fg:cyan) %d (%f)\n",
		state.Registers[base.SIN0_RATE].Value, state.Registers[base.SIN0_RATE].ToFloat64(),
		state.Registers[base.SIN0_RANGE].Value, state.Registers[base.SIN0_RANGE].ToFloat64(),
		state.Registers[base.SIN1_RATE].Value, state.Registers[base.SIN1_RATE].ToFloat64(),
		state.Registers[base.SIN1_RANGE].Value, state.Registers[base.SIN1_RANGE].ToFloat64())
	lfoStr += fmt.Sprintf("[SIN0](fg:yellow) [Angle:](fg:cyan) %f, "+
		"[SIN1](fg:yellow) [Angle:](fg:cyan) %f\n",
		GetLFOValue(0, state, false), GetLFOValue(1, state, false))
	lfoStr += fmt.Sprintf("[RAMP0](fg:yellow) [rate:](fg:cyan) %d\n      [amplitude:](fg:cyan) %d\n"+
		"[RAMP1](fg:yellow) [rate:](fg:cyan) %d\n      [amplitude:](fg:cyan) %d\n",
		state.Registers[base.RAMP0_RATE].Value, state.Registers[base.RAMP0_RANGE].Value,
		state.Registers[base.RAMP1_RATE].Value, state.Registers[base.RAMP1_RANGE].Value)
	lfoStr += fmt.Sprintf("[RAMP0](fg:yellow) [Value:](fg:cyan) %f, "+
		"[RAMP1](fg:yellow) [Value:](fg:cyan) %f\n",
		GetLFOValue(2, state, false), GetLFOValue(3, state, false))

	vPos := 0
	stateP := widgets.NewParagraph()
	stateP.Title = "  State  "
	stateP.TitleStyle = termui.NewStyle(termui.ColorYellow, termui.ColorBlue)
	stateP.Text = stateStr + ioStr + potStr
	stateP.SetRect(twidth/2, vPos, twidth, vPos+8)
	vPos += 8

	lfoP := widgets.NewParagraph()
	lfoP.Title = "  LFOs  "
	lfoP.TitleStyle = termui.NewStyle(termui.ColorYellow, termui.ColorBlue)
	lfoP.Text = lfoStr
	lfoP.SetRect(twidth/2, vPos, twidth, vPos+12)
	vPos += 12

	regStr := ""
	for i := 0x20; i <= 0x3f; i += 2 {
		regStr += fmt.Sprintf("[Reg%2d:](fg:cyan) %f  ", i-0x20, state.Registers[i].ToFloat64())
		regStr += fmt.Sprintf("[Reg%2d:](fg:cyan) %f\n", i-0x20+1, state.Registers[i+1].ToFloat64())
	}

	regP := widgets.NewParagraph()
	regP.Title = "  Registers  "
	regP.TitleStyle = termui.NewStyle(termui.ColorYellow, termui.ColorBlue)
	regP.Text = regStr
	regP.SetRect(twidth/2, vPos, twidth, theight-5)

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
		case "q", "C-c", "<Escape>":
			return "quit"
		case "n", "<Down>":
			return "next op"
		case "s", "<PageDown>":
			return "next sample"
		case "<Resize>":
			onTerminalResized()
		}
	}

	return ""
}

func onTerminalResized() {
	termui.Clear()
	UpdateDebuggerScreen(lastOpCodes, lastState)
}
