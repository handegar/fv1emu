package dsp

import (
	"math"

	"github.com/handegar/fv1emu/base"
)

var opTable = map[string]interface{}{
	"LOG": func(op base.Op, state *State) {
		state.workReg1_14.SetWithIntsAndFracs(op.Args[1].RawValue, 1, 14) // C
		// NOTE: According to the SPIN datasheet the second
		// parameter for LOG is supposed to be a S4.6 number,
		// but the ASFY1 assembler, the DISFY1 disassembler,
		// SpinCAD and Igor's Dissasembler2 interpret this as
		// a S.10 float. The SpinCAD source even has a comment
		// which says "SpinASM compatibility", Both seems to work, though.
		state.workReg0_10.SetWithIntsAndFracs(op.Args[0].RawValue, 4, 6) // D

		// C*LOG(|ACC|) + D
		acc := state.ACC.Abs().ToFloat64()
		state.workRegA.SetFloat64(math.Log10(acc) / math.Log10(2.0) / 16.0)
		state.workRegA.Mult(state.workReg1_14)
		state.workRegA.Add(state.workReg0_10)
		state.ACC.Copy(state.workRegA)
	},
	"EXP": func(op base.Op, state *State) {
		state.workReg1_14.SetWithIntsAndFracs(op.Args[1].RawValue, 1, 14) // C
		state.workReg0_10.SetWithIntsAndFracs(op.Args[0].RawValue, 0, 10) // D

		// C*exp(ACC) + D
		acc := state.ACC.ToFloat64()
		state.ACC.SetFloat64(math.Exp2(acc)).Mult(state.workReg1_14).Add(state.workReg0_10)
	},
	"SOF": func(op base.Op, state *State) {
		state.workReg1_14.SetWithIntsAndFracs(op.Args[1].RawValue, 1, 14) // C
		state.workReg0_10.SetWithIntsAndFracs(op.Args[0].RawValue, 0, 10) // D

		// C * ACC + D
		state.workRegA.Copy(state.ACC).Mult(state.workReg1_14).Add(state.workReg0_10)
		state.ACC.Copy(state.workRegA)
	},
	"AND": func(op base.Op, state *State) {
		state.ACC.And(op.Args[1].RawValue)
	},
	"CLR": func(op base.Op, state *State) {
		state.ACC.Clear()
	},
	"OR": func(op base.Op, state *State) {
		state.ACC.Or(op.Args[1].RawValue)
	},
	"XOR": func(op base.Op, state *State) {
		state.ACC.Xor(op.Args[1].RawValue)
	},
	"NOT": func(op base.Op, state *State) {
		state.ACC.Not(op.Args[1].RawValue)
	},
	"SKP": func(op base.Op, state *State) {
		flags := int(op.Args[2].RawValue)
		N := op.Args[1].RawValue
		jmp := false

		if (flags&base.SKP_RUN > 0) && state.RUN_FLAG == true { // RUN
			jmp = true
		}
		if (flags&base.SKP_GEZ) > 0 && !state.ACC.IsSigned() { // GEZ
			jmp = true
		}
		if (flags&base.SKP_ZRO) > 0 && state.ACC.Value == 0 { // ZRO
			jmp = true
		}
		if (flags&base.SKP_ZRC) > 0 &&
			(state.ACC.IsSigned() != state.PACC.IsSigned()) { // ZRC
			jmp = true
		}
		if (flags&base.SKP_NEG) > 0 && state.ACC.IsSigned() { // NEG
			jmp = true
		}

		if jmp {
			state.IP += uint(N)
		}
	},
	"RDA": func(op base.Op, state *State) {
		addr := op.Args[0].RawValue
		state.workReg1_9.SetWithIntsAndFracs(op.Args[1].RawValue, 1, 9) // C
		idx := capDelayRAMIndex(int(addr) + state.DelayRAMPtr)
		delayValue := state.DelayRAM[idx]
		state.LR.SetWithIntsAndFracs(delayValue, 0, 23)

		// SRAM[ADDR] * C + ACC
		state.workRegA.Copy(state.LR).Mult(state.workReg1_9)
		state.ACC.Add(state.workRegA)
	},
	"RMPA": func(op base.Op, state *State) {
		state.workReg1_9.SetWithIntsAndFracs(op.Args[1].RawValue, 1, 9) // C
		addr := state.Registers[base.ADDR_PTR].Value >> 8               // ADDR_PTR
		idx := capDelayRAMIndex(int(addr) + state.DelayRAMPtr)
		delayValue := state.DelayRAM[idx]
		state.LR.SetWithIntsAndFracs(delayValue, 0, 23)

		// SRAM[PNTR[N]] * C + ACC
		state.workRegA.Copy(state.LR).Mult(state.workReg1_9)
		state.ACC.Add(state.workRegA)
	},
	"WRA": func(op base.Op, state *State) {
		addr := op.Args[0].RawValue
		state.workReg1_9.SetWithIntsAndFracs(op.Args[1].RawValue, 1, 9) // C

		// ACC->SRAM[ADDR], ACC * C
		idx := capDelayRAMIndex(int(addr) + state.DelayRAMPtr)
		state.DelayRAM[idx] = state.ACC.ToQFormat(0, 23)
		state.ACC.Mult(state.workReg1_9)
	},
	"WRAP": func(op base.Op, state *State) {
		addr := op.Args[0].RawValue
		state.workReg1_9.SetWithIntsAndFracs(op.Args[1].RawValue, 1, 9) // C

		// ACC->SRAM[ADDR], (ACC*C) + LR
		idx := capDelayRAMIndex(int(addr) + state.DelayRAMPtr)
		// FIXME: Rescale from S7.24 to S0.23. Is this correct? (20220212 handegar)
		state.DelayRAM[idx] = state.ACC.ToQFormat(0, 23)
		state.ACC.Mult(state.workReg1_9).Add(state.LR)
	},
	"RDAX": func(op base.Op, state *State) {
		regNo := int(op.Args[0].RawValue)
		state.workReg1_14.SetWithIntsAndFracs(op.Args[2].RawValue, 1, 14) // C
		reg := state.GetRegister(regNo)

		// (C * REG) + ACC
		state.workRegA.Copy(reg).Mult(state.workReg1_14)
		state.ACC.Add(state.workRegA)
	},
	"WRAX": func(op base.Op, state *State) {
		regNo := int(op.Args[0].RawValue)
		state.workReg1_14.SetWithIntsAndFracs(op.Args[2].RawValue, 1, 14) // C

		// ACC->REG[ADDR], C * ACC
		state.GetRegister(regNo).Copy(state.ACC)
		state.ACC.Mult(state.workReg1_14)
	},
	"MAXX": func(op base.Op, state *State) {
		regNo := int(op.Args[0].RawValue)
		state.workReg1_14.SetWithIntsAndFracs(op.Args[2].RawValue, 1, 14) // C

		// MAX(|REG[ADDR] * C|, |ACC| )
		reg := state.GetRegister(regNo)
		state.workRegA.Copy(reg).Mult(state.workReg1_14)
		state.workRegB.Copy(state.ACC).Abs()
		if state.workRegA.GreaterThan(state.workRegB) {
			state.ACC.Copy(state.workRegA)
		} else {
			state.ACC.Copy(state.workRegB)
		}
	},
	"ABSA": func(op base.Op, state *State) {
		state.ACC.Abs()
	},
	"MULX": func(op base.Op, state *State) {
		regNo := int(op.Args[0].RawValue)
		reg := state.GetRegister(regNo)
		state.ACC.Mult(reg)
	},
	"RDFX": func(op base.Op, state *State) {
		regNo := int(op.Args[0].RawValue)
		state.workReg1_14.SetWithIntsAndFracs(op.Args[2].RawValue, 1, 14) // C
		reg := state.GetRegister(regNo)

		// (ACC - REG)*C + REG
		state.workRegA.Copy(state.ACC)
		state.workRegA.Sub(reg).Mult(state.workReg1_14).Add(reg)
		state.ACC.Copy(state.workRegA)
	},
	"LDAX": func(op base.Op, state *State) {
		regNo := int(op.Args[0].RawValue)
		reg := state.GetRegister(regNo)
		state.ACC.Copy(reg)
	},
	"WRLX": func(op base.Op, state *State) {
		regNo := int(op.Args[0].RawValue)
		state.workReg1_14.SetWithIntsAndFracs(op.Args[2].RawValue, 1, 14) // C

		// ACC->REG[ADDR], (PACC-ACC)*C + PACC
		state.GetRegister(regNo).Copy(state.ACC)
		state.workRegA.Copy(state.PACC).Sub(state.ACC).Mult(state.workReg1_14).Add(state.PACC)
		state.ACC.Copy(state.workRegA)
	},
	"WRHX": func(op base.Op, state *State) {
		regNo := int(op.Args[0].RawValue)
		state.workReg1_14.SetWithIntsAndFracs(op.Args[2].RawValue, 1, 14) // C

		// ACC->REG[ADDR], (ACC*C)+PACC
		state.GetRegister(regNo).Copy(state.ACC)
		state.workRegA.Copy(state.ACC).Mult(state.workReg1_14).Add(state.PACC)
		state.ACC.Copy(state.workRegA)
	},
	"WLDS": func(op base.Op, state *State) {
		freq := op.Args[1].RawValue
		amp := op.Args[0].RawValue

		if op.Args[2].RawValue == 0 { // SIN0
			state.GetRegister(base.SIN0_RATE).SetInt32(freq)
			state.GetRegister(base.SIN0_RANGE).SetInt32(amp)
			state.Sin0State.Angle = 0
		} else { // SIN1
			state.GetRegister(base.SIN1_RATE).SetInt32(freq)
			state.GetRegister(base.SIN1_RANGE).SetInt32(amp)
			state.Sin1State.Angle = 0
		}
	},
	"WLDR": func(op base.Op, state *State) {
		amp := base.RampAmpValues[int(op.Args[0].RawValue)]
		freq := op.Args[2].RawValue
		if op.Args[3].RawValue == 0 { // RAMP0
			state.GetRegister(base.RAMP0_RATE).SetInt32(freq)
			state.GetRegister(base.RAMP0_RANGE).SetInt32(amp)
			state.Ramp0State.Value = 0
		} else { // RAMP1
			state.GetRegister(base.RAMP1_RATE).SetInt32(freq)
			state.GetRegister(base.RAMP1_RANGE).SetInt32(amp)
			state.Ramp1State.Value = 0
		}
	},
	"JAM": func(op base.Op, state *State) {
		num := op.Args[1].RawValue
		if num == 0 {
			state.Ramp0State.Value = 0
		} else {
			state.Ramp1State.Value = 0
		}
	},
	"CHO RDA": func(op base.Op, state *State) {
		addr := int(op.Args[0].RawValue)
		typ := int(op.Args[1].RawValue)
		flags := int(op.Args[3].RawValue)

		if (flags&base.CHO_COS) != 0 && (typ == 0 || typ == 1) {
			typ += 2 // Make SIN -> COS
		}

		lfoVal := GetLFOValue(typ, state)
		var mod int32 = 1
		_ = mod

		if (flags&base.CHO_COMPA) != 0 && (typ == 0 || typ == 1) {
			lfoVal = -lfoVal
		}

		if (flags & base.CHO_REG) != 0 {
			// FIXME: What function does REG have? (20220206 handegar)
		}
		if (flags & base.CHO_RPTR2) != 0 {
			// FIXME: Implement (20220206 handegar)
		}

		if (flags & base.CHO_COMPC) != 0 {
			//lfoVal = GetLFOMaximum(typ, state) - lfoVal
			//max := GetLFOMaximum(typ, state)
			//mod = (max - lfoVal) / max
		}

		if (flags & 0x20) != 0 { // NA
			// FIXME: Handle the NA flag here (20220207 handegar)
		}

		idx := capDelayRAMIndex(addr + state.DelayRAMPtr + int(lfoVal))
		_ = idx
		//state.ACC.SetValue(state.DelayRAM[idx]).Mult24Value(mod)
	},
	"CHO SOF": func(op base.Op, state *State) {
		/*
			D := int(op.Args[0].RawValue)
			typ := int(op.Args[1].RawValue)
			flags := int(op.Args[3].RawValue)
			// FIXME: Implement support for the flags here. (20220131 handegar)
			_ = flags

			lfo := GetLFOValue(typ, state)
			//state.ACC.Mult24Value(int32(lfo * float64(D)))
		*/
	},
	"CHO RDAL": func(op base.Op, state *State) {
		//typ := int(op.Args[1].RawValue)
		//state.ACC.SetValue(int32(GetLFOValue(typ, state)))
	},
}

func applyOp(opCode base.Op, state *State) {
	opTable[opCode.Name].(func(op base.Op, state *State))(opCode, state)
}
