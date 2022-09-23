package dsp

import (
	"errors"

	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/utils"
)

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
		lfo = GetLFOValuePlusHalfCycle(typ, lfo)
	}

	if (flags&base.CHO_COMPA) != 0 && isSinLFO(typ) {
		lfo = -lfo
	} else if (flags&base.CHO_COMPA) != 0 && !isSinLFO(typ) {
		lfoRange := GetRampRange(typ, state) / 4096.0
		lfo = lfoRange - lfo
	}

	if (flags & base.CHO_NA) != 0 { // === Shall we do the X-FADE? =====
		if isSinLFO(typ) {
			return errors.New("Cannot use the NA flag with SIN LFOs")
		}

		xfade := GetXFadeFromLFO(lfo, typ, state)
		if (flags & base.CHO_COMPA) != 0 {
			xfade = -xfade
		}

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
			return state.DebugFlags.IncreaseOutOfBoundsMemoryRead()
		}

		state.workRegA.SetWithIntsAndFracs(state.DelayRAM[idx], 0, 23)

		// Re-get LFO value for interpolation if RPTR2 was used
		// FIXME: This needed? ElmGen does not.. (20220923 handegar)
		/*
			if (flags & base.CHO_RPTR2) != 0 {
				lfo = GetLFOValue(typ, state, (flags&base.CHO_REG) != 0)
			}
		*/

		if (flags & base.CHO_COMPC) != 0 {
			// FIXME: Is this shift needed? (20220923 handegar)
			/*
				if isSinLFO(typ) {
					lfo = (lfo + 1.0) / 2.0 // Shift to [0 .. 1.0]
				}*/
			state.scaleReg.SetFloat64(0.9999 - lfo)
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
