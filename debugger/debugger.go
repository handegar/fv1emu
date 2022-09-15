package debugger

import (
	"fmt"
	"strings"

	termui "github.com/gizak/termui/v3"
	ui "github.com/gizak/termui/v3"
	widgets "github.com/gizak/termui/v3/widgets"

	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/disasm"
	"github.com/handegar/fv1emu/dsp"
	"github.com/handegar/fv1emu/settings"
)

var lastState *dsp.State
var lastOpCodes []base.Op
var lastSampleNum int

var showRegistersAsFloats bool = true
var showHelpScreen bool = false
var showMemoryMap bool = false

var previousStates map[uint]*dsp.State = make(map[uint]*dsp.State)

var memoryCursor int = 0

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

func UpdateScreen(opCodes []base.Op, state *dsp.State, sampleNum int) {
	lastState = state
	lastOpCodes = opCodes
	lastSampleNum = sampleNum

	if showHelpScreen {
		renderHelpScreen()
	} else if showMemoryMap {
		renderMemoryMap(state, sampleNum, memoryCursor)
	} else {
		updateCodeView(opCodes, state)
		updateStateView(state, sampleNum)
		updateMetaInfoView(opCodes, state)

		width, height := termui.TerminalDimensions()
		helpLine := widgets.NewParagraph()
		// FIXME: Can we center this line? (20220225 handegar)
		helpLine.Text =
			"[ESC/q:](fg:black) Quit [|](bg:black) " +
				"[F1/h/?:](fg:black) Help [|](bg:black) " +
				"[s/PgDn:](fg:black) Next sample [|](bg:black) " +
				"[n/Down:](fg:black) Next op "

		helpLine.Border = false
		helpLine.TextStyle = termui.NewStyle(termui.ColorWhite, termui.ColorBlue)
		helpLine.SetRect(0, height-1, width, height)
		ui.Render(helpLine)
	}
}

/*
   Returns the Event.ID string for events which is relevant for others
   (quit, restart etc.)
*/
func WaitForInput(state *dsp.State) string {
	for e := range ui.PollEvents() {
		switch e.ID {
		case "q", "<C-c>", "<Escape>":
			if showHelpScreen {
				showHelpScreen = false
				onTerminalResized()
			} else if showMemoryMap {
				showMemoryMap = false
				onTerminalResized()
			} else {
				return "quit"
			}
		case "n", "<Down>":
			return "next op"
		case "p", "<Up>":
			if state.IP > 0 {
				return "previous op"
			}
			return WaitForInput(state) // Keep waiting
		case "6":
			increaseMemoryCursor(1)
		case "4":
			decreaseMemoryCursor(1)
		case "2":
			increaseMemoryCursor(128)
		case "8":
			decreaseMemoryCursor(128)
		case "s", "<PageDown>":
			return "next sample"
		case "S":
			return "next 100 samples"
		case "<C-s>":
			return "next 1000 samples"
		case "g":
			return "next 10000 samples"
		case "G":
			return "next 100000 samples"
		case "h", "<F1>", "?":
			showHelpScreen = !showHelpScreen
			UpdateScreen(lastOpCodes, lastState, lastSampleNum)
		case "m", "<F2>":
			showMemoryMap = !showMemoryMap
			UpdateScreen(lastOpCodes, lastState, lastSampleNum)
		case "f":
			showRegistersAsFloats = !showRegistersAsFloats
			UpdateScreen(lastOpCodes, lastState, lastSampleNum)
		case "<Resize>":
			onTerminalResized()
		}
	}

	return ""
}

func increaseMemoryCursor(count int) {
	memoryCursor += calculateMemoryValuesPerChar() * count
	if memoryCursor >= dsp.DELAY_RAM_SIZE {
		memoryCursor = memoryCursor - dsp.DELAY_RAM_SIZE
	}

	if showMemoryMap {
		onTerminalResized()
	}
}

func decreaseMemoryCursor(count int) {
	memoryCursor -= calculateMemoryValuesPerChar() * count
	if memoryCursor < 0 {
		memoryCursor = dsp.DELAY_RAM_SIZE + memoryCursor
	}

	if showMemoryMap {
		onTerminalResized()
	}
}

func renderHelpScreen() {
	width, height := termui.TerminalDimensions()
	ypos := 0

	frame := widgets.NewParagraph()
	frame.Title = "  Help / Keys / Keywords  "
	frame.TitleStyle = termui.NewStyle(termui.ColorYellow, termui.ColorBlue)
	frame.SetRect(0, 0, width, height)
	ypos += 1

	keys := widgets.NewList()
	keys.Border = false
	keys.TextStyle = termui.NewStyle(termui.ColorYellow)
	keys.SetRect(1, 1, width-1, height-1)
	keys.SelectedRowStyle = termui.NewStyle(termui.ColorCyan)

	keys.Rows = append(keys.Rows, "Keys:")
	keys.Rows = append(keys.Rows, " h, F1, ?:          [This help-page](fg:white)")
	keys.Rows = append(keys.Rows, " ESC, q, CTRL-C:    [Quit debugger / exit help](fg:white)")
	keys.Rows = append(keys.Rows, " m, F2:             [Show delay memory map](fg:white)")
	keys.Rows = append(keys.Rows, " 6 (Keypad right):  [Memory map: Next position](fg:white)")
	keys.Rows = append(keys.Rows, " 4 (Keypad left):   [Memory map: Prev position](fg:white)")
	keys.Rows = append(keys.Rows, " 8 (Keypad up):     [Memory map: Back 128 positions](fg:white)")
	keys.Rows = append(keys.Rows, " 2 (Keypad down):   [Memory map: Skip 128 positions](fg:white)")
	keys.Rows = append(keys.Rows, " s, PgDn:           [Next sample](fg:white)")
	keys.Rows = append(keys.Rows, " SHIFT-s:           [Skip 100 samples](fg:white)")
	keys.Rows = append(keys.Rows, " CTRL-s:            [Skip 1000 samples](fg:white)")
	keys.Rows = append(keys.Rows, " g:                 [Skip 10.000 samples](fg:white)")
	keys.Rows = append(keys.Rows, " SHIFT-g:           [Skip 100.000 samples](fg:white)")
	keys.Rows = append(keys.Rows, " n, DownKey:        [Next instruction](fg:white)")
	keys.Rows = append(keys.Rows, " p, UpKey:          [Previous instruction (within current sample)](fg:white)")
	keys.Rows = append(keys.Rows, " f:                 [Display register values as float or integers](fg:white)")

	keys.SetRect(1, ypos, width-1, ypos+len(keys.Rows)+2)
	ypos += len(keys.Rows) + 1

	help := widgets.NewParagraph()
	help.Border = false
	help.Text = "[Keywords:](fg:cyan)\n" +
		" [IP](fg:yellow):           Instruction pointer.\n" +
		" [ACC](fg:yellow):          The Accumulator.\n" +
		" [PACC](fg:yellow):         Accumulator from the previous sample/state.\n" +
		" [LR](fg:yellow):           Last read sample read from the delay memory.\n" +
		" [ADDR_PTR](fg:yellow):     Special memory-pointer register.\n" +
		" [DelayRAMPtr](fg:yellow):  Decreasing memory pointer. Decreases by one each sample.\n" +
		"               Restarts as 32768 when reaching 0.\n" +
		" [RF](fg:yellow):           Run flag. False during the first sample, then True\n" +
		" [ADCL](fg:yellow):         Input value (left)\n" +
		" [ADCR](fg:yellow):         Input value (right)\n" +
		" [DACL](fg:yellow):         Output value (left)\n" +
		" [DACR](fg:yellow):         Output value (right)\n" +
		" [POT0-3](fg:yellow):       Potentiometer value [0 .. 1.0]\n"

	help.SetRect(1, ypos, width-1, ypos+(height-ypos)-12)
	ypos += 12

	ui.Render(frame)
	ui.Render(keys)
	ui.Render(help)
}

// Prints the code with a highlighted current-op
func updateCodeView(opCodes []base.Op, state *dsp.State) {
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

func generateCodeListing(opCodes []base.Op, state *dsp.State, screenHeight int) string {
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
	} else {
		if v > 0.0 {
			color = "bg:black"
		}
	}
	return fmt.Sprintf("[%f](%s)", v, color)
}

// Prints all values in the state
func updateStateView(state *dsp.State, sampleNum int) {
	twidth, theight := termui.TerminalDimensions()
	theight = theight - 1 // Save one for the keys line

	rfInt := 0
	if state.RUN_FLAG {
		rfInt = 1
	}
	stateStr := fmt.Sprintf("[IP:](fg:yellow,mod:bold) %d, [ACC:](fg:yellow,mod:bold) %s (%d)\n"+
		"[PACC:](fg:yellow) %s, [LR:](fg:yellow) %s\n"+
		"[ADDR_PTR:](fg:yellow) %d, [DelayRAMPtr:](fg:yellow) %d, [RF:](fg:yellow) %d\n",
		state.IP,
		// FIXME: Is the ACC an S.23 or an S1.14? (20220305 handegar)
		overflowColored(state.ACC.ToFloat64(), -2.0, 2.0), state.ACC.Value,
		overflowColored(state.PACC.ToFloat64(), -2.0, 2.0),
		overflowColored(state.LR.ToFloat64(), -1.0, 1.0),
		int(state.Registers[base.ADDR_PTR].ToInt32())>>8,
		state.DelayRAMPtr,
		rfInt)

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
		dsp.GetLFOValue(0, state, false),
		state.Registers[base.SIN0_RANGE].Value, state.Registers[base.SIN0_RANGE].ToFloat64(),
		state.Registers[base.SIN1_RATE].Value, state.Registers[base.SIN1_RATE].ToFloat64(),
		dsp.GetLFOValue(1, state, false),
		state.Registers[base.SIN1_RANGE].Value, state.Registers[base.SIN1_RANGE].ToFloat64())
	lfoStr += fmt.Sprintf("[RAMP0](fg:yellow) [Rate:](fg:cyan) %d (%f)  [Value:](fg:cyan) %f\n"+
		"      [Range:](fg:cyan) %d (%f)\n"+
		"[RAMP1](fg:yellow) [Rate:](fg:cyan) %d (%f)  [Value:](fg:cyan) %f\n"+
		"      [Range:](fg:cyan) %d (%f)",
		state.Registers[base.RAMP0_RATE].Value, state.Registers[base.RAMP0_RATE].ToFloat64(),
		dsp.GetLFOValue(2, state, false),
		base.RampAmpValues[state.Registers[base.RAMP0_RANGE].Value], state.Registers[base.RAMP0_RANGE].ToFloat64(),
		state.Registers[base.RAMP1_RATE].Value, state.Registers[base.RAMP1_RATE].ToFloat64(),
		dsp.GetLFOValue(3, state, false),
		base.RampAmpValues[state.Registers[base.RAMP1_RANGE].Value], state.Registers[base.RAMP1_RANGE].ToFloat64())

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
			regStr += fmt.Sprintf("[Reg%3d:](fg:cyan) %s  ",
				i-0x20, overflowColored(state.Registers[i].ToFloat64(), -1.0, 1.0))
			regStr += fmt.Sprintf("[Reg%3d:](fg:cyan) %s\n",
				i-0x20+1, overflowColored(state.Registers[i+1].ToFloat64(), -1.0, 1.0))
		} else {
			regStr += fmt.Sprintf("[Reg%3d:](fg:cyan) %d  ",
				i-0x20, state.Registers[i].ToInt32())
			regStr += fmt.Sprintf("[Reg%3d:](fg:cyan) %d\n",
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
func updateMetaInfoView(opCodes []base.Op, state *dsp.State) {
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

	versionP := widgets.NewParagraph()
	versionP.Border = false
	versionP.PaddingBottom = 0
	versionP.PaddingTop = 0
	versionP.PaddingLeft = 0
	versionP.PaddingRight = 0
	versionP.Text = fmt.Sprintf("[v%s](fg:blue)", settings.Version)
	versionP.SetRect(twidth-len(settings.Version)-6, theight-1,
		twidth-3, theight)

	ui.Render(infoP)
	ui.Render(versionP)
}

func onTerminalResized() {
	termui.Clear()
	UpdateScreen(lastOpCodes, lastState, lastSampleNum)
}
