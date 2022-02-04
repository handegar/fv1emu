package utils

import (
	"fmt"
	"strconv"

	"github.com/handegar/fv1emu/base"
)

func StringToQFormat(str string, intbits int, fractionbits int) uint32 {
	num, err := strconv.ParseFloat(str, 64)
	if err != nil {
		panic("Error parsing string to double")
	}
	return uint32(num*float64(int32(1)<<fractionbits)) & base.ArgBitMasks[intbits+fractionbits+1]
}

// Q<bits>.<bits> format (aka S<bits>.<bits>)
// Ex: Q1.9, Q1.14, Q.10
func QFormatToFloat64(raw uint32, intbits int, fractionbits int) float64 {
	totalbits := intbits + fractionbits + 1
	p1 := uint32((1 << (totalbits - 1)) - 1)
	p2 := uint32(1 << (totalbits - 1))
	p3 := uint32(1 << fractionbits)
	v1 := raw & p1
	v2 := raw & p2
	y1 := int32(v1) - int32(v2)
	return float64(y1) / float64(int32(p3))
}

func PrintUInt32AsBinary(value uint32) {
	fmt.Println()
	fmt.Printf("     33322222222221111111111         \n")
	fmt.Printf("     21098765432109876543210987654321\n")
	fmt.Printf("val=[%32b], 0x%x, %d\n", int32(value), value, value)
}

// Q<bits>.<bits> format (aka S<bits>.<bits>)
// Ex: Q1.9, Q1.14, Q.10
func PrintUInt32AsQFormat(value uint32, intbits int, fractionbits int) {
	totalbits := intbits + fractionbits + 1

	if value > base.ArgBitMasks[totalbits] {
		fmt.Printf("PrintUInt32AsQFormat(): ERROR: Value 0x%x for S%d.%d is larger than 0x%x\n",
			value, intbits, fractionbits, base.ArgBitMasks[intbits+fractionbits+1])
		return
	}

	if intbits > 0 {
		fmt.Printf("S%d.%d", intbits, fractionbits)
	} else {
		fmt.Printf("S.%d", fractionbits)
	}

	strFmt := fmt.Sprintf(" / 0x%%x / [0b%%%db]\n", totalbits)
	fmt.Printf(strFmt, value, value)

	if intbits > 0 {
		fmt.Printf(" S   I   FRACTION\n")
		fmt.Printf("[%b | ", value>>(totalbits-1))   // Signed bit
		fmt.Printf("%b | ", (value<<1)>>(totalbits)) // Integer bit

	} else {
		fmt.Printf(" S   FRACTION\n")
		fmt.Printf("[%b | ", value>>(totalbits-1)) // Signed bit
	}

	strFmt = fmt.Sprintf("'%%%db']\n\n", fractionbits)
	fmt.Printf(strFmt, value&base.ArgBitMasks[fractionbits])
}

func TypeToString(t int) string {
	switch t {
	case base.Real_1_14:
		return "Real_S1.14"
	case base.Real_1_9:
		return "Real_S1.9"
	case base.Real_10:
		return "Real_S.10"
	case base.Real_4_6:
		return "Real_S4.6"
	case base.Const:
		return "Const"
	case base.UInt:
		return "UInt"
	case base.Int:
		return "Int"
	case base.Bin:
		return "Bin"
	case base.Flag:
		return "Flag"
	case base.Blank:
		return "Blank"
	default:
		return "?"
	}
}
