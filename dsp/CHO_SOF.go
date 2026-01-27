package dsp

import (
	"errors"

	"github.com/handegar/fv1emu/base"
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
		if !isSinLFO(typ) {
			return errors.New("Cannot use the COS flag with RAMP LFOs")
		}
		typ += 4 // Make SIN -> COS
	}

	lfo := GetLFOValue(typ, state, (flags&base.CHO_REG) != 0)

	if (flags&base.CHO_COMPA) != 0 && isSinLFO(typ) {
		lfo = -lfo
	} else if (flags&base.CHO_COMPA) != 0 && !isSinLFO(typ) {
		lfo = 1.0 - lfo
	}

	if (flags & base.CHO_RPTR2) != 0 {
		if isSinLFO(typ) {
			return errors.New("Cannot use RPTR2 with SIN LFOs")
		}
		lfo = GetLFOValuePlusHalfCycle(typ, lfo)
	}

	if (flags & base.CHO_NA) != 0 { // ==== Shall we do the X-FADE? ==
		if isSinLFO(typ) {
			return errors.New("Cannot use the NA flag with SIN LFOs")
		}
		xfade := GetXFadeFromLFO(lfo, typ, state)
		if (flags & base.CHO_COMPC) != 0 {
			xfade = 1.0 - xfade
		}
		if (flags & base.CHO_COMPA) != 0 {
			xfade = -xfade
		}

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
