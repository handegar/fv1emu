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
var boxTitleStyle = termui.NewStyle(termui.ColorRed, termui.ColorBlue)

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
		updateStateView(sampleNum, state)
		updateMetaInfoView(opCodes, state)
		updateRegistersView(state)

		width, height := termui.TerminalDimensions()
		helpLine := widgets.NewParagraph()
		// FIXME: Can we center this line? (20220225 handegar)
		helpLine.Text =
			"[ESC/q:](fg:black) Quit [|](fg:white,bg:black) " +
				"[F1/h/?:](fg:black) Help [|](fg:white,bg:black) " +
				"[s/PgDn:](fg:black) Next sample [|](fg:white,bg:black) " +
				"[n/Down:](fg:black) Next op "

		helpLine.Border = false
		helpLine.TextStyle = boxTitleStyle
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
				//onTerminalResized()
			} else if showMemoryMap {
				showMemoryMap = false
				//onTerminalResized()
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
			UpdateScreen(lastOpCodes, lastState, lastSampleNum)
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
		UpdateScreen(lastOpCodes, lastState, lastSampleNum)
	}
}

func decreaseMemoryCursor(count int) {
	memoryCursor -= calculateMemoryValuesPerChar() * count
	if memoryCursor < 0 {
		memoryCursor = dsp.DELAY_RAM_SIZE + memoryCursor
	}

	if showMemoryMap {
		UpdateScreen(lastOpCodes, lastState, lastSampleNum)
	}
}

func renderHelpScreen() {
	width, height := termui.TerminalDimensions()
	ypos := 0

	frame := widgets.NewParagraph()
	frame.Title = "  Help / Keys / Keywords  "
	frame.TitleStyle = boxTitleStyle
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
		" [POT0-3](fg:yellow):       Potensiometer value [0 .. 1.0]\n"

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
	code.TitleStyle = boxTitleStyle

	code.Text = generateCodeListing(opCodes, state)
	code.SetRect(0, 0, width, height)

	ui.Render(code)
}

func generateCodeListing(opCodes []base.Op, state *dsp.State) string {
	_, screenHeight := termui.TerminalDimensions()
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

func makeSineString(typ int, state *dsp.State) string {
	sinRate := int32(0)
	sinRange := int32(0)
	if typ == base.LFO_SIN0 {
		sinRate = state.Registers[base.SIN0_RATE].Value
		sinRange = state.Registers[base.SIN0_RANGE].Value
	} else {
		sinRate = state.Registers[base.SIN1_RATE].Value
		sinRange = state.Registers[base.SIN0_RANGE].Value
	}

	sinhz := (float64(sinRate) / 256.0) * 20.0
	sinrange := float64(sinRange) / 32767.0

	sin := 0.0
	cos := 0.0
	if typ == base.LFO_SIN0 {
		sin = state.Sin0Osc.GetSine()
		cos = state.Sin0Osc.GetCosine()
	} else {
		sin = state.Sin1Osc.GetSine()
		cos = state.Sin1Osc.GetCosine()
	}

	lfoStr := fmt.Sprintf(" [SIN%d](fg:yellow)  [Rate:](fg:cyan) %d ", typ, sinRate)
	lfoStr += fmt.Sprintf("[(%.2f hz)](fg:gray) ", sinhz)
	lfoStr += fmt.Sprintf("[Amp:](fg:cyan) %d [(%.2f%%)](fg:gray) ", sinRange, sinrange)
	lfoStr += fmt.Sprintf("[Value:](fg:cyan) %f \n", sin)
	lfoStr += fmt.Sprintf("       [Cos:](fg:cyan) %.3f, [CmpC:](fg:cyan) %.3f, [CmpA:](fg:cyan) %.3f\n",
		cos,
		1.0-sin,
		-sin)

	return lfoStr
}

func makeRampString(typ int, state *dsp.State) string {
	lfo := 0.0
	regVal := int32(0)
	if typ == base.LFO_RMP0 {
		regVal = state.Registers[base.RAMP0_RATE].Value
		lfo = state.Ramp0Osc.GetValue()
	} else {
		regVal = state.Registers[base.RAMP1_RATE].Value
		lfo = state.Ramp1Osc.GetValue()
	}
	rmphz := (float64(regVal) / 32767) * 20.0

	lfoStr := fmt.Sprintf(" [RAMP%d](fg:yellow) [Rate:](fg:cyan) %d ",
		typ-2, regVal)
	lfoStr += fmt.Sprintf("[(%.2f hz)](fg:gray) ", rmphz)
	lfoStr += fmt.Sprintf("[Amp:](fg:cyan) %d ", regVal)
	lfoStr += fmt.Sprintf("[Value:](fg:cyan) %f\n", lfo)

	lfoStr += fmt.Sprintf("       [NA:](fg:cyan) %.3f, [CmpC:](fg:cyan) %.3f, [CmpA:](fg:cyan) %.3f, [Half:](fg:cyan) %.2f\n",
		dsp.GetXFadeFromLFO(lfo, typ, state),
		1.0-lfo,
		dsp.GetRampRange(typ, state)-lfo,
		dsp.GetLFOValuePlusHalfCycle(typ, state))
	return lfoStr
}

// Prints all values in the state
func updateStateView(sampleNum int, state *dsp.State) {
	twidth, theight := termui.TerminalDimensions()
	theight = theight - 1 // Save one for the keys line

	rfInt := 0
	if state.RUN_FLAG {
		rfInt = 1
	}
	stateStr := fmt.Sprintf(" [ACC:](fg:yellow,mod:bold) %s (%d)"+
		" [PACC:](fg:yellow) %s, [LR:](fg:yellow) %s\n"+
		" [ADDR_PTR:](fg:yellow) %d, [DelayRAMPtr:](fg:yellow) %d, [RUN:](fg:yellow) %d\n",
		// FIXME: Is the ACC an S.23 or an S1.14? (20220305 handegar)
		overflowColored(state.ACC.ToFloat64(), -2.0, 2.0), state.ACC.Value,
		overflowColored(state.PACC.ToFloat64(), -2.0, 2.0),
		overflowColored(state.LR.ToFloat64(), -1.0, 1.0),
		int(state.Registers[base.ADDR_PTR].ToInt32())>>8,
		state.DelayRAMPtr,
		rfInt)

	ioStr := fmt.Sprintf(" [ADCL:](fg:yellow) %f, [ADCR:](fg:yellow) %f\n"+
		" [DACL:](fg:green) %s, [DACR:](fg:green) %s\n",
		state.Registers[base.ADCL].ToFloat64(), state.Registers[base.ADCR].ToFloat64(),
		overflowColored(state.Registers[base.DACL].ToFloat64(), -1.0, 1.0),
		overflowColored(state.Registers[base.DACR].ToFloat64(), -1.0, 1.0))

	potStr := fmt.Sprintf(" [POT0](fg:cyan): %.4f, [POT1](fg:cyan): %.4f, [POT2](fg:cyan): %.4f\n",
		state.Registers[base.POT0].ToFloat64(),
		state.Registers[base.POT1].ToFloat64(),
		state.Registers[base.POT2].ToFloat64())

	lfoStr := makeSineString(base.LFO_SIN0, state)
	lfoStr += makeSineString(base.LFO_SIN1, state)
	lfoStr += makeRampString(base.LFO_RMP0, state)
	lfoStr += makeRampString(base.LFO_RMP1, state)

	vPos := 0
	stateP := widgets.NewParagraph()
	stateP.Title = fmt.Sprintf("  State (sample #%d)  ", sampleNum)
	stateP.TitleStyle = boxTitleStyle
	stateP.BorderStyle = termui.NewStyle(termui.ColorGreen)
	stateP.Text = stateStr + ioStr + potStr
	stateP.SetRect(twidth/2-1, vPos, twidth, vPos+7)
	vPos += 7

	lfoP := widgets.NewParagraph()
	lfoP.Title = "  LFOs  "
	lfoP.TitleStyle = boxTitleStyle
	lfoP.BorderStyle = termui.NewStyle(termui.ColorGreen)
	lfoP.Text = lfoStr
	lfoP.SetRect(twidth/2-1, vPos, twidth, vPos+10)
	vPos += 10

	ui.Render(stateP)
	ui.Render(lfoP)

}

func updateRegistersView(state *dsp.State) {
	vPos := 17
	twidth, theight := termui.TerminalDimensions()
	theight = theight - 1 // Save one for the keys line

	valAsStr := func(registerNr int) string {
		if showRegistersAsFloats {
			return fmt.Sprintf("%s", overflowColored(state.Registers[registerNr].ToFloat64(), -1.0, 1.0))
		} else {
			return fmt.Sprintf("%d", state.Registers[registerNr].ToInt32())
		}
	}

	regStr := ""
	// Special registers
	regStr += fmt.Sprintf(" [SIN0_RATE:](fg:cyan): %s  ", valAsStr(0))
	regStr += fmt.Sprintf(" [SIN0_RANGE:](fg:cyan): %s\n", valAsStr(1))
	regStr += fmt.Sprintf(" [SIN1_RATE:](fg:cyan): %s  ", valAsStr(2))
	regStr += fmt.Sprintf(" [SIN1_RANGE:](fg:cyan): %s\n", valAsStr(3))
	regStr += fmt.Sprintf(" [RMP0_RATE:](fg:cyan): %s  ", valAsStr(4))
	regStr += fmt.Sprintf(" [RMP0_RANGE:](fg:cyan): %s\n", valAsStr(5))
	regStr += fmt.Sprintf(" [RMP1_RATE:](fg:cyan): %s  ", valAsStr(6))
	regStr += fmt.Sprintf(" [RMP1_RANGE:](fg:cyan): %s\n", valAsStr(7))

	regStr += fmt.Sprintf(" [POT0:](fg:cyan): %s  ", valAsStr(16))
	regStr += fmt.Sprintf(" [POT1:](fg:cyan): %s  ", valAsStr(17))
	regStr += fmt.Sprintf(" [POT2:](fg:cyan): %s\n", valAsStr(18))

	// General registers
	for i := 0x20; i <= 0x3f; i += 2 {
		if showRegistersAsFloats {
			regStr += fmt.Sprintf(" [Reg%3d:](fg:cyan) %s  ",
				i-0x20, overflowColored(state.Registers[i].ToFloat64(), -1.0, 1.0))
			regStr += fmt.Sprintf(" [Reg%3d:](fg:cyan) %s\n",
				i-0x20+1, overflowColored(state.Registers[i+1].ToFloat64(), -1.0, 1.0))
		} else {
			regStr += fmt.Sprintf(" [Reg%3d:](fg:cyan) %d  ",
				i-0x20, state.Registers[i].ToInt32())
			regStr += fmt.Sprintf(" [Reg%3d:](fg:cyan) %d\n",
				i-0x20+1, state.Registers[i+1].ToInt32())
		}
	}

	regP := widgets.NewParagraph()
	regP.Title = "  Registers"
	if showRegistersAsFloats {
		regP.Title += " as floats."
	} else {
		regP.Title += " as integers."
	}
	regP.Title += " Toggle type with 'f'  "
	regP.TitleStyle = boxTitleStyle
	regP.BorderStyle = termui.NewStyle(termui.ColorGreen)
	regP.Text = regStr
	regP.SetRect(twidth/2-1, vPos, twidth, theight-5)

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
	infoP.TitleStyle = boxTitleStyle
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
