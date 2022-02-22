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
)

type OpArg struct {
	Len      int // Length of argument (in bits)
	Type     int
	RawValue int32
}

type Op struct {
	Name     string
	Args     []OpArg
	RawValue int32
}
