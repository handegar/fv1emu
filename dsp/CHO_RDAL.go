package dsp

import (
	//"fmt"
	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/settings"
	"github.com/handegar/fv1emu/utils"
)

//
// Loads LFO value into ACC
//

func CHO_RDAL(op base.Op, state *State) error {

	// NOTE: Adding flags to "CHO RDAL" is an undocumented feature. Same
	// restrictions as for the other CHO commands. I think the actual chip will
	// only use the REG flag. The other flags are ignored unless specified.

	flags := int(op.Args[3].RawValue)
	typ := int(op.Args[1].RawValue)

	if settings.AllowAllChoRdalFlags && (flags&base.CHO_COS) != 0 {
		utils.Assert(isSinLFO(typ), "Cannot use the COS flag with RAMP LFOs")
		typ += 4 // Make SIN -> COS
	}

	lfo := GetLFOValue(typ, state, (flags&base.CHO_REG) != 0) // Read LFO from internal reg

	if settings.AllowAllChoRdalFlags && (flags&base.CHO_RPTR2 != 0) {
		utils.Assert(!isSinLFO(typ), "Cannot use RPTR2 with SIN LFOs")
		lfo = GetLFOValuePlusHalfCycle(typ, state)
	}

	if settings.AllowAllChoRdalFlags && (flags&base.CHO_COMPA) != 0 {
		if isSinLFO(typ) {
			lfo = -lfo
		} else {
			lfo = 1.0 - lfo
		}
	}

	if settings.AllowAllChoRdalFlags && (flags&base.CHO_NA) != 0 { // ==== Shall we do the X-FADE? ==
		utils.Assert(!isSinLFO(typ), "Cannot use the NA flag with SIN LFOs")

		xfade := GetXFadeFromLFO(lfo, typ, state)
		if (flags & base.CHO_COMPC) != 0 {
			xfade = 1.0 - xfade
		}

		lfo = xfade

	} else if (flags & base.CHO_COMPC) != 0 {
		// Doing "CHO RDAL SINx" with COMPC is not really possible on a real FV-1,
		// so we'll have to limit the output to the 0..1 range.
		lfo = min(1.0, 1.0-lfo)
	}

	state.ACC.SetFloat64(lfo)
	return nil
}
