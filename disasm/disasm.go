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
		op := OpCodeToString(opCode, pos, true)
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

func OpCodeToString(opcode base.Op, ip int, showParamData bool) string {
	ret := "  "

	switch opcode.Name {
	case "SKP":
		ret += SKP_ToString(opcode, ip)
	case "NOP":
		ret += NOP_ToString(opcode)
	case "SOF":
		ret += SOF_ToString(opcode)
	case "EXP":
		ret += EXP_ToString(opcode)
	case "AND":
		ret += AND_ToString(opcode)
	case "OR":
		ret += OR_ToString(opcode)
	case "XOR":
		ret += XOR_ToString(opcode)
	case "NOT":
		ret += NOT_ToString(opcode)
	case "LDAX":
		ret += LDAX_ToString(opcode)
	case "WRAX":
		ret += WRAX_ToString(opcode)
	case "MULX":
		ret += MULX_ToString(opcode)
	case "WRA":
		ret += WRA_ToString(opcode)
	case "WRAP":
		ret += WRAP_ToString(opcode)
	case "RDA":
		ret += RDA_ToString(opcode)
	case "RDAX":
		ret += RDAX_ToString(opcode)
	case "RDFX":
		ret += RDFX_ToString(opcode)
	case "LOG":
		ret += LOG_ToString(opcode)
	case "WLDS":
		ret += WLDS_ToString(opcode)
	case "WLDR":
		ret += WLDR_ToString(opcode)
	case "CHO RDA":
		ret += CHO_ToString(opcode)
	case "CHO SOF":
		ret += CHO_ToString(opcode)
	case "CHO RDAL":
		ret += CHO_ToString(opcode)
	case "JAM":
		ret += JAM_ToString(opcode)
	case "MAXX":
		ret += MAXX_ToString(opcode)
	case "WRLX":
		ret += WRLX_ToString(opcode)
	case "WRHX":
		ret += WRHX_ToString(opcode)
	case "RMPA":
		ret += RMPA_ToString(opcode)
	case "CLR":
		ret += CLR_ToString(opcode)
	case "ABSA":
		ret += ABSA_ToString(opcode)
	default:
		ret += fmt.Sprintf("<%s 0b%b>", opcode.Name, opcode.RawValue)
	}

	if showParamData {
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
	}

	return ret
}

func CLR_ToString(op base.Op) string {
	return fmt.Sprintf("CLR   ")
}

func ABSA_ToString(op base.Op) string {
	return fmt.Sprintf("ABSA  ")
}

func RMPA_ToString(op base.Op) string {
	return fmt.Sprintf("RMPA  %f",
		utils.QFormatToFloat64(op.Args[1].RawValue, 1, 9))
}

func WRLX_ToString(op base.Op) string {
	return fmt.Sprintf("WRLX  %s, %f",
		base.Symbols[int(op.Args[0].RawValue)],
		utils.QFormatToFloat64(op.Args[2].RawValue, 1, 14))
}

func WRHX_ToString(op base.Op) string {
	return fmt.Sprintf("WRHX  %s, %f",
		base.Symbols[int(op.Args[0].RawValue)],
		utils.QFormatToFloat64(op.Args[2].RawValue, 1, 14))
}

func MAXX_ToString(op base.Op) string {
	return fmt.Sprintf("MAXX  %s, %f",
		base.Symbols[int(op.Args[0].RawValue)],
		utils.QFormatToFloat64(op.Args[2].RawValue, 1, 14))
}

func JAM_ToString(op base.Op) string {
	return fmt.Sprintf("JAM   %d", op.Args[1].RawValue)
}

func AND_ToString(op base.Op) string {
	return fmt.Sprintf("AND   %%%b", op.Args[1].RawValue)
}

func OR_ToString(op base.Op) string {
	return fmt.Sprintf("OR    %%%b", op.Args[1].RawValue)
}

func XOR_ToString(op base.Op) string {
	return fmt.Sprintf("XOR   %%%b", op.Args[1].RawValue)
}

func NOT_ToString(op base.Op) string {
	return fmt.Sprintf("NOT   ")
}

func SKP_ToString(op base.Op, ip int) string {
	var cmds []string
	flags := int(op.Args[2].RawValue)
	for i := 0; i < len(base.SkpFlagSymbols); i++ {
		if (flags & (1 << i)) != 0 {
			cmds = append(cmds, base.SkpFlagSymbols[flags&(1<<i)])
		}
	}

	return fmt.Sprintf("SKP   %s, addr_%d",
		strings.Join(cmds, "|"),
		ip+int(op.Args[1].RawValue)+1)
}

func NOP_ToString(op base.Op) string {
	return fmt.Sprintf("NOP   ")
}

func LDAX_ToString(op base.Op) string {
	return fmt.Sprintf("LDAX  %s",
		base.Symbols[int(op.Args[0].RawValue)])
}

func WRAX_ToString(op base.Op) string {
	regNo := int(op.Args[0].RawValue)
	return fmt.Sprintf("WRAX  %s, %f",
		base.Symbols[regNo],
		utils.QFormatToFloat64(op.Args[2].RawValue, 1, 14))
}

func RDAX_ToString(op base.Op) string {
	return fmt.Sprintf("RDAX  %s, %f",
		base.Symbols[int(op.Args[0].RawValue)],
		utils.QFormatToFloat64(op.Args[2].RawValue, 1, 14))
}

func RDFX_ToString(op base.Op) string {
	return fmt.Sprintf("RDFX  %s, %f",
		base.Symbols[int(op.Args[0].RawValue)],
		utils.QFormatToFloat64(op.Args[2].RawValue, 1, 14))
}

func MULX_ToString(op base.Op) string {
	return fmt.Sprintf("MULX  %s",
		base.Symbols[int(op.Args[0].RawValue)])
}

func WRA_ToString(op base.Op) string {
	return fmt.Sprintf("WRA   %d, %f",
		op.Args[0].RawValue,
		utils.QFormatToFloat64(op.Args[1].RawValue, 1, 9))
}

func WRAP_ToString(op base.Op) string {
	return fmt.Sprintf("WRAP  %d, %f",
		op.Args[0].RawValue,
		utils.QFormatToFloat64(op.Args[1].RawValue, 1, 9))
}

func RDA_ToString(op base.Op) string {
	return fmt.Sprintf("RDA   %d, %f",
		op.Args[0].RawValue,
		utils.QFormatToFloat64(op.Args[1].RawValue, 1, 9))
}

func SOF_ToString(op base.Op) string {
	return fmt.Sprintf("SOF   %f, %f",
		utils.QFormatToFloat64(op.Args[1].RawValue, 1, 14),
		utils.QFormatToFloat64(op.Args[0].RawValue, 0, 10))
}

func EXP_ToString(op base.Op) string {
	return fmt.Sprintf("EXP   %f, %f",
		utils.QFormatToFloat64(op.Args[1].RawValue, 1, 14),
		utils.QFormatToFloat64(op.Args[0].RawValue, 0, 10))
}

func LOG_ToString(op base.Op) string {
	return fmt.Sprintf("LOG   %f, %f",
		utils.QFormatToFloat64(op.Args[1].RawValue, 1, 14),
		utils.QFormatToFloat64(op.Args[0].RawValue, 0, 10))
}

func WLDS_ToString(op base.Op) string {
	amp := int(op.Args[0].RawValue)
	freq := int(op.Args[1].RawValue)
	typ := "SIN0"
	if op.Args[2].RawValue == 1 {
		typ = "SIN1"
	}
	return fmt.Sprintf("WLDS  %s, %d, %d", typ, freq, amp)
}

func WLDR_ToString(op base.Op) string {
	amp := int(base.RampAmpValues[op.Args[0].RawValue])
	freq := int(op.Args[2].RawValue)
	typ := "RMP0"
	if op.Args[3].RawValue == 1 {
		typ = "RMP1"
	}
	return fmt.Sprintf("WLDR  %s, %d, %d", typ, freq, amp)
}

func CHO_ToString(op base.Op) string {
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
		return fmt.Sprintf("CHO   SOF, %s, %s, %d",
			typ, strings.Join(flags, "|"), addr)

	case 0b0:
		return fmt.Sprintf("CHO   RDA, %s, %s, %d",
			typ, strings.Join(flags, "|"), addr)

	case 0b11:
		cmd = "RDAL"
		return fmt.Sprintf("CHO   RDAL, %s", typ)

	default:
		cmd = fmt.Sprintf("<0b%b>", op.Args[4].RawValue)
	}

	return fmt.Sprintf("CHO   %s\t", cmd)
}
