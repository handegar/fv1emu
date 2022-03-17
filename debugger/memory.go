package debugger

import (
	"fmt"

	termui "github.com/gizak/termui/v3"
	ui "github.com/gizak/termui/v3"
	widgets "github.com/gizak/termui/v3/widgets"

	"github.com/handegar/fv1emu/dsp"
)

func renderMemoryMap(state *dsp.State, sampleNum int) {
	width, height := termui.TerminalDimensions()

	valuesPerChar := 0
	memMap := widgets.NewParagraph()
	memMap.Title = fmt.Sprintf("  '\u2593': All values set | '\u2592': Some values set | '.': All zero  "+
		"(sample #%d)  ", sampleNum)
	memMap.TitleStyle = termui.NewStyle(termui.ColorYellow, termui.ColorBlue)
	memMap.SetRect(0, 0, width, height)
	memMap.Text, valuesPerChar = buildMemMapText(width-1, height-1, state)
	memMap.BorderStyle = termui.NewStyle(termui.ColorGreen)

	infoP := widgets.NewParagraph()
	infoP.Border = false
	infoP.PaddingBottom = 0
	infoP.PaddingTop = 0
	infoP.PaddingLeft = 0
	infoP.PaddingRight = 0
	txt := fmt.Sprintf("One char is %d values", valuesPerChar)
	infoP.Text = fmt.Sprintf("[%s](fg:blue)", txt)
	infoP.SetRect(width-len(txt)-4, height-1, width-2, height)
	infoP.TextStyle = termui.NewStyle(termui.ColorBlue)

	ui.Render(memMap)
	ui.Render(infoP)
}

func buildMemMapText(width int, height int, state *dsp.State) (string, int) {
	width = width - 1
	numChars := width * (height - 1)
	valuesPerChar := len(state.DelayRAM) / numChars
	numLines := (len(state.DelayRAM) / valuesPerChar) / width

	ret := ""
	c := 0
	for i := 0; i < numLines; i++ {
		for j := 0; j < width; j++ {
			numNonNull := 0
			for k := 0; k < valuesPerChar; k++ {
				if state.DelayRAM[c+k] != 0.0 {
					numNonNull += 1
				}
			}
			c += valuesPerChar
			if numNonNull == 0 {
				ret += "."
			} else if numNonNull == valuesPerChar {
				ret += "\u2593"
			} else {
				ret += "\u2592"
			}
		}
		ret += "\n"
	}

	on := false
	c = 0
	color := "(fg:green)"
	for i := 0; i < numChars; i++ {
		if ret[c] != '.' && !on {
			ret = ret[:c] + "[" + ret[c:]
			c += 1
			on = true
		} else if ret[c] == '.' && on {
			ret = ret[:c] + "]" + color + ret[c:]
			c += len(color) + 1
			on = false
		}

		c += 1
	}

	return ret, valuesPerChar
}
