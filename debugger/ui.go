package debugger

import (
	"fmt"

	termui "github.com/gizak/termui/v3"
	ui "github.com/gizak/termui/v3"
	widgets "github.com/gizak/termui/v3/widgets"

	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/disasm"
	"github.com/handegar/fv1emu/dsp"
	"github.com/handegar/fv1emu/settings"
	"github.com/handegar/fv1emu/utils"
)

const (
	MainScreen int = iota
	MemoryScreen
	HelpScreen
)

type UIState struct {
	terminalWidth  int
	terminalHeight int
	centerLine     int

	showRegistersAsFloats bool
	currentScreen         int
	memoryCursor          int

	codeView        *widgets.Paragraph
	metaInfoView    *widgets.Paragraph
	mainStateView   *widgets.Paragraph
	lfoStateView    *widgets.Paragraph
	infoLineView    *widgets.Paragraph
	versionLineView *widgets.Paragraph
	registerView    *widgets.Paragraph
	helpLineView    *widgets.Paragraph
}

var uiState UIState

var boxTitleStyle = termui.NewStyle(termui.ColorRed, termui.ColorBlue)

func Init() {
	width, height := termui.TerminalDimensions()
	uiState.terminalHeight = height
	uiState.terminalWidth = width
	uiState.centerLine = max(width/2, 53)
}

/*
Returns the Event.ID string for events which is relevant for others
(quit, restart etc.)
*/
func WaitForInput(state *dsp.State) string {
	for e := range ui.PollEvents() {
		switch e.ID {
		case "q", "<C-c>", "<Escape>":
			return "quit"
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
			if uiState.currentScreen == HelpScreen {
				uiState.currentScreen = MainScreen
			} else {
				uiState.currentScreen = HelpScreen
			}
			UpdateScreen(lastOpCodes, lastState, lastSampleNum)
		case "m", "<F2>":
			if uiState.currentScreen == MemoryScreen {
				uiState.currentScreen = MainScreen
			} else {
				uiState.currentScreen = MemoryScreen
			}
			UpdateScreen(lastOpCodes, lastState, lastSampleNum)
		case "f":
			uiState.showRegistersAsFloats = !uiState.showRegistersAsFloats
			UpdateScreen(lastOpCodes, lastState, lastSampleNum)
		case "<Resize>":
			width, height := termui.TerminalDimensions()
			uiState.terminalHeight = height
			uiState.terminalWidth = width
			uiState.centerLine = max(width/2, 53)

			UpdateScreen(lastOpCodes, lastState, lastSampleNum)
		}
	}

	return ""
}

func UpdateScreen(opCodes []base.Op, state *dsp.State, sampleNum int) {
	lastState = state
	lastOpCodes = opCodes
	lastSampleNum = sampleNum

	switch uiState.currentScreen {
	case HelpScreen:
		renderHelpScreen()
	case MemoryScreen:
		renderMemoryMap(state, sampleNum, uiState.memoryCursor)
	case MainScreen:
		renderMainScreen(opCodes, state, sampleNum)
	default:
		utils.Assert(false, "Unknown ui-screen: %d", uiState.currentScreen)
	}
}

func renderMainScreen(opCodes []base.Op, state *dsp.State, sampleNum int) {
	updateCodeView(opCodes, state)
	updateStateView(sampleNum, state)
	updateMetaInfoView(opCodes, state)
	updateRegistersView(state)
	updateHelpLineView()

	renderUiState()
}

func renderUiState() {
	ui.Render(uiState.codeView, uiState.helpLineView, uiState.infoLineView,
		uiState.mainStateView, uiState.lfoStateView, uiState.registerView,
		uiState.helpLineView)
}

func increaseMemoryCursor(count int) {
	uiState.memoryCursor += count
	if uiState.memoryCursor >= dsp.DELAY_RAM_SIZE {
		uiState.memoryCursor = uiState.memoryCursor - dsp.DELAY_RAM_SIZE
	}

	UpdateScreen(lastOpCodes, lastState, lastSampleNum)
}

func decreaseMemoryCursor(count int) {
	uiState.memoryCursor -= count
	if uiState.memoryCursor < 0 {
		uiState.memoryCursor = dsp.DELAY_RAM_SIZE + uiState.memoryCursor
	}

	UpdateScreen(lastOpCodes, lastState, lastSampleNum)
}

func updateHelpLineView() {
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

	uiState.helpLineView = helpLine
}

func renderHelpScreen() {
	ypos := 0

	frame := widgets.NewParagraph()
	frame.Title = "  Help / Keys / Keywords  "
	frame.TitleStyle = boxTitleStyle
	frame.SetRect(0, 0, uiState.terminalWidth, uiState.terminalHeight)
	ypos += 1

	keys := widgets.NewList()
	keys.Border = false
	keys.TextStyle = termui.NewStyle(termui.ColorYellow)
	keys.SetRect(1, 1, uiState.terminalWidth-1, uiState.terminalHeight-1)
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

	keys.SetRect(1, ypos, uiState.terminalWidth-1, ypos+len(keys.Rows)+2)
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

	help.SetRect(1, ypos, uiState.terminalWidth-1, ypos+(uiState.terminalHeight-ypos)-12)
	ypos += 12

	ui.Render(frame)
	ui.Render(keys)
	ui.Render(help)
}

// Prints the code with a highlighted current-op
func updateCodeView(opCodes []base.Op, state *dsp.State) {
	width := uiState.centerLine
	height := uiState.terminalHeight - 5 - 1

	code := widgets.NewParagraph()
	code.Title = fmt.Sprintf("  Instructions (%d) ", len(opCodes))
	code.TitleStyle = boxTitleStyle

	code.Text = generateCodeListing(opCodes, state)
	code.SetRect(0, 0, width, height)

	uiState.codeView = code
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
	reg := 0.0

	sin0Reg, sin1Reg, _, _ := state.GetLFORegisterValues()

	if typ == base.LFO_SIN0 {
		sin = state.Sin0Osc.GetSine()
		cos = state.Sin0Osc.GetCosine()
		reg = sin0Reg
	} else {
		sin = state.Sin1Osc.GetSine()
		cos = state.Sin1Osc.GetCosine()
		reg = sin1Reg
	}

	lfoStr := fmt.Sprintf(" [SIN%d](fg:yellow)  [Rate:](fg:cyan) %d ", typ, sinRate)
	lfoStr += fmt.Sprintf("[(%.2f hz)](fg:gray)   ", sinhz)
	lfoStr += fmt.Sprintf("[Amp:](fg:cyan) %d [(%.2f%%)](fg:gray)   ", sinRange, sinrange)
	lfoStr += fmt.Sprintf("[Value:](fg:cyan) %f\n", sin)
	lfoStr += fmt.Sprintf("       [Cos:](fg:cyan) %.3f   [CompC:](fg:cyan) %.3f   [CompA:](fg:cyan) %.3f\n",
		cos, 1.0-sin, -sin)
	lfoStr += fmt.Sprintf("       [Reg:](fg:cyan) %.2f\n", reg)

	return lfoStr
}

func makeRampString(typ int, state *dsp.State) string {
	lfo := 0.0
	reg := 0.0
	rateRegValue := int32(0)
	ampRegValue := int32(0)

	_, _, ramp0Reg, ramp1Reg := state.GetLFORegisterValues()

	if typ == base.LFO_RMP0 {
		rateRegValue = state.Registers[base.RAMP0_RATE].Value
		ampRegValue = state.Registers[base.RAMP0_RANGE].Value
		lfo = state.Ramp0Osc.GetValue()
		reg = ramp0Reg
	} else {
		rateRegValue = state.Registers[base.RAMP1_RATE].Value
		ampRegValue = state.Registers[base.RAMP1_RANGE].Value
		lfo = state.Ramp1Osc.GetValue()
		reg = ramp1Reg
	}
	rmphz := (float64(rateRegValue) / 32767) * 20.0

	lfoStr := fmt.Sprintf(" [RAMP%d](fg:yellow) [Rate:](fg:cyan) %d ",
		typ-2, rateRegValue)
	lfoStr += fmt.Sprintf("[(%.2f hz)](fg:gray)   ", rmphz)
	lfoStr += fmt.Sprintf("[Range:](fg:cyan) %d   ", ampRegValue)
	lfoStr += fmt.Sprintf("[Value:](fg:cyan) %f\n", lfo)

	lfoStr += fmt.Sprintf("       [NA:](fg:cyan) %.3f   [CompC:](fg:cyan) %.3f   [CompA:](fg:cyan) %.3f\n",
		dsp.GetXFadeFromLFO(lfo, typ, state),
		1.0-lfo,
		dsp.GetRampRange(typ, state)-lfo)
	lfoStr += fmt.Sprintf("       [Half:](fg:cyan) %.2f  [Reg:](fg:cyan) %.2f\n",
		dsp.GetLFOValuePlusHalfCycle(typ, state),
		reg)
	return lfoStr
}

// Prints all values in the state
func updateStateView(sampleNum int, state *dsp.State) {
	twidth := uiState.terminalWidth
	hPos := uiState.centerLine

	rfInt := 0
	if state.RUN_FLAG {
		rfInt = 1
	}
	stateStr := fmt.Sprintf(" [ACC:](fg:yellow,mod:bold) %s (%d)   "+
		"[PACC:](fg:yellow) %s   [LR:](fg:yellow) %s\n"+
		" [ADDR_PTR:](fg:yellow) %d   [DelayRAMPtr:](fg:yellow) %d   [RUN:](fg:yellow) %d\n",
		// FIXME: Is the ACC an S.23 or an S1.14? (20220305 handegar)
		overflowColored(state.ACC.ToFloat64(), -1.0, 1.0), state.ACC.ToInt32(),
		overflowColored(state.PACC.ToFloat64(), -1.0, 1.0),
		overflowColored(state.LR.ToFloat64(), -1.0, 1.0),
		state.Registers[base.ADDR_PTR].ToInt32(),
		state.DelayRAMPtr,
		rfInt)

	ioStr := fmt.Sprintf(" [ADCL:](fg:yellow) %f   [ADCR:](fg:yellow) %f\n"+
		" [DACL:](fg:green) %s   [DACR:](fg:green) %s\n",
		state.Registers[base.ADCL].ToFloat64(), state.Registers[base.ADCR].ToFloat64(),
		overflowColored(state.Registers[base.DACL].ToFloat64(), -1.0, 1.0),
		overflowColored(state.Registers[base.DACR].ToFloat64(), -1.0, 1.0))

	potStr := fmt.Sprintf(" [POT0:](fg:cyan) %.2f   [POT1:](fg:cyan) %.2f   [POT2:](fg:cyan) %.2f\n",
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
	stateP.SetRect(hPos, vPos, twidth, vPos+7)
	vPos += 7

	lfoP := widgets.NewParagraph()
	lfoP.Title = "  LFOs  "
	lfoP.TitleStyle = boxTitleStyle
	lfoP.BorderStyle = termui.NewStyle(termui.ColorGreen)
	lfoP.Text = lfoStr
	lfoP.SetRect(hPos, vPos, twidth, vPos+14)
	vPos += 10

	uiState.mainStateView = stateP
	uiState.lfoStateView = lfoP
}

func updateRegistersView(state *dsp.State) {
	twidth := uiState.terminalWidth
	theight := uiState.terminalHeight - 1 // Save one for the keys line
	vPos := 21
	hPos := uiState.centerLine

	valAsStr := func(registerNr int) string {
		if uiState.showRegistersAsFloats {
			return fmt.Sprintf("%s", overflowColored(state.Registers[registerNr].ToFloat64(), -1.0, 1.0))
		} else {
			return fmt.Sprintf("%d", state.Registers[registerNr].ToInt32())
		}
	}

	regStr := ""
	// Special registers
	regStr += fmt.Sprintf(" [SIN0_RATE:](fg:cyan) %s  ", valAsStr(base.SIN0_RATE))
	regStr += fmt.Sprintf(" [SIN0_RANGE:](fg:cyan) %s\n", valAsStr(base.SIN0_RANGE))
	regStr += fmt.Sprintf(" [SIN1_RATE:](fg:cyan) %s  ", valAsStr(base.SIN1_RATE))
	regStr += fmt.Sprintf(" [SIN1_RANGE:](fg:cyan) %s\n", valAsStr(base.SIN1_RANGE))

	regStr += "\n"

	regStr += fmt.Sprintf(" [RMP0_RATE:](fg:cyan) %s  ", valAsStr(base.RAMP0_RATE))
	regStr += fmt.Sprintf(" [RMP0_RANGE:](fg:cyan) %s\n", valAsStr(base.RAMP0_RANGE))
	regStr += fmt.Sprintf(" [RMP1_RATE:](fg:cyan) %s  ", valAsStr(base.RAMP1_RATE))
	regStr += fmt.Sprintf(" [RMP1_RANGE:](fg:cyan) %s\n", valAsStr(base.RAMP1_RANGE))

	regStr += "\n"

	regStr += fmt.Sprintf(" [POT0:](fg:cyan) %s  ", valAsStr(base.POT0))
	regStr += fmt.Sprintf(" [POT1:](fg:cyan) %s  ", valAsStr(base.POT1))
	regStr += fmt.Sprintf(" [POT2:](fg:cyan) %s\n", valAsStr(base.POT2))

	regStr += "\n"

	// General registers
	for i := 0x20; i <= 0x3f; i += 3 {
		regStr += fmt.Sprintf(" [REG%d:](fg:cyan)", i-0x20)

		if i-0x20 < 10 {
			regStr += " "
		}

		if uiState.showRegistersAsFloats {
			regStr += fmt.Sprintf(" %s  ", overflowColored(state.Registers[i].ToFloat64(), -1.0, 1.0))
		} else {
			regStr += fmt.Sprintf(" %d  ", state.Registers[i].ToInt32())
		}

		regStr += fmt.Sprintf(" [REG%d:](fg:cyan)", i+1-0x20)
		if i+1-0x20 < 10 {
			regStr += " "
		}

		if uiState.showRegistersAsFloats {
			regStr += fmt.Sprintf(" %s  ", overflowColored(state.Registers[i+1].ToFloat64(), -1.0, 1.0))
		} else {
			regStr += fmt.Sprintf(" %d  ", state.Registers[i+1].ToInt32())
		}

		if i < (0x3f - 1) {
			regStr += fmt.Sprintf(" [REG%d:](fg:cyan)", i+2-0x20)
			if i+2-0x20 < 10 {
				regStr += " "
			}

			if uiState.showRegistersAsFloats {
				regStr += fmt.Sprintf(" %s\n", overflowColored(state.Registers[i+1].ToFloat64(), -1.0, 1.0))
			} else {
				regStr += fmt.Sprintf(" %d\n", state.Registers[i+1].ToInt32())
			}
		}
	}

	regP := widgets.NewParagraph()
	regP.Title = "  Registers"
	if uiState.showRegistersAsFloats {
		regP.Title += " as floats."
	} else {
		regP.Title += " as integers."
	}
	regP.Title += " Toggle type with 'f'  "
	regP.TitleStyle = boxTitleStyle
	regP.BorderStyle = termui.NewStyle(termui.ColorGreen)
	regP.Text = regStr
	regP.SetRect(hPos, vPos, twidth, theight-5)

	uiState.registerView = regP
}

// Prints misc info regarding current state and op
func updateMetaInfoView(opCodes []base.Op, state *dsp.State) {
	op := opCodes[state.IP]
	opDoc := disasm.OpDocs[op.Name]

	twidth := uiState.terminalWidth
	theight := uiState.terminalHeight - 1 // Make one line free at the bottom for keys

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

	uiState.infoLineView = infoP
	uiState.versionLineView = versionP
}
