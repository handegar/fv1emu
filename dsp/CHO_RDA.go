package dsp

import (
	"errors"
	"fmt"

	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/utils"
)

/**
  From the datasheet:

  Like the RDA instruction, CHO RDA will read a sample from the delay ram,
multiply it by a coefficient and add the product to the previous content of
ACC. However, in contrast to RDA the coefficient is not explicitly embedded
within the instruction and the effective delay ram address is not solely
determined by the address parameter. Instead, both values are modulated by the
selected LFO at run time, for an in depth explanation please consult the FV-1
datasheet alongside with application note AN-0001. CHO RDA is a very flexible
and powerful instruction, especially useful for delay line modulation effects
such as chorus or pitch shifting.
*/

/*
Flags:

	 Sine LFO only:
	    SIN   - Sine output
	    COS   - Cosine output
	 Ramp LFO only:
		  NA    - Use xfade as coeff. Do not add adress offset
			RPTR2 - Value from 1/2 phase ahead
	 Both LFO types:
		  COMPA - Complement the offset address
		  COMPC - Use (1-LFO) as value
		  REG   - Save LFO value to a register for reuse
*/

func CHO_RDA(op base.Op, state *State) error {
	/*
	   LFO related post:
	   http://www.spinsemi.com/forum/viewtopic.php?f=3&t=505
	*/

	addr := int(op.Args[0].RawValue)
	typ := int(op.Args[1].RawValue)
	flags := int(op.Args[3].RawValue)

	if (flags & base.CHO_COS) != 0 {
		if !isSinLFO(typ) {
			return errors.New("Cannot use the COS flag with RAMP LFOs")
		}
		typ += 4 // Make SIN -> COS
	}

	lfo := GetLFOValue(typ, state, (flags&base.CHO_REG) != 0)

	if (flags & base.CHO_RPTR2) != 0 {
		if isSinLFO(typ) {
			return errors.New("Cannot use RPTR2 with SIN LFOs")
		}
		lfo = GetLFOValuePlusHalfCycle(typ, state)
	}

	if flags&base.CHO_COMPA != 0 {
		if isSinLFO(typ) {
			lfo = -lfo
		} else {
			lfo = GetRampRange(typ, state) - lfo
		}
	}

	if (flags & base.CHO_NA) != 0 { // === Shall we do the X-FADE? =====
		if isSinLFO(typ) {
			return errors.New("Cannot use the NA flag with SIN LFOs")
		}

		xfade := GetXFadeFromLFO(lfo, typ, state)
		if (flags & base.CHO_COMPC) != 0 {
			xfade = 1.0 - xfade
		}
		state.scaleReg.SetFloat64(xfade)
		state.offsetReg.SetInt32(int32(addr))

		state.ACC.Add(state.offsetReg.Mult(state.scaleReg))
	} else { // == Regular LFO envelope ================================
		scaledLFO := ScaleLFOValue(lfo, typ, state)
		delayIndex := addr + int(scaledLFO)

		idx, err := capDelayRAMIndex(state.DelayRAMPtr+delayIndex, state)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			utils.Assert(false, "Mem access out of bounds")
			return state.DebugFlags.IncreaseOutOfBoundsMemoryRead()
		}

		delayValue := state.DelayRAM[idx]
		state.LR.SetWithIntsAndFracs(delayValue, 0, 23)
		state.workRegA.SetWithIntsAndFracs(delayValue, 0, 23)

		if (flags & base.CHO_COMPC) != 0 {
			// FIXME: Is this shift needed? (20220923 handegar)
			if isSinLFO(typ) {
				lfo = (lfo + 1.0) / 2.0 // Shift to [0 .. 1.0]
			}
			state.scaleReg.SetFloat64(1.0 - lfo)
			lfo = (2*lfo - 1.0)
		} else {
			state.scaleReg.SetFloat64(lfo)
		}

		utils.Assert(lfo >= -1.0 && lfo <= 1.0,
			"LFO is < 0 || > 1.0 (was %f, type=%s)", lfo, base.LFOTypeNames[typ])

		state.workRegA.Mult(state.scaleReg)
		state.ACC.Add(state.workRegA)
	}

	return nil
}
