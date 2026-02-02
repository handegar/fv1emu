package base

const (
	Real_1_14 int = iota // S1.14
	Real_1_9
	Real_10
	Real_4_6
	UInt
	Int
	Bin
	Flag
	Const
	Blank
)

// Common register names
const (
	SIN0_RATE   = 0
	SIN0_RANGE  = 1
	SIN1_RATE   = 2
	SIN1_RANGE  = 3
	RAMP0_RATE  = 4
	RAMP0_RANGE = 5
	RAMP1_RATE  = 6
	RAMP1_RANGE = 7

	POT0 = 0x10
	POT1 = 0x11
	POT2 = 0x12

	ADCL     = 0x14
	ADCR     = 0x15
	DACL     = 0x16
	DACR     = 0x17
	ADDR_PTR = 0x18

	REG0 = 0x20
)

const (
	SKP_NEG = 0x1
	SKP_GEZ = 0x2
	SKP_ZRO = 0x4
	SKP_ZRC = 0x8
	SKP_RUN = 0x10
)

const (
	CHO_SIN   = 0x0
	CHO_COS   = 0x1
	CHO_REG   = 0x2
	CHO_COMPC = 0x4
	CHO_COMPA = 0x8
	CHO_RPTR2 = 0x10
	CHO_NA    = 0x20
)

const (
	LFO_SIN0 = 0
	LFO_SIN1 = 1
	LFO_RMP0 = 2
	LFO_RMP1 = 3
	LFO_COS0 = 4
	LFO_COS1 = 5
)

// FIXME: Maybe we should open for increasing this for wild experiments?
// (20260119 handegar)
const MEMORY_SIZE = 32768
