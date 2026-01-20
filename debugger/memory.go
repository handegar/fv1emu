package debugger

import (
	"fmt"

	termui "github.com/gizak/termui/v3"
	ui "github.com/gizak/termui/v3"
	widgets "github.com/gizak/termui/v3/widgets"

	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/dsp"
	"github.com/handegar/fv1emu/utils"
)

const VALUE_TABLE_ROWS = 1
const CURSOR_COLOR = "(bg:green)"
const VALUE_COLOR = "(fg:red)"

func renderMemoryMap(state *dsp.State, sampleNum int, cursorPosition int) {
	width, height := termui.TerminalDimensions()
	ypos := 0

	memMap := widgets.NewParagraph()
	memMap.Title = fmt.Sprintf("  '\u2593': 3/3 set | '\u2592': 2/3 set | '\u2591': 1/3 set | '.': Zero  "+
		"(sample #%d)  ", sampleNum)
	memMap.TitleStyle = termui.NewStyle(termui.ColorYellow, termui.ColorBlue)
	memMapTxt, valuesPerChar := buildMemMapText(width-1, height-1, state)
	memMapTxt = colorMemoryMap(memMapTxt, cursorPosition, valuesPerChar)

	memMap.Text = memMapTxt
	memMap.BorderStyle = termui.NewStyle(termui.ColorGreen)
	memMap.SetRect(0, 0, width, height-(VALUE_TABLE_ROWS+3))
	ypos += height - (VALUE_TABLE_ROWS + 3)

	infoP := widgets.NewParagraph()
	infoP.Border = false
	infoP.PaddingBottom = 0
	infoP.PaddingTop = 0
	infoP.PaddingLeft = 0
	infoP.PaddingRight = 0
	txt := fmt.Sprintf("'P': ADDR_PTR | One character is %d values", valuesPerChar)
	infoP.Text = fmt.Sprintf("[%s](fg:blue)", txt)
	infoP.SetRect(width-len(txt)-4, ypos-1, width-2, ypos)
	infoP.TextStyle = termui.NewStyle(termui.ColorBlue)

	cursorP := widgets.NewParagraph()
	cursorP.Border = true
	cursorP.PaddingBottom = 0
	cursorP.PaddingTop = 0
	cursorP.PaddingLeft = 0
	cursorP.PaddingRight = 0
	txt = fmt.Sprintf("%d", cursorPosition)
	cursorP.Text = "[" + txt + "](fg:yellow)"
	curX, curY := calculateCursorPosition(cursorPosition, valuesPerChar, width)
	cursorP.SetRect(curX, curY+2, curX+len(txt)+2, curY+3+2)

	zoomTable := buildZoomTable(state, VALUE_TABLE_ROWS+3, cursorPosition)
	zoomTable.SetRect(0, ypos, width, ypos+VALUE_TABLE_ROWS+3)
	ypos += VALUE_TABLE_ROWS + 3

	ui.Render(memMap)
	ui.Render(cursorP)
	ui.Render(infoP)
	ui.Render(zoomTable)
}

func calculateCursorPosition(pos int, valuesPerChar int, width int) (int, int) {
	width = width - 1
	ypos := (pos / valuesPerChar) / width
	xpos := (pos / valuesPerChar) - ypos*width
	return xpos, ypos
}

func buildZoomTable(state *dsp.State, columns int, cursorPosition int) *widgets.Table {
	table := widgets.NewTable()
	table.RowSeparator = false

	valuesPerChar := calculateMemoryValuesPerChar()
	index := cursorPosition / valuesPerChar
	index = index * valuesPerChar

	// Make header row
	var header []string
	for i := 0; i < valuesPerChar; i++ {
		header = append(header, fmt.Sprintf("%d", cursorPosition+i))
	}
	table.Rows = append(table.Rows, header)
	table.RowStyles[0] = ui.NewStyle(ui.ColorYellow)

	memPos := index

	// Fill in values
	for i := 0; i < VALUE_TABLE_ROWS; i++ {
		var values []string
		for j := 0; j < valuesPerChar; j++ {
			val := utils.QFormatToFloat64(state.DelayRAM[memPos+(i*valuesPerChar)+j], 0, 23)
			values = append(values, fmt.Sprintf("%f", val))
		}
		table.Rows = append(table.Rows, values)
		table.RowStyles[i+1] = ui.NewStyle(ui.ColorWhite, ui.ColorBlack, ui.ModifierBold)
	}

	return table
}

func calculateMemoryValuesPerChar() int {
	width, height := termui.TerminalDimensions()
	width = width - 1
	numChars := width * (height - ((VALUE_TABLE_ROWS + 3) + 3))
	ret := dsp.DELAY_RAM_SIZE / numChars

	if ret < 8 {
		return 8
	}

	return ret
}

func buildMemMapText(width int, height int, state *dsp.State) (string, int) {
	width = width - 1
	valuesPerChar := calculateMemoryValuesPerChar()
	numLines := (dsp.DELAY_RAM_SIZE / valuesPerChar) / width
	currentAddrPtr := int(state.GetRegister(base.ADDR_PTR).ToInt32() >> 8)

	ret := ""
	c := 0
	for i := 0; i < numLines; i++ {
		for j := 0; j < width; j++ {
			numNonNull := 0
			for k := 0; k < valuesPerChar; k++ {
				if c+k == currentAddrPtr {
					numNonNull = -1
					break
				}
				if state.DelayRAM[c+k] != 0.0 {
					numNonNull += 1
				}
			}
			c += valuesPerChar

			if numNonNull < 0 {
				ret += "P"
			} else {
				f := float64(numNonNull) / float64(valuesPerChar)
				if f == 0 {
					ret += "."
				} else if f < 1.0/3.0 {
					ret += "\u2591"
				} else if f < 2.0/3.0 {
					ret += "\u2592"
				} else if f <= 1.0 {
					ret += "\u2593"
				}
			}
		}
		ret += "\n"
	}

	return ret, valuesPerChar
}

func colorMemoryMap(mem string, cursorPosition int, valuesPerChar int) string {
	c := 0
	var isNonZero = func(char byte) bool {
		return char != '.' && char != '\n' && char != 'P'
	}

	var isZero = func(char byte) bool {
		return char == '.'
	}

	coloring := false

	for i := 0; i < len(mem); i++ {
		if c >= len(mem) { // Safety
			break
		}

		if isNonZero(mem[c]) && !coloring { // Start coloring
			mem = mem[:c] + "[" + mem[c:]
			c += 1
			coloring = true
		} else if isZero(mem[c]) && coloring { // Stop coloring
			inj := "]" + VALUE_COLOR
			mem = mem[:c] + inj + mem[c:]
			c += len(inj)
			coloring = false
		}

		c += 1
	}

	if coloring {
		mem += "]" + VALUE_COLOR
	}

	return mem
}
