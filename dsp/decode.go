package dsp

import (
	"github.com/handegar/fv1emu/base"
)

func DecodeOp(opcode uint32) base.Op {
	opcodeNum := opcode & 0x1F          // Lower 5 bits
	subOpcodeNum := opcode & 0xc0000000 // Upper 2 bits
	opOriginal := base.Ops[opcodeNum]

	var op base.Op
	op.Name = opOriginal.Name
	op.RawValue = opcode

	// Copy over all args
	for _, a := range opOriginal.Args {
		op.Args = append(op.Args, a)
	}

	// Special case for WLDR -> WLDS
	if opcodeNum == 0x12 && subOpcodeNum == 0 {
		op.Name = "WLDS"
		op.Args = []base.OpArg{
			{Len: 15, Type: base.UInt, RawValue: 0},
			{Len: 9, Type: base.UInt, RawValue: 0},
			{Len: 1, Type: base.Flag, RawValue: 0},
			{Len: 2, Type: base.Const, RawValue: 0}}
	}

	bitPos := 5 // Skip the opcode field
	for i, arg := range op.Args {
		var paramBits uint32 = opcode
		paramBits = (opcode >> bitPos) & base.ArgBitMasks[arg.Len]
		if arg.Type != base.Blank {
			op.Args[i].RawValue = paramBits
		}
		bitPos += arg.Len
	}

	//
	// Special cases
	//
	if op.Name == "RDFX" && op.Args[1].RawValue == 0 && op.Args[2].RawValue == 0 {
		op.Name = "LDAX"
		// Ignore last arg
		op.Args[2].Type = base.Blank
	} else if op.Name == "CHO" {
		switch op.Args[len(op.Args)-1].RawValue {
		case 0b0:
			op.Name = "CHO RDA"
		case 0b10:
			op.Name = "CHO SOF"
		case 0b11:
			op.Name = "CHO RDAL"
		default:
			op.Name = "CHO <?>"
		}
	} else if op.Name == "AND" {
		if op.Args[1].RawValue == 0 {
			op.Name = "CLR"
		}
	} else if op.Name == "XOR" && op.Args[1].RawValue == 0xFFFFF8 {
		op.Name = "NOT"
	} else if op.Name == "MAXX" {
		if op.Args[0].RawValue == 0 && op.Args[2].RawValue == 0 {
			op.Name = "ABSA"
		}
	} else if op.Name == "SKP" && op.Args[1].RawValue == 0 && op.Args[2].RawValue == 0 {
		op.Name = "NOP" // Undocumented but used by SpinASM
	}

	return op
}

func DecodeOpCodes(buffer []uint32) []base.Op {
	var ret []base.Op
	for _, b := range buffer {
		op := DecodeOp(b)

		if (op.Name == "SKP" || op.Name == "NOP") && op.Args[1].RawValue == 0 {
			break
		}

		ret = append(ret, op)
	}

	return ret
}
