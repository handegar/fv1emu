package dsp

import (
	"math"

	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/utils"
)

var opTable = map[string]interface{}{
	"LOG": func(op base.Op, state *State) {
		C := float64(utils.QFormatToFloat64(op.Args[1].RawValue, 1, 14))
		// NOTE: According to the SPIN datasheet the second
		// parameter for LOG is supposed to be a S4.6 number,
		// but the ASFY1 assembler, the DISFY1 disassembler,
		// SpinCAD and Igor's Dissasembler2 interpret this as
		// a S.10 float. The SpinCAD source even has a comment
		// which says "SpinASM compatibility", so I assume the
		// data-sheet is wrong.
		// FIXME: Test agains the official SpinASM
		// assembler. (20220205 handegar)
		D := float64(utils.QFormatToFloat64(op.Args[0].RawValue, 0, 10))

		//state.ACC = float32(C*math.Log2(math.Abs(float64(state.ACC))) + D)

		// This is how SpinCAD does it:
		x := 16.0 * 2.0 // NOTE: Just 16.0 seems to caus
		state.ACC = float32(C*(math.Log10(float64(state.ACC))/math.Log10(2.0)/x) + D)

	},
	"EXP": func(op base.Op, state *State) {
		C := float64(utils.QFormatToFloat64(op.Args[1].RawValue, 1, 14))
		D := float64(utils.QFormatToFloat64(op.Args[0].RawValue, 0, 10))
		state.ACC = float32(C*math.Exp2(float64(state.ACC)) + D)
	},
	"SOF": func(op base.Op, state *State) {
		C := utils.QFormatToFloat64(op.Args[1].RawValue, 1, 14)
		D := utils.QFormatToFloat64(op.Args[0].RawValue, 0, 10)
		state.ACC = float32((C * float64(state.ACC)) + D)
	},
	"AND": func(op base.Op, state *State) {
		ACC_bits := utils.Float64ToQFormat(float64(state.ACC), 1, 23)
		state.ACC = float32(utils.QFormatToFloat64(ACC_bits&op.Args[1].RawValue, 1, 23))
	},
	"CLR": func(op base.Op, state *State) {
		state.ACC = 0.0
	},
	"OR": func(op base.Op, state *State) {
		ACC_bits := utils.Float64ToQFormat(float64(state.ACC), 1, 23)
		state.ACC = float32(utils.QFormatToFloat64(ACC_bits|op.Args[1].RawValue, 1, 23))
	},
	"XOR": func(op base.Op, state *State) {
		ACC_bits := utils.Float64ToQFormat(float64(state.ACC), 1, 23)
		state.ACC = float32(utils.QFormatToFloat64(ACC_bits^op.Args[1].RawValue, 1, 23))
	},
	"NOT": func(op base.Op, state *State) {
		ACC_bits := utils.Float64ToQFormat(float64(state.ACC), 1, 23)
		state.ACC = float32(utils.QFormatToFloat64(ACC_bits&^op.Args[1].RawValue, 1, 23))
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
		C := utils.QFormatToFloat64(op.Args[1].RawValue, 1, 9)
		idx := capDelayRAMIndex(int(addr) + state.DelayRAMPtr)
		delayValue := state.DelayRAM[idx]
		state.LR = delayValue
		state.ACC += delayValue * float32(C)
	},
	"RMPA": func(op base.Op, state *State) {
		C := utils.QFormatToFloat64(op.Args[1].RawValue, 1, 9)
		addr := state.Registers[base.ADDR_PTR].(int) >> 8 // ADDR_PTR
		idx := capDelayRAMIndex(int(addr) + state.DelayRAMPtr)
		delayValue := state.DelayRAM[idx]
		state.LR = delayValue
		state.ACC += delayValue * float32(C)
	},
	"WRA": func(op base.Op, state *State) {
		addr := op.Args[0].RawValue
		C := utils.QFormatToFloat64(op.Args[1].RawValue, 1, 9)
		idx := capDelayRAMIndex(int(addr) + state.DelayRAMPtr)
		state.DelayRAM[idx] = state.ACC
		state.ACC = state.ACC * float32(C)
	},
	"WRAP": func(op base.Op, state *State) {
		addr := op.Args[0].RawValue
		C := utils.QFormatToFloat64(op.Args[1].RawValue, 1, 9)
		idx := capDelayRAMIndex(int(addr) + state.DelayRAMPtr)
		state.DelayRAM[idx] = state.ACC
		state.ACC = state.ACC*float32(C) + state.LR
	},
	"RDAX": func(op base.Op, state *State) {
		regNo := int(op.Args[0].RawValue)
		C := utils.QFormatToFloat64(op.Args[2].RawValue, 1, 14)
		regValue := state.Registers[regNo].(float64)
		state.ACC += float32(regValue * C)
	},
	"WRAX": func(op base.Op, state *State) {
		regNo := int(op.Args[0].RawValue)
		C := utils.QFormatToFloat64(op.Args[2].RawValue, 1, 14)
		if regNo == base.ADDR_PTR {
			ACC_bits := utils.Float64ToQFormat(float64(state.ACC), 1, 23)
			//fmt.Printf("WRITE ADDR_PTR %f/%d/[%24b]\n", state.ACC, int(ACC_bits), int(ACC_bits))
			state.Registers[regNo] = int(ACC_bits)
		} else {
			state.Registers[regNo] = float64(state.ACC)
		}
		state.ACC = state.ACC * float32(C)
	},
	"MAXX": func(op base.Op, state *State) {
		regNo := int(op.Args[0].RawValue)
		C := utils.QFormatToFloat64(op.Args[2].RawValue, 1, 14)
		regValue := state.Registers[regNo].(float64)
		state.ACC = float32(math.Max(math.Abs(regValue*float64(C)), float64(state.ACC)))
	},
	"ABSA": func(op base.Op, state *State) {
		state.ACC = float32(math.Abs(float64(state.ACC)))
	},
	"MULX": func(op base.Op, state *State) {
		regNo := int(op.Args[0].RawValue)
		regValue := 0.0
		if regNo == base.ADDR_PTR {
			regValue = float64(state.Registers[regNo].(int))
		} else {
			regValue = state.Registers[regNo].(float64)
		}

		state.ACC = state.ACC * float32(regValue)
	},
	"RDFX": func(op base.Op, state *State) {
		regNo := int(op.Args[0].RawValue)
		C := utils.QFormatToFloat64(op.Args[2].RawValue, 1, 14)
		regValue := 0.0
		if regNo == base.ADDR_PTR {
			regValue = float64(state.Registers[regNo].(int))
		} else {
			regValue = state.Registers[regNo].(float64)
		}
		state.ACC = float32((float64(state.ACC)-regValue)*C + regValue)
	},
	"LDAX": func(op base.Op, state *State) {
		regNo := int(op.Args[0].RawValue)
		if regNo == base.ADDR_PTR {
			state.ACC = float32(state.Registers[regNo].(int))
		} else {
			state.ACC = float32(state.Registers[regNo].(float64))
		}
	},
	"WRLX": func(op base.Op, state *State) {
		regNo := int(op.Args[0].RawValue)
		C := utils.QFormatToFloat64(op.Args[2].RawValue, 1, 14)
		state.Registers[regNo] = float64(state.ACC)
		state.ACC = (state.PACC-state.ACC)*float32(C) + state.PACC
	},
	"WRHX": func(op base.Op, state *State) {
		regNo := int(op.Args[0].RawValue)
		C := utils.QFormatToFloat64(op.Args[2].RawValue, 1, 14)
		state.Registers[regNo] = float64(state.ACC)
		state.ACC = state.ACC*float32(C) + state.PACC
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
		idx := capDelayRAMIndex(addr + state.DelayRAMPtr + int(offset))
		state.ACC = state.ACC + state.DelayRAM[idx]*coefficient
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
