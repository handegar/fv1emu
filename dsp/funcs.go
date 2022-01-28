package dsp

func DecodeOp(opcode uint32) Op {
	opcodeNum := opcode & 0x1F
	opOriginal := Ops[opcodeNum]

	var op Op
	op.Name = opOriginal.Name
	op.RawValue = opcode

	// Copy over all args
	for _, a := range opOriginal.Args {
		op.Args = append(op.Args, a)
	}

	bitPos := 5 // Skip the opcode field
	for i, arg := range op.Args {
		var paramBits uint32 = opcode
		paramBits = (opcode >> bitPos) & ArgBitMasks[arg.Len]
		if arg.Type != Blank {
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
		op.Args[2].Type = Blank
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
		if op.Args[0].RawValue == 0 {
			op.Name = "CLR"
		}
	} else if op.Name == "XOR" {
		if op.Args[0].RawValue == 0xFFFFFFFF {
			op.Name = "NOT"
		}
	} else if op.Name == "MAXX" {
		if op.Args[0].RawValue == 0 && op.Args[2].RawValue == 0 {
			op.Name = "ABSA"
		}
	} else if op.Name == "WLDx" {
		if op.Args[len(op.Args)-1].RawValue == 0 {
			op.Name = "WLDS"
		} else {
			op.Name = "WLDR"
		}
	}

	return op
}

func ParseBuffer(buffer []uint32) []Op {
	var ret []Op
	for _, b := range buffer {
		op := DecodeOp(b)

		if op.Name == "SKP" && op.Args[1].RawValue == 0 {
			break
		}

		ret = append(ret, op)
	}

	return ret
}

func TypeToString(t int) string {
	switch t {
	case Real_1_14:
		return "Real_S1.14"
	case Real_1_9:
		return "Real_S1.9"
	case Real_10:
		return "Real_S.19"
	case Real_4_6:
		return "Real_S4.6"
	case Const:
		return "Const"
	case UInt:
		return "UInt"
	case Int:
		return "Int"
	case Bin:
		return "Bin"
	case Flag:
		return "Flag"
	case Blank:
		return "Blank"
	default:
		return "?"
	}
}
