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
const BLOCKS_PER_LINE = 64
const VALUES_PER_BLOCK = 8
const MARGIN_LEFT = "       "

func buildAxis(blocksPerLine int, valuesPerBlock int) *widgets.Paragraph {
	axisP := widgets.NewParagraph()
	axisP.Border = false

	axis := "["

	axis += "      0"
	for i := 0; i < (blocksPerLine - 2); i += 1 {
		axis += "-"
	}
	axis += fmt.Sprintf("%d\n", blocksPerLine*valuesPerBlock-1)
	width := len(axis)

	numLines := (dsp.DELAY_RAM_SIZE / valuesPerBlock) / blocksPerLine

	axis += "\n"
	for i := 0; i < numLines; i += 1 {
		axis += fmt.Sprintf("%5d\n", i*valuesPerBlock*blocksPerLine)
	}

	axis += "](fg:yellow)\n"

	axisP.Text = axis

	height := uiState.terminalHeight
	contentHeight := numLines + 5
	if contentHeight > height-5 {
		contentHeight = height - 5
	}

	axisP.SetRect(1, 1, width+1, contentHeight)

	return axisP
}

func renderMemoryMap(state *dsp.State, sampleNum int, cursorPosition int) {
	width := uiState.terminalWidth
	height := uiState.terminalHeight
	ypos := 0

	memMap := widgets.NewParagraph()
	memMap.Title = fmt.Sprintf("  '\u2593': 3/3 set | '\u2592': 2/3 set | '\u2591': 1/3 set | '.': Zero  "+
		"(sample #%d)  ", sampleNum)
	memMap.TitleStyle = boxTitleStyle
	memMap.BorderStyle = termui.NewStyle(termui.ColorGreen)
	memMap.SetRect(0, 0, width, height-(VALUE_TABLE_ROWS+3))
	ui.Render(memMap)

	topAxis := buildAxis(BLOCKS_PER_LINE, VALUES_PER_BLOCK)
	ui.Render(topAxis)

	memMapTxt := buildMemMapText(width-1, state)
	memMapTxt = colorMemoryMap(memMapTxt, cursorPosition)

	memMapContent := widgets.NewParagraph()
	memMapContent.Border = false
	memMapContent.Text = memMapTxt
	numLines := (dsp.DELAY_RAM_SIZE / VALUES_PER_BLOCK) / BLOCKS_PER_LINE
	memMapContentHeight := numLines + 5
	if memMapContentHeight > height-5 {
		memMapContentHeight = height - 5
	}
	memMapContent.SetRect(len(MARGIN_LEFT), 3, width, memMapContentHeight)
	ui.Render(memMapContent)

	ypos += height - (VALUE_TABLE_ROWS + 3)

	infoP := widgets.NewParagraph()
	infoP.Border = false
	infoP.PaddingBottom = 0
	infoP.PaddingTop = 0
	infoP.PaddingLeft = 0
	infoP.PaddingRight = 0
	txt := fmt.Sprintf("'P': ADDR_PTR | One character is %d values", VALUES_PER_BLOCK)
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
	cursorP.Text = "[" + txt + "](fg:cyan)"
	cursorP.Title = "\\"
	curX, curY := calculateCursorPosition(cursorPosition, width)
	cursorP.SetRect(curX, curY+2, curX+len(txt)+2, curY+3+2)

	zoomTable := buildZoomTable(state, cursorPosition, VALUES_PER_BLOCK)
	zoomTable.SetRect(0, ypos, width, ypos+VALUE_TABLE_ROWS+3)
	ypos += VALUE_TABLE_ROWS + 3

	ui.Render(cursorP)
	ui.Render(infoP)
	ui.Render(zoomTable)
}

func calculateCursorPosition(pos int, width int) (int, int) {
	width = BLOCKS_PER_LINE // - 1
	ypos := (pos / VALUES_PER_BLOCK) / width
	xpos := (pos / VALUES_PER_BLOCK) - ypos*width
	return xpos + len(MARGIN_LEFT), ypos + 3
}

func buildZoomTable(state *dsp.State, cursorPosition int, valuesPerBlock int) *widgets.Table {
	table := widgets.NewTable()
	table.RowSeparator = false

	index := cursorPosition

	// Make header row
	var header []string
	for i := 0; i < valuesPerBlock; i++ {
		header = append(header, fmt.Sprintf("%d", cursorPosition+i))
	}
	table.Rows = append(table.Rows, header)
	table.RowStyles[0] = ui.NewStyle(ui.ColorYellow)

	memPos := index

	// Fill in values
	for i := 0; i < VALUE_TABLE_ROWS; i++ {
		var values []string
		for j := 0; j < valuesPerBlock; j++ {
			if memPos+(i*valuesPerBlock)+j > (32768 - 1) {
				break
			}
			val := utils.QFormatToFloat64(state.DelayRAM[memPos+(i*valuesPerBlock)+j], 0, 23)
			values = append(values, fmt.Sprintf("%f", val))
		}
		table.Rows = append(table.Rows, values)
		table.RowStyles[i+1] = ui.NewStyle(ui.ColorWhite, ui.ColorBlack, ui.ModifierBold)
	}

	return table
}

func buildMemMapText(width int, state *dsp.State) string {
	width = BLOCKS_PER_LINE
	numLines := (dsp.DELAY_RAM_SIZE / VALUES_PER_BLOCK) / width
	currentAddrPtr := int(state.GetRegister(base.ADDR_PTR).ToInt32() >> 8)

	ret := ""
	c := 0
	for i := 0; i < numLines; i++ {
		for j := 0; j < width; j++ {
			numNonNull := 0
			for k := 0; k < VALUES_PER_BLOCK; k++ {
				if c+k == currentAddrPtr {
					numNonNull = -1
					break
				}
				if state.DelayRAM[c+k] != 0.0 {
					numNonNull += 1
				}
			}
			c += VALUES_PER_BLOCK

			if numNonNull < 0 {
				ret += "P"
			} else {
				f := float64(numNonNull) / float64(VALUES_PER_BLOCK)
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

	return ret
}

func colorMemoryMap(mem string, cursorPosition int) string {
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
