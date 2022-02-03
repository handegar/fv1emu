package dsp

import (
	"math"

	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/utils"
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

func ProcessSample(opCodes []base.Op, state *State) {
	state.IP = 0
	for state.IP = 0; state.IP < uint(len(opCodes)); {
		op := opCodes[state.IP]
		applyOp(op, state)
		state.IP += 1
	}
	state.RUN_FLAG = false
	state.PACC = state.ACC
}

func GetLFOValue(lfoType int, state *State) float64 {
	lfo := 0.0
	switch lfoType {
	case 0:
		lfo = math.Sin(state.Sin0State.Angle) * float64(state.Registers[base.SIN0_RANGE].(int))
	case 1:
		lfo = math.Sin(state.Sin1State.Angle) * float64(state.Registers[base.SIN1_RANGE].(int))
	case 2:
		lfo = state.Ramp0State.Value * float64(state.Registers[base.RAMP0_RANGE].(int))
	case 3:
		lfo = state.Ramp1State.Value * float64(state.Registers[base.RAMP1_RANGE].(int))
	}
	return lfo
}

var opTable = map[string]interface{}{
	"LOG": func(op base.Op, state *State) {
		C := float64(utils.Real2ToFloat(op.Args[1].Len, op.Args[1].RawValue))
		D := float64(utils.Real1ToFloat(op.Args[0].Len, op.Args[0].RawValue))
		state.ACC = float32(C*math.Log2(math.Abs(float64(state.ACC))) + D)
	},
	"EXP": func(op base.Op, state *State) {
		C := float64(utils.Real2ToFloat(op.Args[1].Len, op.Args[1].RawValue))
		D := float64(utils.Real1ToFloat(op.Args[0].Len, op.Args[0].RawValue))
		state.ACC = float32(C*math.Exp2(float64(state.ACC)) + D)
	},
	"SOF": func(op base.Op, state *State) {
		C := utils.Real2ToFloat(op.Args[1].Len, op.Args[1].RawValue)
		D := utils.Real1ToFloat(op.Args[0].Len, op.Args[0].RawValue)
		state.ACC = (C * state.ACC) + D
	},
	"AND": func(op base.Op, state *State) {
		state.ACC = math.Float32frombits(math.Float32bits(state.ACC) & op.Args[0].RawValue)
	},
	"CLR": func(op base.Op, state *State) {
		state.ACC = 0.0
	},
	"OR": func(op base.Op, state *State) {
		state.ACC = math.Float32frombits(math.Float32bits(state.ACC) | op.Args[0].RawValue)
	},
	"XOR": func(op base.Op, state *State) {
		state.ACC = math.Float32frombits(math.Float32bits(state.ACC) ^ op.Args[0].RawValue)
	},
	"NOT": func(op base.Op, state *State) {
		state.ACC = math.Float32frombits(math.Float32bits(state.ACC) &^ op.Args[0].RawValue)
	},
	"SKP": func(op base.Op, state *State) {
		flags := int(op.Args[2].RawValue)
		N := op.Args[1].RawValue
		jmp := false
		// FIXME: All flags must be fulfilled before
		// jumping. (20220131 handegar)
		if (flags&0b10000 > 0) && state.RUN_FLAG { // RUN
			jmp = true
		}
		if (flags&0b00010) > 0 && state.ACC >= 0 { // GEZ
			jmp = true
		}
		if (flags&0b00100) > 0 && state.ACC == 0 { // ZRO
			jmp = true
		}
		if (flags&0b01000) > 0 &&
			math.Signbit(float64(state.ACC)) != math.Signbit(float64(state.PACC)) { // ZRC
			jmp = true
		}
		if (flags&0b00001) > 0 && state.ACC < 0 { // NEG
			jmp = true
		}

		if jmp {
			state.IP += uint(N)
		}
	},
	"RDA": func(op base.Op, state *State) {
		addr := op.Args[0].RawValue
		C := utils.Real2ToFloat(op.Args[1].Len, op.Args[1].RawValue)
		state.LR = state.DelayRAM[addr]
		/*
			fmt.Printf("RDA %f*%f + %f = %f\n",
				state.DelayRAM[addr], C, state.ACC,
				state.DelayRAM[addr]*C+state.ACC)
		*/
		state.ACC = state.DelayRAM[addr]*C + state.ACC
	},
	"RMPA": func(op base.Op, state *State) {
		C := utils.Real2ToFloat(op.Args[1].Len, op.Args[1].RawValue)
		addr := state.Registers[base.ADDR_PTR].(int) // ADDR_PTR
		state.LR = state.DelayRAM[addr]
		state.ACC = state.DelayRAM[addr]*C + state.ACC
	},
	"WRA": func(op base.Op, state *State) {
		addr := op.Args[0].RawValue
		C := utils.Real2ToFloat(op.Args[1].Len, op.Args[1].RawValue)

		//fmt.Printf("WRA acc=%f, C=%f, dram=%f\n", state.ACC*C, C, state.ACC)

		state.DelayRAM[addr] = state.ACC
		state.ACC = state.ACC * C

	},
	"WRAP": func(op base.Op, state *State) {
		addr := op.Args[0].RawValue
		C := utils.Real2ToFloat(op.Args[1].Len, op.Args[1].RawValue)
		/*
			fmt.Printf("ACC=%f, %f*%f+%f=%f\n", state.ACC,
				state.ACC, C, state.LR,
				state.ACC*C+state.LR)
		*/
		state.DelayRAM[addr] = state.ACC
		state.ACC = state.ACC*C + state.LR
	},
	"RDAX": func(op base.Op, state *State) {
		regNo := int(op.Args[0].RawValue)
		C := utils.Real2ToFloat(op.Args[2].Len, op.Args[2].RawValue)
		regValue := state.Registers[regNo].(float64)
		//fmt.Printf("RDAX %f*%f + %f = %f (reg 0x%x)\n", C, regValue, state.ACC, C*float32(regValue)+state.ACC, regNo)
		state.ACC = C*float32(regValue) + state.ACC
	},
	"WRAX": func(op base.Op, state *State) {
		regNo := int(op.Args[0].RawValue)
		C := utils.Real2ToFloat(op.Args[2].Len, op.Args[2].RawValue)

		//fmt.Printf("WRAX 0x%x = %f*%f = %f\n", regNo, state.ACC, C, state.ACC*C)
		state.Registers[regNo] = float64(state.ACC)
		state.ACC = state.ACC * C

	},
	"MAXX": func(op base.Op, state *State) {
		regNo := int(op.Args[0].RawValue)
		C := utils.Real2ToFloat(op.Args[2].Len, op.Args[2].RawValue)
		regValue := state.Registers[regNo].(float64)
		state.ACC = float32(math.Max(math.Abs(regValue*float64(C)), float64(state.ACC)))
	},
	"ABSA": func(op base.Op, state *State) {
		state.ACC = float32(math.Abs(float64(state.ACC)))
	},
	"MULX": func(op base.Op, state *State) {
		regNo := int(op.Args[0].RawValue)
		regValue := state.Registers[regNo].(float64)
		state.ACC = state.ACC * float32(regValue)
	},
	"RDFX": func(op base.Op, state *State) {
		regNo := int(op.Args[0].RawValue)
		C := utils.Real2ToFloat(op.Args[2].Len, op.Args[2].RawValue)
		regValue := state.Registers[regNo].(float64)
		state.ACC = (state.ACC-float32(regValue))*C + float32(regValue)
	},
	"LDAX": func(op base.Op, state *State) {
		regNo := int(op.Args[0].RawValue)
		state.ACC = float32(state.Registers[regNo].(float64))
	},
	"WRLX": func(op base.Op, state *State) {
		regNo := int(op.Args[0].RawValue)
		C := utils.Real2ToFloat(op.Args[2].Len, op.Args[2].RawValue)
		state.Registers[regNo] = float64(state.ACC)
		state.ACC = (state.PACC-state.ACC)*C + state.PACC
	},
	"WRHX": func(op base.Op, state *State) {
		regNo := int(op.Args[0].RawValue)
		C := utils.Real2ToFloat(op.Args[2].Len, op.Args[2].RawValue)
		state.Registers[regNo] = float64(state.ACC)
		state.ACC = state.ACC*C + state.PACC
	},
	"WLDS": func(op base.Op, state *State) {
		amp := int(op.Args[0].RawValue)
		freq := int(op.Args[1].RawValue)
		if op.Args[2].RawValue == 0 { // SIN0
			state.Registers[base.SIN0_RATE] = freq
			state.Registers[base.SIN0_RANGE] = amp
		} else { // SIN1
			state.Registers[base.SIN1_RATE] = freq
			state.Registers[base.SIN1_RANGE] = amp
		}
	},
	"WLDR": func(op base.Op, state *State) {
		amp := int(base.RampAmpValues[int(op.Args[0].RawValue)])
		freq := int(op.Args[2].RawValue)
		if op.Args[3].RawValue == 0 { // RAMP0
			state.Registers[base.RAMP0_RATE] = freq
			state.Registers[base.RAMP0_RANGE] = amp
		} else { // RAMP1
			state.Registers[base.RAMP1_RATE] = freq
			state.Registers[base.RAMP1_RANGE] = amp
		}
	},
	"JAM": func(op base.Op, state *State) {
		num := int(op.Args[1].RawValue)
		if num == 0 {
			state.Ramp0State.Value = 0
		} else {
			state.Ramp1State.Value = 0
		}
	},
	"CHO RDA": func(op base.Op, state *State) {
		addr := (int(op.Args[0].RawValue) << 1) >> 1
		typ := int(op.Args[1].RawValue)
		flags := int(op.Args[3].RawValue)
		// FIXME: Implement support for the flags here. (20220131 handegar)
		_ = flags

		// FIXME: This var shall also be modulated (according to flags)
		// (20220202 handegar)
		var coefficient float32 = 1.0

		offset := GetLFOValue(typ, state)
		state.ACC = state.ACC + state.DelayRAM[addr+int(offset)]*coefficient
	},
	"CHO SOF": func(op base.Op, state *State) {
		D := (int(op.Args[0].RawValue) << 1) >> 1
		typ := int(op.Args[1].RawValue)
		flags := int(op.Args[3].RawValue)
		// FIXME: Implement support for the flags here. (20220131 handegar)
		_ = flags

		lfo := GetLFOValue(typ, state)
		state.ACC = float32(float64(state.ACC) * lfo * float64(D))
	},
	"CHO RDAL": func(op base.Op, state *State) {
		typ := int(op.Args[1].RawValue)
		state.ACC = float32(GetLFOValue(typ, state))
	},
}

func applyOp(opCode base.Op, state *State) {
	opTable[opCode.Name].(func(op base.Op, state *State))(opCode, state)
}
