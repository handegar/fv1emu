package utils

import (
	"fmt"

	"github.com/handegar/fv1emu/dsp"
	"github.com/handegar/fv1emu/settings"
)

func OpCodeToString(opcode dsp.Op) string {
	ret := " "

	switch opcode.Name {
	case "SKP":
		ret += op_SKP_ToString(opcode)
	case "SOF":
		ret += op_SOF_ToString(opcode)
	case "EXP":
		ret += op_EXP_ToString(opcode)
	case "LDAX":
		ret += op_LDAX_ToString(opcode)
	case "WRAX":
		ret += op_WRAX_ToString(opcode)
	case "MULX":
		ret += op_MULX_ToString(opcode)
	case "WRA":
		ret += op_WRA_ToString(opcode)
	case "WRAP":
		ret += op_WRAP_ToString(opcode)
	case "RDA":
		ret += op_RDA_ToString(opcode)
	case "RDAX":
		ret += op_RDAX_ToString(opcode)
	case "RDFX":
		ret += op_RDFX_ToString(opcode)
	default:
		ret += fmt.Sprintf("<%s 0b%b>", opcode.Name)
	}

	diff := 25 - len(ret)
	if diff > 1 {
		for i := 0; i < diff; i++ {
			ret += " "
		}
	}

	ret += "\t;;"
	ret += fmt.Sprintf(" [0b%32b] ", opcode.RawValue)
	for i, v := range opcode.Args {
		ret += fmt.Sprintf("#%d: 0b%b/0x%x (%dbit), ", i, v.RawValue, v.RawValue, v.Len)
	}
	ret += "\n"

	return ret
}

func op_SKP_ToString(op dsp.Op) string {
	return fmt.Sprintf("SKP\t %s, addr_%d",
		dsp.Symbols[int(op.Args[2].RawValue)],
		op.Args[1].RawValue+1)
}

func op_LDAX_ToString(op dsp.Op) string {
	return fmt.Sprintf("LDAX\t %s",
		dsp.Symbols[int(op.Args[0].RawValue)])
}

func op_WRAX_ToString(op dsp.Op) string {
	return fmt.Sprintf("WRAX\t %s, %f",
		dsp.Symbols[int(op.Args[0].RawValue)],
		Real2ToFloat(op.Args[2].Len, op.Args[2].RawValue))
}

func op_RDAX_ToString(op dsp.Op) string {
	return fmt.Sprintf("RDAX\t %s, %f",
		dsp.Symbols[int(op.Args[0].RawValue)],
		Real2ToFloat(op.Args[2].Len, op.Args[2].RawValue))
}

func op_RDFX_ToString(op dsp.Op) string {
	return fmt.Sprintf("RDFX\t %s, %f",
		dsp.Symbols[int(op.Args[0].RawValue)],
		Real2ToFloat(op.Args[2].Len, op.Args[2].RawValue))
}

func op_MULX_ToString(op dsp.Op) string {
	return fmt.Sprintf("MULX\t %s",
		dsp.Symbols[int(op.Args[0].RawValue)])
}

func op_WRA_ToString(op dsp.Op) string {
	return fmt.Sprintf("WRA\t %d, %f",
		op.Args[0].RawValue,
		Real2ToFloat(op.Args[1].Len, op.Args[1].RawValue))
}

func op_WRAP_ToString(op dsp.Op) string {
	return fmt.Sprintf("WRAP\t %d, %f",
		op.Args[0].RawValue,
		Real2ToFloat(op.Args[1].Len, op.Args[1].RawValue))
}

func op_RDA_ToString(op dsp.Op) string {
	return fmt.Sprintf("RDA\t %d, %f",
		op.Args[0].RawValue,
		Real2ToFloat(op.Args[2].Len, op.Args[2].RawValue))
}

func op_SOF_ToString(op dsp.Op) string {
	return fmt.Sprintf("SOF\t %f, %f",
		Real2ToFloat(op.Args[1].Len, op.Args[1].RawValue),
		Real1ToFloat(op.Args[0].Len, op.Args[0].RawValue))
}

func op_EXP_ToString(op dsp.Op) string {
	return fmt.Sprintf("EXP\t %f, %f",
		Real2ToFloat(op.Args[1].Len, op.Args[1].RawValue),
		Real1ToFloat(op.Args[0].Len, op.Args[0].RawValue))
}

// S1.9 or S1.14: -2...1.9999
func Real2ToFloat(bits int, raw uint32) float32 {
	isSigned := raw >> (bits - 1)
	num := (raw & dsp.ArgBitMasks[bits-1]) << 1
	ret := (float32(num) / float32(dsp.ArgBitMasks[bits-1]))
	if isSigned == 1 {
		return -ret
	} else {
		return ret
	}
}

// S.10: -1...0.999999
func Real1ToFloat(bits int, raw uint32) float32 {
	isSigned := raw >> (bits - 1)
	num := (raw & dsp.ArgBitMasks[bits-1])
	ret := (float32(num) / float32(dsp.ArgBitMasks[bits-1]))
	if isSigned == 1 {
		return -ret
	} else {
		return ret
	}
}

func PrintCodeListing(opCodes []dsp.Op) {
	fmt.Printf("\n;;\n;; Dissassembly (%d opcodes)\n;;\n", len(opCodes))
	var skpTargets []int
	for pos, opCode := range opCodes {
		op := OpCodeToString(opCode)
		if opCode.Name == "SKP" {
			skpTargets = append(skpTargets, int(opCode.Args[1].RawValue))
		}

		// Is current "pos" registered in 'skpTargets'?
		for _, p := range skpTargets {
			if p == (pos - 1) {
				fmt.Printf("addr_%d:\n", pos)
			}
		}

		fmt.Print(op)

		if pos > settings.MaxNumberOfOps {
			fmt.Printf(";; Max number of instructions reached (%d)\n",
				settings.MaxNumberOfOps)
			break
		}
	}
	fmt.Println()
}
