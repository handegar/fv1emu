package dsp

import (
	//"fmt"
	"math"

	"github.com/handegar/fv1emu/base"
	//"github.com/handegar/fv1emu/utils"
)

func isSinLFO(typ int) bool {
	return typ == base.LFO_SIN0 || typ == base.LFO_SIN1 ||
		typ == base.LFO_COS0 || typ == base.LFO_COS1
}

var opTable = map[string]interface{}{
	"LOG": func(op base.Op, state *State) error {
		state.workReg1_14.SetWithIntsAndFracs(op.Args[1].RawValue, 1, 14) // C
		// NOTE: According to the SPIN datasheet the second
		// parameter for LOG is supposed to be a S4.6 number,
		// but the ASFY1 assembler, the DISFY1 disassembler,
		// SpinCAD and Igor's Dissasembler2 interpret this as
		// a S.10 float. The SpinCAD source even has a comment
		// which says "SpinASM compatibility", Both seems to work, though.
		// UPDATE: According to this post the manual might
		// contain an error:
		// http://www.spinsemi.com/forum/viewtopic.php?f=3&t=511
		state.workReg0_10.SetWithIntsAndFracs(op.Args[0].RawValue, 0, 10) // D

		// C*LOG(|ACC|) + D
		state.PACC.Copy(state.ACC)
		acc := state.ACC.Abs().ToFloat64()

		val := (math.Log10(acc) / math.Log10(2.0)) / 16.0
		val = val * state.workReg1_14.ToFloat64()
		val = val + state.workReg0_10.ToFloat64()
		state.ACC.SetClampedFloat64(val)

		return nil
	},
	"EXP": func(op base.Op, state *State) error {
		state.workReg1_14.SetWithIntsAndFracs(op.Args[1].RawValue, 1, 14) // C
		state.workReg0_10.SetWithIntsAndFracs(op.Args[0].RawValue, 0, 10) // D

		// C*exp(ACC) + D
		state.PACC.Copy(state.ACC)
		acc := state.ACC.ToFloat64()
		if acc >= 0 {
			state.ACC.SetToMax24Bit().Mult(state.workReg1_14).Add(state.workReg0_10)
		} else {
			acc = acc * 16.0
			state.ACC.SetFloat64(math.Exp2(acc)).Mult(state.workReg1_14).Add(state.workReg0_10)
		}
		return nil
	},
	"SOF": func(op base.Op, state *State) error {
		state.workReg1_14.SetWithIntsAndFracs(op.Args[1].RawValue, 1, 14) // C
		state.workReg0_10.SetWithIntsAndFracs(op.Args[0].RawValue, 0, 10) // D

		// C * ACC + D
		state.PACC.Copy(state.ACC)
		state.ACC.Mult(state.workReg1_14).Add(state.workReg0_10)
		return nil
	},
	"AND": func(op base.Op, state *State) error {
		state.PACC.Copy(state.ACC)
		state.ACC.And(op.Args[1].RawValue)
		return nil
	},
	"CLR": func(op base.Op, state *State) error {
		state.PACC.Copy(state.ACC)
		state.ACC.Clear()
		return nil
	},
	"OR": func(op base.Op, state *State) error {
		// FIXME: Shall we not update PACC? (20220305 handegar)
		state.ACC.Or(op.Args[1].RawValue)
		return nil
	},
	"XOR": func(op base.Op, state *State) error {
		state.PACC.Copy(state.ACC)
		state.ACC.Xor(op.Args[1].RawValue)
		return nil
	},
	"NOT": func(op base.Op, state *State) error {
		state.PACC.Copy(state.ACC)
		state.ACC.Not(op.Args[1].RawValue)
		return nil
	},
	"SKP": func(op base.Op, state *State) error {
		flags := int(op.Args[2].RawValue)
		N := op.Args[1].RawValue

		jmp := false
		if (flags&base.SKP_RUN > 0) && state.RUN_FLAG == true { // RUN
			jmp = true
		}
		if (flags&base.SKP_GEZ) > 0 && !state.ACC.IsSigned() && state.ACC.ToInt32() >= 0 { // GEZ
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
		return nil
	},
	"RDA": func(op base.Op, state *State) error {
		addr := op.Args[0].RawValue
		state.workReg1_9.SetWithIntsAndFracs(op.Args[1].RawValue, 1, 9) // C
		idx, err := capDelayRAMIndex(int(addr)+state.DelayRAMPtr, state)
		if err != nil {
			return state.DebugFlags.IncreaseOutOfBoundsMemoryRead()
		}

		delayValue := state.DelayRAM[idx]
		state.LR.SetWithIntsAndFracs(delayValue, 0, 23)

		// SRAM[ADDR] * C + ACC
		state.PACC.Copy(state.ACC)
		state.workRegA.Copy(state.LR).Mult(state.workReg1_9)
		state.ACC.Add(state.workRegA)
		return nil
	},
	"RMPA": func(op base.Op, state *State) error {
		state.workReg1_9.SetWithIntsAndFracs(op.Args[1].RawValue, 1, 9) // C
		addr := state.GetRegister(base.ADDR_PTR).ToInt32() >> 8         // ADDR_PTR
		idx, err := capDelayRAMIndex(int(addr)+state.DelayRAMPtr, state)
		if err != nil {
			return state.DebugFlags.IncreaseOutOfBoundsMemoryRead()
		}

		delayValue := state.DelayRAM[idx]
		state.LR.SetWithIntsAndFracs(delayValue, 0, 23)

		// SRAM[PNTR[N]] * C + ACC
		state.PACC.Copy(state.ACC)
		state.workRegA.Copy(state.LR).Mult(state.workReg1_9)
		state.ACC.Add(state.workRegA)
		return nil
	},
	"WRA": func(op base.Op, state *State) error {
		addr := op.Args[0].RawValue
		state.workReg1_9.SetWithIntsAndFracs(op.Args[1].RawValue, 1, 9) // C

		// ACC->SRAM[ADDR], ACC * C
		state.PACC.Copy(state.ACC)
		idx, err := capDelayRAMIndex(int(addr)+state.DelayRAMPtr, state)
		if err != nil {
			return state.DebugFlags.IncreaseOutOfBoundsMemoryWrite()
		}

		state.DelayRAM[idx] = state.ACC.ToQFormat(0, 23)
		state.ACC.Mult(state.workReg1_9)
		return nil
	},
	"WRAP": func(op base.Op, state *State) error {
		addr := op.Args[0].RawValue
		state.workReg1_9.SetWithIntsAndFracs(op.Args[1].RawValue, 1, 9) // C

		// ACC->SRAM[ADDR], (ACC*C) + LR
		state.PACC.Copy(state.ACC)
		idx, err := capDelayRAMIndex(int(addr)+state.DelayRAMPtr, state)
		if err != nil {
			return state.DebugFlags.IncreaseOutOfBoundsMemoryWrite()
		}

		state.DelayRAM[idx] = state.ACC.ToQFormat(0, 23)
		state.ACC.Mult(state.workReg1_9).Add(state.LR)
		return nil
	},
	"RDAX": func(op base.Op, state *State) error {
		regNo := int(op.Args[0].RawValue)
		state.workReg1_14.SetWithIntsAndFracs(op.Args[2].RawValue, 1, 14) // C
		reg := state.GetRegister(regNo)

		// (C * REG) + ACC
		state.PACC.Copy(state.ACC)
		state.workRegA.Copy(reg).Mult(state.workReg1_14)
		state.ACC.Add(state.workRegA)
		return nil
	},
	"RDFX": func(op base.Op, state *State) error {
		regNo := int(op.Args[0].RawValue)
		reg := state.GetRegister(regNo)
		state.workReg1_14.SetWithIntsAndFracs(op.Args[2].RawValue, 1, 14) // C

		// (ACC - REG)*C + REG
		state.PACC.Copy(state.ACC)
		state.workRegA.Copy(state.ACC)
		state.workRegA.Sub(reg).Mult(state.workReg1_14).Add(reg)
		state.ACC.Copy(state.workRegA)
		return nil
	},
	"WRAX": func(op base.Op, state *State) error {
		regNo := int(op.Args[0].RawValue)
		state.workReg1_14.SetWithIntsAndFracs(op.Args[2].RawValue, 1, 14) // C

		// ACC->REG[ADDR], C * ACC
		state.PACC.Copy(state.ACC)
		reg := state.GetRegister(regNo)

		// Special handling for LFO registers and ADDR_PTR
		accAsInt := state.ACC.ToInt32()

		switch regNo {
		case base.SIN0_RATE, base.SIN1_RATE:
			// FIXME: Allow negative freqs for sine? (20220923 handegar)
			if accAsInt < 0 { // Don't allow a negative rate/freq
				accAsInt = -accAsInt
			}
			rate := accAsInt >> (24 - 9 - 1)
			reg.SetInt32(rate)
			if regNo == base.SIN0_RATE {
				state.Sin0Osc.SetFreq(rate)
			} else {
				state.Sin1Osc.SetFreq(rate)
			}

		case base.SIN0_RANGE, base.SIN1_RANGE:
			amp := accAsInt >> (24 - 15 - 1)
			reg.SetInt32(amp)
			if regNo == base.SIN0_RATE {
				state.Sin0Osc.SetAmp(amp)
			} else {
				state.Sin1Osc.SetAmp(amp)
			}

		case base.RAMP0_RATE, base.RAMP1_RATE:
			freq := int32(int16(accAsInt >> (24 - 16)))
			reg.SetInt32(freq)
			if regNo == base.RAMP0_RATE {
				state.Ramp0Osc.SetFreq(freq)
			} else {
				state.Ramp1Osc.SetFreq(freq)
			}

		case base.RAMP0_RANGE, base.RAMP1_RANGE:
			amp := state.ACC.ToInt32()
			reg.SetInt32(amp)
			//utils.Assert(false, "FIXME: How do we use WRAX with RAMP-range?")

			// FIXME: Do some bit-tricks here to extract the bits determining the
			// range. Two highest bits? (20260125 handegar)
			if regNo == base.RAMP0_RANGE {
				state.Ramp0Osc.SetAmp(state.ACC.ToFloat64())
			} else {
				state.Ramp1Osc.SetAmp(state.ACC.ToFloat64())
			}

		case base.ADDR_PTR:
			// This value will be down-scaled when used as
			// we can only address 32768 memory bytes
			reg.SetInt32(accAsInt)

		default:
			// Just a regular WRAX with ACC as a floating point value
			reg.Copy(state.ACC)
		}

		state.ACC.Mult(state.workReg1_14)
		return nil
	},
	"MAXX": func(op base.Op, state *State) error {
		regNo := int(op.Args[0].RawValue)
		state.workReg1_14.SetWithIntsAndFracs(op.Args[2].RawValue, 1, 14) // C

		// MAX(|REG[ADDR] * C|, |ACC| )
		state.PACC.Copy(state.ACC)
		reg := state.GetRegister(regNo)
		state.workRegA.Copy(reg).Mult(state.workReg1_14).Abs()
		state.workRegB.Copy(state.ACC).Abs()
		if state.workRegA.GreaterThan(state.workRegB) {
			state.ACC.Copy(state.workRegA)
		} else {
			state.ACC.Copy(state.workRegB)
		}
		return nil
	},
	"ABSA": func(op base.Op, state *State) error {
		state.PACC.Copy(state.ACC)
		state.ACC.Abs()
		return nil
	},
	"MULX": func(op base.Op, state *State) error {
		regNo := int(op.Args[0].RawValue)
		reg := state.GetRegister(regNo)

		// ACC * REG[ADDR]
		state.PACC.Copy(state.ACC)
		state.ACC.Mult(reg)
		return nil
	},
	"LDAX": func(op base.Op, state *State) error {
		regNo := int(op.Args[0].RawValue)
		reg := state.GetRegister(regNo)
		state.ACC.Copy(reg)
		return nil
	},
	"WRLX": func(op base.Op, state *State) error {
		regNo := int(op.Args[0].RawValue)
		state.workReg1_14.SetWithIntsAndFracs(op.Args[2].RawValue, 1, 14) // C

		// ACC->REG[ADDR], (PACC-ACC)*C + PACC
		state.workRegB.Copy(state.ACC)
		state.GetRegister(regNo).Copy(state.ACC)
		state.workRegA.Copy(state.PACC).Sub(state.ACC).Mult(state.workReg1_14).Add(state.PACC)
		state.ACC.Copy(state.workRegA)
		state.PACC.Copy(state.workRegB)
		return nil
	},
	"WRHX": func(op base.Op, state *State) error {
		regNo := int(op.Args[0].RawValue)
		state.workReg1_14.SetWithIntsAndFracs(op.Args[2].RawValue, 1, 14) // C

		// ACC->REG[ADDR], (ACC*C)+PACC
		state.workRegB.Copy(state.ACC)
		state.GetRegister(regNo).Copy(state.ACC)
		state.workRegA.Copy(state.ACC).Mult(state.workReg1_14).Add(state.PACC)
		state.ACC.Copy(state.workRegA)
		state.PACC.Copy(state.workRegB)
		return nil
	},
	"WLDS": func(op base.Op, state *State) error {
		freq := op.Args[1].RawValue
		amp := op.Args[0].RawValue
		typ := op.Args[2].RawValue

		// Cap values
		if freq < 0 {
			freq = 0
			state.DebugFlags.SetInvalidSinLFOFlag(typ)
		} else if freq > ((1 << 9) - 1) { // > 511
			freq = (1 << 9) - 1
			state.DebugFlags.SetInvalidSinLFOFlag(typ)
		}

		if amp < 0 {
			amp = 0
			state.DebugFlags.SetInvalidSinLFOFlag(typ)
		} else if amp > ((1 << 15) - 1) { // > 32768
			amp = (1 << 15) - 1
			state.DebugFlags.SetInvalidSinLFOFlag(typ)
		}

		if typ == 0 { // SIN0
			state.Registers[base.SIN0_RATE].SetInt32(freq)
			state.Registers[base.SIN0_RANGE].SetInt32(int32(amp))
			state.Sin0Osc.SetFreq(freq)
			state.Sin0Osc.SetAmp(amp)
		} else { // SIN1
			state.Registers[base.SIN1_RATE].SetInt32(freq)
			state.Registers[base.SIN1_RANGE].SetInt32(int32(amp))
			state.Sin1Osc.SetFreq(freq)
			state.Sin1Osc.SetAmp(amp)
		}

		return nil
	},
	"WLDR": func(op base.Op, state *State) error {
		ampIdx := int8(op.Args[0].RawValue)
		freq := int32(int16(op.Args[2].RawValue)) // Aka. 'rate'
		typ := op.Args[3].RawValue

		// Cap frequency value
		if freq < -(1 << 14) { // -16384
			freq = -(1 << 14) + 1
			state.DebugFlags.SetInvalidRampLFOFlag(typ)
		} else if freq > ((1 << 15) - 1) { // +32768
			freq = (1 << 15) - 1
			state.DebugFlags.SetInvalidRampLFOFlag(typ)
		}

		if typ == 0 { // RAMP0
			state.Registers[base.RAMP0_RATE].SetInt32(freq)
			state.Registers[base.RAMP0_RANGE].SetInt32(int32(base.RampAmpValues[int32(ampIdx)]))
			state.Ramp0Osc.SetFreq(freq)
			state.Ramp0Osc.SetAmpIdx(ampIdx)
		} else { // RAMP1
			state.Registers[base.RAMP1_RATE].SetInt32(freq)
			state.Registers[base.RAMP1_RANGE].SetInt32(int32(base.RampAmpValues[int32(ampIdx)]))
			state.Ramp1Osc.SetFreq(freq)
			state.Ramp1Osc.SetAmpIdx(ampIdx)
		}

		return nil
	},
	"JAM": func(op base.Op, state *State) error {
		typ := op.Args[0].RawValue
		if typ == 0 {
			state.Ramp0Osc.Reset()
		} else {
			state.Ramp1Osc.Reset()
		}
		return nil
	},
	"CHO RDA":  CHO_RDA,
	"CHO SOF":  CHO_SOF,
	"CHO RDAL": CHO_RDAL,
}

// Returns the LFO value and resulting scale value depending on the flags.
func CHO(typ int, flags int, state *State) (float64, float64) {
	lfo := GetLFOValue(typ, state, (flags&base.CHO_REG) != 0)
	scale := lfo

	if isSinLFO(typ) { // Sine ==============================
		if (flags & base.CHO_COMPC) != 0 {
			scale = 1.0 - lfo
		} else {
			scale = lfo
		}

		if (flags & base.CHO_COMPA) != 0 {
			lfo = -lfo
		}

	} else { // Ramp ========================================
		if (flags & base.CHO_COMPA) != 0 {
			lfo = 1.0 - lfo
		}

		if (flags & base.CHO_NA) != 0 {
			lfo = GetXFadeFromLFO(lfo, typ, state)
		}

		if (flags & base.CHO_COMPC) != 0 {
			scale = 1.0 - lfo
		} else {
			scale = lfo
		}
	}

	return lfo, scale
}

func applyOp(opCode base.Op, state *State) error {
	err := opTable[opCode.Name].(func(op base.Op, state *State) error)(opCode, state)
	return err
}
