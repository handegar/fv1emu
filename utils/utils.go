package utils

import "github.com/handegar/fv1emu/base"

// S1.9 or S1.14: -2...1.9999
func Real2ToFloat(bits int, raw uint32) float32 {
	// FIXME: I suspect that this FP->FLOAT operation can be done
	// by simple AND+SHIFT operations -- No logic needed (20220128
	// handegar)
	isSigned := raw >> (bits - 1)
	num := (raw & base.ArgBitMasks[bits-1]) << 1
	ret := (float32(num) / float32(base.ArgBitMasks[bits-1]))
	if isSigned == 1 {
		return -ret
	} else {
		return ret
	}
}

// S.10: -1...0.999999
func Real1ToFloat(bits int, raw uint32) float32 {
	// FIXME: I suspect that this FP->FLOAT operation can be done
	// by simple AND+SHIFT operations -- No logic needed (20220128
	// handegar)
	isSigned := raw >> (bits - 1)
	unSigned := raw & base.ArgBitMasks[bits-1]
	ret := float32(unSigned) / float32(int(1)<<(bits-1))

	if isSigned == 1 {
		return ret - 1.0
	} else {
		return ret
	}
}

// S4.6: –16...15.999998
func Real4ToFloat(bits int, raw uint32) float32 {
	// FIXME: I suspect that this FP->FLOAT operation can be done
	// by simple AND+SHIFT operations -- No logic needed (20220128
	// handegar)
	isSigned := raw >> (bits - 1)
	unSigned := raw & base.ArgBitMasks[bits-1]
	ret := float32(unSigned) / float32(int(1)<<(bits-(1+4)))

	if isSigned == 1 {
		return -ret
	} else {
		return ret
	}
}

func TypeToString(t int) string {
	switch t {
	case base.Real_1_14:
		return "Real_S1.14"
	case base.Real_1_9:
		return "Real_S1.9"
	case base.Real_10:
		return "Real_S.19"
	case base.Real_4_6:
		return "Real_S4.6"
	case base.Const:
		return "Const"
	case base.UInt:
		return "UInt"
	case base.Int:
		return "Int"
	case base.Bin:
		return "Bin"
	case base.Flag:
		return "Flag"
	case base.Blank:
		return "Blank"
	default:
		return "?"
	}
}
