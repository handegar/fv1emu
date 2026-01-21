package dsp

import (
	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/settings"
	"github.com/handegar/fv1emu/utils"
)

func CHO_RDAL(op base.Op, state *State) error {
	typ := int(op.Args[1].RawValue)
	lfoValue := GetLFOValue(typ, state, true)

	// NOTE: The debug-flags only apply to SIN0/RMP0, *not* SIN1/RMP1
	if settings.CHO_RDAL_is_NA && !isSinLFO(typ) && typ == base.LFO_RMP0 {
		// Used when debugging the NA envelope
		xfade := GetXFadeFromLFO(lfoValue, typ, state)
		state.ACC.SetFloat64(xfade)

	} else if settings.CHO_RDAL_is_NA_COMPC && !isSinLFO(typ) && typ == base.LFO_RMP0 {
		// Used when debugging the NA envelope
		xfade := GetXFadeFromLFO(lfoValue, typ, state)
		// FIXME: Double check to see if we can COMPC xFade like this. (20220922 handegar)
		xfade = 1.0 - xfade
		state.ACC.SetFloat64(xfade)

	} else if settings.CHO_RDAL_is_RPTR2 && !isSinLFO(typ) && typ == base.LFO_RMP0 {
		// Used when debugging the RPTR2 envelope
		lfoPlusHalf := GetLFOValuePlusHalfCycle(typ, lfoValue)
		utils.Assert(lfoValue != lfoPlusHalf, "Internal RPTR2 error! (%f == %f)", lfoValue, lfoPlusHalf)
		state.ACC.SetFloat64(lfoPlusHalf)

	} else if settings.CHO_RDAL_is_RPTR2_COMPC && !isSinLFO(typ) && typ == base.LFO_RMP0 {
		// Used when debugging the RPTR2 envelope
		lfoPlusHalf := GetLFOValuePlusHalfCycle(typ, lfoValue)
		utils.Assert(lfoValue != lfoPlusHalf, "Internal RPTR2 error!")
		lfoPlusHalf = 1.0 - lfoPlusHalf
		state.ACC.SetFloat64(lfoPlusHalf)

	} else if settings.CHO_RDAL_is_COMPA && (typ == base.LFO_RMP0 || typ == base.LFO_SIN0) {
		// Used when debugging the COMPA envelope
		state.ACC.SetFloat64(-lfoValue)

	} else if settings.CHO_RDAL_is_COMPC && (typ == base.LFO_SIN0 || typ == base.LFO_RMP0) {
		// Used when debugging the COMPC envelope
		lfoValue = 1.0 - lfoValue
		state.ACC.SetFloat64(lfoValue)

	} else if settings.CHO_RDAL_is_COS && typ == base.LFO_SIN0 {
		// Used when debugging the COS envelope
		lfoValue = GetLFOValue(typ+4, state, false)
		state.ACC.SetFloat64(lfoValue)

	} else {
		// Regular CHO RDAL
		state.ACC.SetFloat64(lfoValue)
	}

	return nil
}
