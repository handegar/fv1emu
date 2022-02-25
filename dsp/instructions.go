package dsp

import (
	"errors"
	"fmt"
	"math"

	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/settings"
)

func isSinLFO(typ int) bool {
	return typ == base.LFO_SIN0 || typ == base.LFO_SIN1 ||
		typ == base.LFO_COS0 || typ == base.LFO_COS1
}

func assert(mustBeTrue bool, msg string) {
	if !mustBeTrue {
		fmt.Printf("ERROR: %s\n", msg)
		panic("ASSERT failed")
	}
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
		state.workReg0_10.SetWithIntsAndFracs(op.Args[0].RawValue, 0, 10) //4, 6) // D

		// C*LOG(|ACC|) + D
		acc := state.ACC.Abs().ToFloat64()
		if acc <= 0.0 {
			acc = 1.0 / (1 << 23)
		}

		state.workRegA.SetFloat64(math.Log10(acc) / math.Log10(2.0) / 16.0)
		state.workRegA.Mult(state.workReg1_14)
		state.workRegA.Add(state.workReg0_10)
		state.ACC.Copy(state.workRegA)
		return nil
	},
	"EXP": func(op base.Op, state *State) error {
		state.workReg1_14.SetWithIntsAndFracs(op.Args[1].RawValue, 1, 14) // C
		state.workReg0_10.SetWithIntsAndFracs(op.Args[0].RawValue, 0, 10) // D

		// C*exp(ACC) + D
		acc := state.ACC.ToFloat64()
		if acc >= 0.0 {
			state.ACC.SetFloat64(0.9999998807907104).Mult(state.workReg1_14).Add(state.workReg0_10)
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
		state.workRegA.Copy(state.ACC).Mult(state.workReg1_14).Add(state.workReg0_10)
		state.ACC.Copy(state.workRegA)
		return nil
	},
	"AND": func(op base.Op, state *State) error {
		state.ACC.And(op.Args[1].RawValue)
		return nil
	},
	"CLR": func(op base.Op, state *State) error {
		state.ACC.Clear()
		return nil
	},
	"OR": func(op base.Op, state *State) error {
		state.ACC.Or(op.Args[1].RawValue)
		return nil
	},
	"XOR": func(op base.Op, state *State) error {
		state.ACC.Xor(op.Args[1].RawValue)
		return nil
	},
	"NOT": func(op base.Op, state *State) error {
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
		state.workRegA.Copy(state.LR).Mult(state.workReg1_9)
		state.ACC.Add(state.workRegA)
		return nil
	},
	"RMPA": func(op base.Op, state *State) error {
		state.workReg1_9.SetWithIntsAndFracs(op.Args[1].RawValue, 1, 9) // C
		addr := state.Registers[base.ADDR_PTR].Value >> 8               // ADDR_PTR
		idx, err := capDelayRAMIndex(int(addr)+state.DelayRAMPtr, state)
		if err != nil {
			return state.DebugFlags.IncreaseOutOfBoundsMemoryRead()
		}

		delayValue := state.DelayRAM[idx]
		state.LR.SetWithIntsAndFracs(delayValue, 0, 23)

		// SRAM[PNTR[N]] * C + ACC
		state.workRegA.Copy(state.LR).Mult(state.workReg1_9)
		state.ACC.Add(state.workRegA)
		return nil
	},
	"WRA": func(op base.Op, state *State) error {
		addr := op.Args[0].RawValue
		state.workReg1_9.SetWithIntsAndFracs(op.Args[1].RawValue, 1, 9) // C

		// ACC->SRAM[ADDR], ACC * C
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
		idx, err := capDelayRAMIndex(int(addr)+state.DelayRAMPtr, state)
		if err != nil {
			return state.DebugFlags.IncreaseOutOfBoundsMemoryWrite()
		}

		// FIXME: Rescale from S7.24 to S0.23. Is this correct? (20220212 handegar)
		state.DelayRAM[idx] = state.ACC.ToQFormat(0, 23)
		state.ACC.Mult(state.workReg1_9).Add(state.LR)
		return nil
	},
	"RDAX": func(op base.Op, state *State) error {
		regNo := int(op.Args[0].RawValue)
		state.workReg1_14.SetWithIntsAndFracs(op.Args[2].RawValue, 1, 14) // C
		reg := state.GetRegister(regNo)

		// (C * REG) + ACC
		state.workRegA.Copy(reg).Mult(state.workReg1_14)
		state.ACC.Add(state.workRegA)
		return nil
	},
	"WRAX": func(op base.Op, state *State) error {
		regNo := int(op.Args[0].RawValue)
		state.workReg1_14.SetWithIntsAndFracs(op.Args[2].RawValue, 1, 14) // C

		// ACC->REG[ADDR], C * ACC
		reg := state.GetRegister(regNo)
		if regNo == base.RAMP0_RANGE || regNo == base.RAMP1_RANGE { // Special case?
			reg.SetInt32(state.ACC.ToInt32() >> 14)
		} else if regNo == base.RAMP0_RATE || regNo == base.RAMP1_RATE {
			reg.SetInt32(state.ACC.ToInt32() >> 10)
		} else { // Just a regular WRAX
			reg.Copy(state.ACC)
		}

		state.ACC.Mult(state.workReg1_14)
		return nil
	},
	"MAXX": func(op base.Op, state *State) error {
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
		return nil
	},
	"ABSA": func(op base.Op, state *State) error {
		state.ACC.Abs()
		return nil
	},
	"MULX": func(op base.Op, state *State) error {
		regNo := int(op.Args[0].RawValue)
		reg := state.GetRegister(regNo)
		state.ACC.Mult(reg)
		return nil
	},
	"RDFX": func(op base.Op, state *State) error {
		regNo := int(op.Args[0].RawValue)
		state.workReg1_14.SetWithIntsAndFracs(op.Args[2].RawValue, 1, 14) // C
		reg := state.GetRegister(regNo)

		// (ACC - REG)*C + REG
		state.workRegA.Copy(state.ACC)
		state.workRegA.Sub(reg).Mult(state.workReg1_14).Add(reg)
		state.ACC.Copy(state.workRegA)
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
		state.GetRegister(regNo).Copy(state.ACC)
		state.workRegA.Copy(state.PACC).Sub(state.ACC).Mult(state.workReg1_14).Add(state.PACC)
		state.ACC.Copy(state.workRegA)
		return nil
	},
	"WRHX": func(op base.Op, state *State) error {
		regNo := int(op.Args[0].RawValue)
		state.workReg1_14.SetWithIntsAndFracs(op.Args[2].RawValue, 1, 14) // C

		// ACC->REG[ADDR], (ACC*C)+PACC
		state.GetRegister(regNo).Copy(state.ACC)
		state.workRegA.Copy(state.ACC).Mult(state.workReg1_14).Add(state.PACC)
		state.ACC.Copy(state.workRegA)
		return nil
	},
	"WLDS": func(op base.Op, state *State) error {
		freq := op.Args[1].RawValue // Really "1/freq"
		amp := op.Args[0].RawValue

		// Cap values
		if freq < 0 {
			freq = 0
			state.DebugFlags.SetSinLFOFlag(op.Args[2].RawValue)
		} else if freq > 511 {
			freq = 511
			state.DebugFlags.SetSinLFOFlag(op.Args[2].RawValue)
		}

		if amp < 0 {
			amp = 0
			state.DebugFlags.SetSinLFOFlag(op.Args[2].RawValue)
		} else if amp > (1 << 15) { // 32768
			amp = (1 << 15) - 1
			state.DebugFlags.SetSinLFOFlag(op.Args[2].RawValue)
		}

		if op.Args[2].RawValue == 0 { // SIN0
			state.GetRegister(base.SIN0_RATE).SetInt32(freq)
			state.GetRegister(base.SIN0_RANGE).SetInt32(amp)
			state.Sin0State.Angle = 0.0
		} else { // SIN1
			state.GetRegister(base.SIN1_RATE).SetInt32(freq)
			state.GetRegister(base.SIN1_RANGE).SetInt32(amp)
			state.Sin1State.Angle = 0.0
		}
		return nil
	},
	"WLDR": func(op base.Op, state *State) error {
		amp := base.RampAmpValues[op.Args[0].RawValue]
		_, valid := base.RampAmpValuesMap[amp]
		if !valid {
			state.DebugFlags.SetRampLFOFlag(op.Args[3].RawValue)
			amp = 4096
			msg := fmt.Sprintf("Ramp amplitude value (%d) is not 512, 1024, 2048 or 4096", amp)
			return errors.New(msg)
		}

		freq := op.Args[2].RawValue
		// cap frequency value
		if freq < -(1 << 14) { // -16384
			freq = -(1 << 14) + 1
			state.DebugFlags.SetRampLFOFlag(op.Args[3].RawValue)
		} else if freq > (1 << 15) { // 32768
			freq = (1 << 15) - 1
			state.DebugFlags.SetRampLFOFlag(op.Args[3].RawValue)
		}

		if op.Args[3].RawValue == 0 { // RAMP0
			state.GetRegister(base.RAMP0_RATE).SetInt32(freq)
			state.GetRegister(base.RAMP0_RANGE).SetInt32(amp)
			state.Ramp0State.Value = 0
		} else { // RAMP1
			state.GetRegister(base.RAMP1_RATE).SetInt32(freq)
			state.GetRegister(base.RAMP1_RANGE).SetInt32(amp)
			state.Ramp1State.Value = 0
		}
		return nil
	},
	"JAM": func(op base.Op, state *State) error {
		num := op.Args[1].RawValue
		if num == 0 {
			state.Ramp0State.Value = 0
		} else {
			state.Ramp1State.Value = 0
		}
		return nil
	},
	"CHO RDA": func(op base.Op, state *State) error {
		addr := int(op.Args[0].RawValue)
		typ := int(op.Args[1].RawValue)
		flags := int(op.Args[3].RawValue)

		if (flags & base.CHO_COS) != 0 {
			if !isSinLFO(typ) {
				return errors.New("Cannot use the COS flag with RAMP LFOs")
			}
			typ += 4 // Make SIN -> COS
		}

		lfo := GetLFOValue(typ, state, (flags&base.CHO_REG) == 0)
		if (flags & base.CHO_COMPA) != 0 {
			lfo = -lfo
		}

		if (flags & base.CHO_RPTR2) != 0 {
			if isSinLFO(typ) {
				return errors.New("Cannot use RPTR2 with SIN LFOs")
			}
			lfo = GetLFOValuePlusHalfCycle(typ, state)
		}

		scaledLFO := ScaleLFOValue(lfo, typ, state)

		if (flags & base.CHO_NA) != 0 { // Shall we do the X-FADE?
			if isSinLFO(typ) {
				return errors.New("Cannot use the NA flag with SIN LFOs")
			}

			xfade := GetXFadeFromLFO(lfo, typ, state)
			if (flags & base.CHO_COMPC) != 0 {
				xfade = 1.0 - xfade
			}
			state.workRegB.SetFloat64(xfade)

			delayIndex := addr
			idx, err := capDelayRAMIndex(state.DelayRAMPtr+delayIndex, state)
			if err != nil {
				return state.DebugFlags.IncreaseOutOfBoundsMemoryRead()
			}
			state.workRegA.SetWithIntsAndFracs(state.DelayRAM[idx], 0, 23)

			state.workRegA.Mult(state.workRegB) // delayram*xfade
			state.ACC.Add(state.workRegA)
		} else {
			delayIndex := addr + int(scaledLFO)
			idx, err := capDelayRAMIndex(state.DelayRAMPtr+delayIndex, state)
			if err != nil {
				return state.DebugFlags.IncreaseOutOfBoundsMemoryRead()
			}

			state.workRegA.SetWithIntsAndFracs(state.DelayRAM[idx], 0, 23)

			interpolate := lfo
			if isSinLFO(typ) {
				// LFO is [-1 .. 1]
				interpolate = (lfo + 1.0) / 2.0 // get 0...1.0
				assert(interpolate >= 0.0 && interpolate <= 1.0,
					"interpolate is < 0 || > 1.0")
			}

			if (flags & base.CHO_COMPC) != 0 {
				state.workRegB.SetFloat64(1.0 - interpolate)
			} else {
				state.workRegB.SetFloat64(interpolate)
			}

			state.workRegA.Mult(state.workRegB)
			state.ACC.Add(state.workRegA)
		}

		return nil
	},
	"CHO SOF": func(op base.Op, state *State) error {
		D := int32(op.Args[0].RawValue)
		typ := int(op.Args[1].RawValue)
		flags := int(op.Args[3].RawValue)

		if (flags & base.CHO_COS) != 0 {
			if !isSinLFO(typ) {
				return errors.New("Cannot use the COS flag with RAMP LFOs")
			}
			typ += 4 // Make SIN -> COS
		}

		lfo := GetLFOValue(typ, state, (flags&base.CHO_REG) == 0)
		if (flags & base.CHO_COMPA) != 0 {
			lfo = -lfo
		}

		if (flags & base.CHO_RPTR2) != 0 {
			if isSinLFO(typ) {
				return errors.New("Cannot use RPTR2 with SIN LFOs")
			}
			lfo = GetLFOValuePlusHalfCycle(typ, state)
		}

		if (flags & base.CHO_NA) != 0 { // Shall we do the X-FADE?
			if isSinLFO(typ) {
				return errors.New("Cannot use the NA flag with SIN LFOs")
			}
			xfade := GetXFadeFromLFO(lfo, typ, state)
			if (flags & base.CHO_COMPC) != 0 {
				xfade = 1.0 - xfade
			}
			scaledXFade := ScaleLFOValue(xfade, typ, state)
			state.workRegA.SetFloat64(scaledXFade)
		} else {
			scaledLFO := ScaleLFOValue(lfo, typ, state)
			state.workRegA.SetFloat64(scaledLFO)
		}

		state.workRegB.SetWithIntsAndFracs(D, 0, 15)
		state.ACC.Mult(state.workRegA).Add(state.workRegB)
		return nil
	},
	"CHO RDAL": func(op base.Op, state *State) error {
		typ := int(op.Args[1].RawValue)
		lfo := GetLFOValue(typ, state, false)

		if settings.CHO_RDAL_is_NA && !isSinLFO(typ) && typ == base.LFO_RMP0 {
			// Used when debugging the NA envelope
			xfade := GetXFadeFromLFO(lfo, typ, state)
			scaledXFade := ScaleLFOValue(xfade, typ, state)
			state.ACC.SetFloat64(scaledXFade)
		} else if settings.CHO_RDAL_is_RPTR2 && !isSinLFO(typ) && typ == base.LFO_RMP0 {
			// Used when debugging the RPTR2 envelope
			lfo = GetLFOValuePlusHalfCycle(typ, state)
			lfoScaled := ScaleLFOValue(lfo, typ, state)
			state.ACC.SetFloat64(lfoScaled)
		} else if settings.CHO_RDAL_is_COMPA && (typ == base.LFO_RMP0 || typ == base.LFO_SIN0) {
			// Used when debugging the COMPA envelope
			lfoScaled := -ScaleLFOValue(lfo, typ, state)
			state.ACC.SetFloat64(lfoScaled)
		} else if settings.CHO_RDAL_is_COS && typ == base.LFO_SIN0 {
			// Used when debugging the COS envelope
			lfo = GetLFOValue(typ+4, state, false)
			lfoScaled := ScaleLFOValue(lfo, typ+4, state)
			state.ACC.SetFloat64(lfoScaled)
		} else {
			lfoScaled := ScaleLFOValue(lfo, typ, state)
			state.ACC.SetFloat64(lfoScaled)
		}
		return nil
	},
}

func applyOp(opCode base.Op, state *State) error {
	err := opTable[opCode.Name].(func(op base.Op, state *State) error)(opCode, state)
	return err
}
