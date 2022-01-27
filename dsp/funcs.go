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

	//fmt.Printf("Name=%s (0b%32b)\n", op.Name, opcode)
	bitPos := 5 // Skip the opcode field
	for i, arg := range op.Args {
		var paramBits uint32 = opcode
		paramBits = (opcode >> bitPos) & ArgBitMasks[arg.Len]
		//fmt.Printf("Len=%d, Type=%s, RawValue=%b, Mask=0b%b (>> %d) \n", arg.Len, TypeToString(arg.Type), paramBits, ArgBitMasks[arg.Len], bitPos)
		if arg.Type != Blank {
			//newArg := OpArg{arg.Len, arg.Type, paramBits}
			//op.Args = append(op.Args, newArg)
			op.Args[i].RawValue = paramBits
		}
		bitPos += arg.Len
	}

	// Special cases
	if op.Name == "RDFX" && op.Args[1].RawValue == 0 && op.Args[2].RawValue == 0 {
		op.Name = "LDAX"
		// Ignore first arg
		op.Args[2].Type = Blank
	}

	//fmt.Printf(" -> %d args\n", len(op.Args))
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
	case Real:
		return "Real"
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
