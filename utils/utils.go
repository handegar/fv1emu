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

	if intbits == 1 && fractionbits == 9 {
		return (uint32)(num*(1024.0/2.0)) & 0x7ff
	} else if intbits == 1 && fractionbits == 14 {
		return (uint32)(num*(32768.0/2.0)) & 0xffff
	}

	panic("Unsupported intbits/fractionbits combo")
}

func QFormatToFloat64(raw uint32, intbits int, fractionbits int) float64 {
	totalbits := intbits + fractionbits
	isSigned := raw >> (totalbits)
	_ = isSigned

	if intbits == 1 && fractionbits == 9 {
		i := (raw & (1 << fractionbits)) >> (fractionbits)
		f := (raw & base.ArgBitMasks[fractionbits])
		fmt.Printf("i=%d, f=%d (%b, %b)\n", i, f, i, f)

		return float64(i) + float64(f)/float64(base.ArgBitMasks[fractionbits])

		//return (uint32)(num*(1024.0/2.0)) & 0x7ff
		//return float64(raw / 512)
		/*
			num := int32((raw & base.ArgBitMasks[totalbits-1]) << 1)
			if isSigned == 1 {
				num = -num
			}

			fmt.Printf("num=%d, isSigned=%d\n", num, isSigned)

			r := float64(num) * math.Pow(2, -float64(fractionbits))
			return r*/

	} /*else if intbits == 1 && fractionbits == 14 {
		return (uint32)(num*(32768.0/2.0)) & 0xffff
	}*/

	panic("Unsupported intbits/fractionbits combo")
}

func PrintUInt32AsBinary(value uint32) {
	fmt.Println()
	fmt.Printf("     33322222222221111111111         \n")
	fmt.Printf("     21098765432109876543210987654321\n")
	fmt.Printf("val=[%32b], 0x%x, %d\n", int32(value), value, value)
}

// Q1.<bits> format
// FIXME: Test for S4.5, S4.19 & S.23 as well (20220203 handegar)
func PrintUInt32AsQFormat(intbits int, fractionbits int, value uint32) {
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

// S1.9 or S1.14: -2...1.9999
func Real2ToFloat(bits int, raw uint32) float32 {
	// FIXME: I suspect that this FP->FLOAT operation can be done
	// by simple AND+SHIFT operations -- No logic needed (20220128
	// handegar)

	//PrintUInt32AsQFormat(1, bits-2, raw)

	isSigned := raw >> (bits - 1)
	num := (raw & base.ArgBitMasks[bits-1]) << 1
	ret := (float32(num) / float32(base.ArgBitMasks[bits-1]))

	if isSigned == 1 {
		return -ret
	} else {
		return ret
	}

}

// S.10: -1...0.999999
func Real1ToFloat(bits int, raw uint32) float32 {
	// FIXME: I suspect that this FP->FLOAT operation can be done
	// by simple AND+SHIFT operations -- No logic needed (20220128
	// handegar)

	//PrintUInt32AsQFormat(0, bits-1, raw)

	isSigned := raw >> (bits - 1)
	unSigned := raw & base.ArgBitMasks[bits-1]
	ret := float32(unSigned) / float32(int(1)<<(bits-1))

	if isSigned == 1 {
		return ret - 1.0
	} else {
		return ret
	}
}

// S4.6: –16...15.999998
func Real4ToFloat(bits int, raw uint32) float32 {
	// FIXME: I suspect that this FP->FLOAT operation can be done
	// by simple AND+SHIFT operations -- No logic needed (20220128
	// handegar)

	isSigned := raw >> (bits - 1)
	unSigned := raw & base.ArgBitMasks[bits-1]
	ret := float32(unSigned) / float32(int(1)<<(bits-(1+4)))

	if isSigned == 1 {
		return -ret
	} else {
		return ret
	}
}

func TypeToString(t int) string {
	switch t {
	case base.Real_1_14:
		return "Real_S1.14"
	case base.Real_1_9:
		return "Real_S1.9"
	case base.Real_10:
		return "Real_S.19"
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
