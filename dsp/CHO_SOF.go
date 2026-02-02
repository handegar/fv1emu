package dsp

import (
	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/utils"
)

//
// Multiplies ACC with an LFO value and adds a constant D
//

func CHO_SOF(op base.Op, state *State) error {
	D := int32(op.Args[0].RawValue)
	typ := int(op.Args[1].RawValue)
	flags := int(op.Args[3].RawValue)

	state.offsetReg.SetWithIntsAndFracs(D, 0, 15)

	if (flags & base.CHO_COS) != 0 {
		utils.Assert(isSinLFO(typ), "Cannot use the COS flag with RAMP LFOs")
		typ += 4 // Make SIN -> COS
	}

	lfo := GetLFOValue(typ, state, (flags&base.CHO_REG) != 0)

	if flags&base.CHO_COMPA != 0 {
		if isSinLFO(typ) {
			lfo = -lfo
		} else {
			lfo = 1.0 - lfo
		}
	}

	if (flags & base.CHO_RPTR2) != 0 {
		utils.Assert(!isSinLFO(typ), "Cannot use RPTR2 with SIN LFOs")
		lfo = GetLFOValuePlusHalfCycle(typ, state)
	}

	if (flags & base.CHO_NA) != 0 { // ==== Shall we do the X-FADE? ==
		utils.Assert(!isSinLFO(typ), "Cannot use the NA flag with SIN LFOs")

		// XFade before COMPC
		xfade := GetXFadeFromLFO(lfo, typ, state)
		if (flags & base.CHO_COMPC) != 0 {
			xfade = 1.0 - xfade
		}
		state.scaleReg.SetFloat64(xfade)

		/*
			   // COMPC before XFade
			if (flags & base.CHO_COMPC) != 0 {
				lfo = 1.0 - lfo
			}
			xfade := GetXFadeFromLFO(lfo, typ, state)
		*/

		state.scaleReg.SetFloat64(xfade)

	} else { // =================================  Regular envelope ==
		if (flags & base.CHO_COMPC) != 0 {
			lfo = 1.0 - lfo
		}
		state.scaleReg.SetFloat64(lfo)
	}

	state.ACC.Mult(state.scaleReg).Add(state.offsetReg)
	return nil
}
