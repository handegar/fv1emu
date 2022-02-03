package disasm

import (
	"fmt"
	"strings"

	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/settings"
	"github.com/handegar/fv1emu/utils"
)

func PrintCodeListing(opCodes []base.Op) {
	fmt.Printf("\n;;\n;; Dissassembly (%d opcodes)\n;;\n", len(opCodes))
	var skpTargets []int
	for pos, opCode := range opCodes {
		op := OpCodeToString(opCode)
		if opCode.Name == "SKP" {
			skpTargets = append(skpTargets, pos+int(opCode.Args[1].RawValue))
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

func OpCodeToString(opcode base.Op) string {
	ret := "  "

	switch opcode.Name {
	case "SKP":
		ret += op_SKP_ToString(opcode)
	case "NOP":
		ret += op_NOP_ToString(opcode)
	case "SOF":
		ret += op_SOF_ToString(opcode)
	case "EXP":
		ret += op_EXP_ToString(opcode)
	case "AND":
		ret += op_AND(opcode)
	case "OR":
		ret += op_OR(opcode)
	case "XOR":
		ret += op_XOR(opcode)
	case "NOT":
		ret += op_NOT(opcode)
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
	case "LOG":
		ret += op_LOG_ToString(opcode)
	case "WLDS":
		ret += op_WLDS_ToString(opcode)
	case "WLDR":
		ret += op_WLDR_ToString(opcode)
	case "CHO RDA":
		ret += op_CHO(opcode)
	case "CHO SOF":
		ret += op_CHO(opcode)
	case "CHO RDAL":
		ret += op_CHO(opcode)
	case "JAM":
		ret += op_JAM(opcode)
	case "MAXX":
		ret += op_MAXX(opcode)
	case "WRLX":
		ret += op_WRLX(opcode)
	case "WRHX":
		ret += op_WRHX(opcode)
	case "RMPA":
		ret += op_RMPA(opcode)
	case "CLR":
		ret += op_CLR(opcode)
	case "ABSA":
		ret += op_ABSA(opcode)
	default:
		ret += fmt.Sprintf("<%s 0b%b>", opcode.Name, opcode.RawValue)
	}

	diff := 25 - len(ret)
	if diff > 1 {
		for i := 0; i < diff; i++ {
			ret += " "
		}
	}

	ret += "\t;;"
	if settings.PrintDebug {
		ret += fmt.Sprintf(" [0b%32b] ", opcode.RawValue)
	}
	for i, v := range opcode.Args {
		binStr := ""
		if settings.PrintDebug {
			binStr = fmt.Sprintf("0b%b/", v.RawValue)
		}
		ret += fmt.Sprintf("#%d: %s0x%x (%dbit), ", i, binStr, v.RawValue, v.Len)
	}
	ret += "\n"

	return ret
}

func op_CLR(op base.Op) string {
	return fmt.Sprintf("CLR\t  ")
}

func op_ABSA(op base.Op) string {
	return fmt.Sprintf("ABSA\t  ")
}

func op_RMPA(op base.Op) string {
	return fmt.Sprintf("RMPA\t  %f",
		utils.Real2ToFloat(op.Args[0].Len, op.Args[0].RawValue))
}

func op_WRLX(op base.Op) string {
	return fmt.Sprintf("WRLX\t  %s, %f",
		base.Symbols[int(op.Args[0].RawValue)],
		utils.Real2ToFloat(op.Args[2].Len, op.Args[2].RawValue))
}

func op_WRHX(op base.Op) string {
	return fmt.Sprintf("WRHX\t  %s, %f",
		base.Symbols[int(op.Args[0].RawValue)],
		utils.Real2ToFloat(op.Args[2].Len, op.Args[2].RawValue))
}

func op_MAXX(op base.Op) string {
	return fmt.Sprintf("MAXX\t  %s, %f",
		base.Symbols[int(op.Args[0].RawValue)],
		utils.Real2ToFloat(op.Args[2].Len, op.Args[2].RawValue))
}

func op_JAM(op base.Op) string {
	return fmt.Sprintf("JAM\t  %d", op.Args[1].RawValue)
}

func op_AND(op base.Op) string {
	return fmt.Sprintf("AND\t  %%%b", op.Args[1].RawValue)
}

func op_OR(op base.Op) string {
	return fmt.Sprintf("OR\t  %%%b", op.Args[1].RawValue)
}

func op_XOR(op base.Op) string {
	return fmt.Sprintf("XOR\t  %%%b", op.Args[1].RawValue)
}

func op_NOT(op base.Op) string {
	return fmt.Sprintf("NOT\t")
}

func op_SKP_ToString(op base.Op) string {
	var cmds []string
	flags := int(op.Args[2].RawValue)
	for i := 0; i < len(base.SkpFlagSymbols); i++ {
		if (flags & (1 << i)) != 0 {
			cmds = append(cmds, base.SkpFlagSymbols[flags&(1<<i)])
		}
	}

	return fmt.Sprintf("SKP\t  %s, addr_%d",
		strings.Join(cmds, "|"),
		op.Args[1].RawValue+1)
}

func op_NOP_ToString(op base.Op) string {
	return fmt.Sprintf("NOP\t  \t")
}

func op_LDAX_ToString(op base.Op) string {
	return fmt.Sprintf("LDAX\t  %s",
		base.Symbols[int(op.Args[0].RawValue)])
}

func op_WRAX_ToString(op base.Op) string {
	return fmt.Sprintf("WRAX\t  %s, %f",
		base.Symbols[int(op.Args[0].RawValue)],
		utils.Real2ToFloat(op.Args[2].Len, op.Args[2].RawValue))
}

func op_RDAX_ToString(op base.Op) string {
	return fmt.Sprintf("RDAX\t  %s, %f",
		base.Symbols[int(op.Args[0].RawValue)],
		utils.Real2ToFloat(op.Args[2].Len, op.Args[2].RawValue))
}

func op_RDFX_ToString(op base.Op) string {
	return fmt.Sprintf("RDFX\t  %s, %f",
		base.Symbols[int(op.Args[0].RawValue)],
		utils.Real2ToFloat(op.Args[2].Len, op.Args[2].RawValue))
}

func op_MULX_ToString(op base.Op) string {
	return fmt.Sprintf("MULX\t  %s",
		base.Symbols[int(op.Args[0].RawValue)])
}

func op_WRA_ToString(op base.Op) string {
	return fmt.Sprintf("WRA\t  %d, %f",
		op.Args[0].RawValue,
		utils.Real2ToFloat(op.Args[1].Len, op.Args[1].RawValue))
}

func op_WRAP_ToString(op base.Op) string {
	return fmt.Sprintf("WRAP\t  %d, %f",
		op.Args[0].RawValue,
		utils.Real2ToFloat(op.Args[1].Len, op.Args[1].RawValue))
}

func op_RDA_ToString(op base.Op) string {
	return fmt.Sprintf("RDA\t  %d, %f",
		op.Args[0].RawValue,
		utils.Real2ToFloat(op.Args[1].Len, op.Args[1].RawValue))
}

func op_SOF_ToString(op base.Op) string {
	return fmt.Sprintf("SOF\t  %f, %f",
		utils.Real2ToFloat(op.Args[1].Len, op.Args[1].RawValue),
		utils.Real1ToFloat(op.Args[0].Len, op.Args[0].RawValue))
}

func op_EXP_ToString(op base.Op) string {
	return fmt.Sprintf("EXP\t  %f, %f",
		utils.Real2ToFloat(op.Args[1].Len, op.Args[1].RawValue),
		utils.Real1ToFloat(op.Args[0].Len, op.Args[0].RawValue))
}

func op_LOG_ToString(op base.Op) string {
	return fmt.Sprintf("LOG\t  %f, %f",
		utils.Real2ToFloat(op.Args[1].Len, op.Args[1].RawValue),
		utils.Real4ToFloat(op.Args[0].Len, op.Args[0].RawValue))
}

func op_WLDS_ToString(op base.Op) string {
	amp := int(op.Args[0].RawValue)
	freq := int(op.Args[1].RawValue)
	typ := "SIN0"
	if op.Args[2].RawValue == 1 {
		typ = "SIN1"
	}
	return fmt.Sprintf("WLDS\t  %s, %d, %d", typ, freq, amp)
}

func op_WLDR_ToString(op base.Op) string {
	amp := base.RampAmpValues[int(op.Args[0].RawValue)]
	freq := int(op.Args[2].RawValue)
	typ := "RMP0"
	if op.Args[3].RawValue == 1 {
		typ = "RMP1"
	}
	return fmt.Sprintf("WLDR\t  %s, %d, %d", typ, freq, amp)
}

func op_CHO(op base.Op) string {
	addr := (int(op.Args[0].RawValue) << 1) >> 1
	typ := "<?>"
	switch op.Args[1].RawValue {
	case 0b0:
		typ = "SIN0"
	case 0b01:
		typ = "SIN1"
	case 0b10:
		typ = "RMP0"
	case 0b11:
		typ = "RMP1"
	}
	var flags []string
	f := int(op.Args[3].RawValue)
	if f != 0 {
		for i := 0; i < len(base.ChoFlagSymbols); i++ {
			if (f & (1 << i)) != 0 {
				flags = append(flags, base.ChoFlagSymbols[f&(1<<i)])
			}
		}
	} else {
		flags = append(flags, base.ChoFlagSymbols[0])
	}

	cmd := ""
	switch op.Args[4].RawValue {
	case 0b10:
		return fmt.Sprintf("CHO\t  SOF, %s, %s, %d",
			typ, strings.Join(flags, "|"), addr)

	case 0b0:
		return fmt.Sprintf("CHO\t  RDA, %s, %s, %d",
			typ, strings.Join(flags, "|"), addr)

	case 0b11:
		cmd = "RDAL"
		return fmt.Sprintf("CHO\t  RDAL, %s", typ)

	default:
		cmd = fmt.Sprintf("<0b%b>", op.Args[4].RawValue)
	}

	return fmt.Sprintf("CHO %s\t", cmd)
}
