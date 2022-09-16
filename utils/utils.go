package utils

import (
	"fmt"
	"math"
	"strconv"

	"github.com/handegar/fv1emu/base"
)

func Assert(mustBeTrue bool, msg string, args ...interface{}) {
	if !mustBeTrue {
		fmt.Printf("ERROR: %s\n", fmt.Sprintf("%s", args...))
		panic("ASSERT failed")
	}
}

func AssertFloat64(mustBeTrue bool, val float64, msg string, args ...interface{}) {
	if !mustBeTrue {
		fmt.Printf("ERROR: %s.\n       Value was %f.\n",
			fmt.Sprintf(msg, args...), val)
		panic("ASSERT failed")
	}
}

func StringToQFormat(str string, intbits int, fractionbits int) int32 {
	num, err := strconv.ParseFloat(str, 64)
	if err != nil {
		panic("Error parsing string to double")
	}
	return Float64ToQFormat(num, intbits, fractionbits)
}

func Float64ToQFormat(val float64, intbits int, fractionbits int) int32 {
	// FIXME: Perform some limit-checks here (20220204 handegar)
	return int32(val * math.Pow(2, float64(fractionbits)))
}

// Q<bits>.<bits> format (aka S<bits>.<bits>)
// Ex: Q1.9, Q1.14, Q.10
func QFormatToFloat64(raw int32, intbits int, fractionbits int) float64 {
	if intbits > 1 {
		return qFormatToFloat64_alternative(raw, intbits, fractionbits)
	}

	totalbits := intbits + fractionbits + 1

	p1 := int32((1 << (totalbits - 1)) - 1)
	p2 := int32(1 << (totalbits - 1))
	p3 := int32(1 << fractionbits)
	v1 := raw & p1
	v2 := raw & p2
	y1 := int32(v1) - int32(v2)
	return float64(y1) / float64(int32(p3))

}

// This is an alternative "more academic" way to transform a Q-format
// integer to a float. Needed to do S4.6 numbers
func qFormatToFloat64_alternative(raw int32, intbits int, fractionbits int) float64 {
	totalbits := intbits + fractionbits + 1
	isSigned := raw >> (intbits + fractionbits)
	unsigned := (raw & ((1 << (totalbits - 1)) - 1))
	ret := float64(unsigned) / float64(uint(1)<<fractionbits)
	if isSigned == 1 {
		ret = -ret
	}
	return ret
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

	if value > ((1 << totalbits) - 1) {
		fmt.Printf("PrintUInt32AsQFormat(): WARNING: Value 0x%x for S%d.%d is larger than 0x%x\n",
			value, intbits, fractionbits, ((1 << totalbits) - 1))
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
		fmt.Printf("[%b | ", value>>(totalbits-1)) // Signed bit
		fmt.Printf("%b | ", (value&((1<<(totalbits-1))-1))>>fractionbits)

	} else {
		fmt.Printf(" S   FRACTION\n")
		fmt.Printf("[%b | ", value>>(totalbits-1)) // Signed bit
	}

	strFmt = fmt.Sprintf("'%%%db']\n\n", fractionbits)
	fmt.Printf(strFmt, value&((1<<fractionbits)-1))
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

// Scale up an S1.14 to S1.24
func S114_to_S124(val int32) int32 {
	return val << (25 - 16)
}

// Scale up an S1.9 to S1.24
func S19_to_S124(val int32) int32 {
	return val << (25 - 11)
}

// Scale up an S.10 to S1.24
func S10_to_S124(val int32) int32 {
	return val << (25 - 11)
}
