package dsp

import (
	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/utils"
)

//
// Loads LFO value into ACC
//

func CHO_RDAL(op base.Op, state *State) error {

	// NOTE: Adding flags to "CHO RDAL" is an undocumented feature. Same
	// restrictions as for the other CHO commands.
	flags := int(op.Args[3].RawValue)
	typ := int(op.Args[1].RawValue)
	
	if (flags & base.CHO_COS) != 0 {
		utils.Assert(isSinLFO(typ), "Cannot use the COS flag with RAMP LFOs")
		typ += 4 // Make SIN -> COS
	}


	lfo := GetLFOValue(typ, state, (flags&base.CHO_REG) != 0) // Read LFO from internal reg


	if (flags & base.CHO_RPTR2) != 0 {
		utils.Assert(!isSinLFO(typ), "Cannot use RPTR2 with SIN LFOs")
		lfo = GetLFOValuePlusHalfCycle(typ, state)
	}

	if flags&base.CHO_COMPA != 0 {
		if isSinLFO(typ) {
			lfo = -lfo
		} else {
			lfo = 1.0 - lfo
		}
	}


	if (flags & base.CHO_NA) != 0 { // ==== Shall we do the X-FADE? ==
		utils.Assert(!isSinLFO(typ), "Cannot use the NA flag with SIN LFOs")
		
		xfade := GetXFadeFromLFO(lfo, typ, state)
		if (flags & base.CHO_COMPC) != 0 {
			xfade = 1.0 - xfade
		}

		lfo = xfade
	}
	
	state.ACC.SetFloat64(lfo)
	return nil
}
