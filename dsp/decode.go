package dsp

import (
	//"fmt"
	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/settings"
)

func DecodeOp(opcode uint32) base.Op {
	opcodeNum := opcode & 0x1F          // Lower 5 bits
	subOpcodeNum := opcode & 0xc0000000 // Upper 2 bits
	opDef := base.Ops[opcodeNum]

	var op base.Op
	op.Name = opDef.Name
	op.RawValue = int32(opcode)

	// Copy over all args
	for _, a := range opDef.Args {
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

	if op.Name == "OR" || op.Name == "AND" || op.Name == "XOR" {
		// HACK: These ops has only one 24 bits parameter which can be signed.
		// FIXME: Clean this up. There has GOT to be a better way to solve
		// this. (20260201 handegar)
		op.Args[1].RawValue = int32(opcode) >> (5 + 3)
	} else {
		// Clear op code
		args := int32(opcode) >> 5
		bitPos := 0
		for i, arg := range op.Args {
			mask := int32((1 << arg.Len) - 1)
			var paramBits = (args >> bitPos) & mask
			if arg.Type != base.Blank {
				op.Args[i].RawValue = paramBits
			} else {
				op.Args[i].RawValue = 0
			}
			bitPos += arg.Len
		}
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
	for n, b := range buffer {
		op := DecodeOp(b)

		if (op.Name == "SKP" || op.Name == "NOP") && op.Args[1].RawValue == 0 {
			break
		}

		ret = append(ret, op)
		if n > settings.InstructionsPerSample { // Fuse
			break
		}
	}

	return ret
}

// Figure out which potensiometers are in use
func PotensiometersInUse(ops []base.Op) (bool, bool, bool) {
	pot0 := false
	pot1 := false
	pot2 := false

	for _, op := range ops {
		if op.Name == "RDAX" || op.Name == "MULX" || op.Name == "LDAX" {
			pot0 = op.Args[0].RawValue == base.POT0 || pot0
			pot1 = op.Args[0].RawValue == base.POT1 || pot1
			pot2 = op.Args[0].RawValue == base.POT2 || pot2
		}
	}

	return pot0, pot1, pot2
}

// Figure out which DACs are in use
func DACsInUse(ops []base.Op) (bool, bool) {
	dacr := false
	dacl := false

	for _, op := range ops {
		if op.Name == "WRAX" {
			dacr = op.Args[0].RawValue == base.DACR || dacr
			dacl = op.Args[0].RawValue == base.DACL || dacl
		}
	}

	return dacr, dacl
}

// Figure out which DACs are in use
func ADCsInUse(ops []base.Op) (bool, bool) {
	adcr := false
	adcl := false

	for _, op := range ops {
		if op.Name == "RDAX" || op.Name == "LDAX" || op.Name == "MULX" {
			adcr = op.Args[0].RawValue == base.ADCR || adcr
			adcl = op.Args[0].RawValue == base.ADCL || adcl
		}
	}

	return adcr, adcl
}
