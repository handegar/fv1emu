package dsp

// Bitmasks for the lower X bits
var ArgBitMasks = map[int]uint32{
	32: 0b11111111111111111111111111111111,
	24: 0b111111111111111111111111,
	16: 0b1111111111111111,
	15: 0b111111111111111,
	11: 0b11111111111,
	9:  0b111111111,
	6:  0b111111,
	5:  0b11111,
	2:  0b11,
	1:  0b1,
}

const (
	Real int = iota // S1.14
	UInt
	Int
	Bin
	Flag
	Blank
)

type OpArg struct {
	Len      int // Length of argument (in bits)
	Type     int
	RawValue uint32
}

type Op struct {
	Name     string
	Args     []OpArg
	RawValue uint32
}

var Ops = map[uint32]Op{
	// Accumulator instructions
	0x0B: {"LOG",
		[]OpArg{{16, Real, 0}, {11, Real, 0}}, // C * LOG(|ACC|) + D
		0},
	0x0C: {"EXP",
		[]OpArg{{16, Real, 0}, {11, Real, 0}}, // C * EXP(ACC) + D
		0},
	0x0D: {"SOF",
		[]OpArg{{11, Real, 0}, {16, Real, 0}}, // C * ACC + D
		0},
	0x0E: {"AND",
		[]OpArg{{24, Bin, 0}}, // ACC & MASK  (Also CLR)
		0},
	0x0F: {"OR",
		[]OpArg{{24, Bin, 0}}, // ACC | MASK
		0},
	0x10: {"XOR",
		[]OpArg{{24, Bin, 0}}, // ACC ^ MASK  (Also NOT)
		0},
	0x11: {"SKP",
		[]OpArg{{16, Blank, 0}, {6, UInt, 0}, {5, Flag, 0}}, // Jump according to test
		0},
	// Delay ram instructions
	0x00: {"RDA",
		[]OpArg{{15, UInt, 0}, {1, Blank, 0}, {11, Real, 0}}, // SRAM[ADDR] * C + ACC
		0},
	0x01: {"RMPA",
		[]OpArg{{11, Real, 0}}, // SRAM[PNTR[N]] * C + ACC
		0},
	0x02: {"WRA",
		[]OpArg{{15, UInt, 0}, {11, Real, 0}}, // ACC -> SRAM[ADDR], ACC * C
		0},
	0x03: {"WRAP",
		[]OpArg{{15, UInt, 0}, {11, Real, 0}}, // ACC -> SRAM[ADDR], (ACC*C) + LR
		0},
	// Register instructions
	0x04: {"RDAX",
		[]OpArg{{6, UInt, 0}, {16, Real, 0}}, // C * REG[ADDR] + ACC
		0},
	0x06: {"WRAX",
		[]OpArg{{6, UInt, 0}, {5, Blank, 0}, {16, Real, 0}}, // ACC -> REG[ADDR], C * ACC
		0},
	0x09: {"MAXX",
		[]OpArg{{6, UInt, 0}, {16, Real, 0}}, // MAX(|REG[ADDR] * C|, |ACC|)  (Also AVSA)
		0},
	0x0A: {"MULX",
		[]OpArg{{6, UInt, 0}, {21, Blank, 0}}, // ACC * REG[ADDR]
		0},
	0x05: {"RDFX",
		[]OpArg{{6, UInt, 0}, {5, Blank, 0}, {16, Real, 0}}, // (ACC-REG[ADDR])*C + REG[ADDR]  (Also LDAX)
		0},
	0x08: {"WRLX",
		[]OpArg{{6, UInt, 0}, {16, Real, 0}}, // ACC -> REG[ADDR], (PACC-ACC)*C+PACC
		0},
	0x07: {"WRHX",
		[]OpArg{{6, UInt, 0}, {16, Real, 0}}, // ACC -> REG[ADDR], (ACC*C)+PACC
		0},
	// LFO instructions
	0x12: {"WLDx",
		[]OpArg{{1, Flag, 0}, {9, UInt, 0}, {15, UInt, 0}}, // WLDS and WLDR
		0},
	0x13: {"JAM",
		[]OpArg{{1, Flag, 0}}, // 0 -> RAMP LFO N
		0},
	0x14: {"CHO", // CMD           C             N             ADDR
		[]OpArg{{2, Flag, 0}, {6, Flag, 0}, {2, Flag, 0}, {16, UInt, 0}}, // special
		0},
}

// LDAX = 00000000000000000000000000000101
// RDFX = CCCCCCCCCCCCCCCC00000AAAAAA00101
//RDFX=                         1000000101
//RDFX=                         1010000101
// 0b10000000010000000000000000010001

var Symbols = map[int]string{
	0x00:       "SIN0_RATE",  // (0)  SIN 0 rate
	0x01:       "SIN0_RANGE", // (1)  SIN 0 range
	0x02:       "SIN1_RATE",  // (2)  SIN 1 rate
	0x03:       "SIN1_RANGE", // (3)  SIN 1 range
	0x04:       "RMP0_RATE",  // (4)  RMP 0 rate
	0x05:       "RMP0_RANGE", // (5)  RMP 0 range
	0x06:       "RMP1_RATE",  // (6)  RMP 1 rate
	0x07:       "RMP1_RANGE", // (7)  RMP 1 range
	0x08:       "COMPA",      // (8) USED with 'CHO' instruction: 1's comp address offset (Generate SIN or COS)
	0x10:       "POT0",       // (16)  Pot 0 input register
	0x11:       "POT1",       // (17)  Pot 1 input register
	0x12:       "POT2",       // (18)  Pot 2 input register
	0x14:       "ADCL",       // (20)  ADC input register left channel
	0x15:       "ADCR",       // (21)  ADC input register  right channel
	0x16:       "DACL",       // (22)  DAC output register  left channel
	0x17:       "DACR",       // (23)  DAC output register  right channel
	0x18:       "ADDR_PTR",   // (24)  Used with 'RMPA' instruction for indirect read
	0x20:       "REG0",       // (32)  Register 00
	0x21:       "REG1",       // (33)  Register 01
	0x22:       "REG2",       // (34)  Register 02
	0x23:       "REG3",       // (35)  Register 03
	0x24:       "REG4",       // (36)  Register 04
	0x25:       "REG5",       // (37)  Register 05
	0x26:       "REG6",       // (38)  Register 06
	0x27:       "REG7",       // (39)  Register 07
	0x28:       "REG8",       // (40)  Register 08
	0x29:       "REG9",       // (41)  Register 09
	0x2A:       "REG10",      // (42)  Register 10
	0x2B:       "REG11",      // (43)  Register 11
	0x2C:       "REG12",      // (44)  Register 12
	0x2D:       "REG13",      // (45)  Register 13
	0x2E:       "REG14",      // (46)  Register 14
	0x2F:       "REG15",      // (47)  Register 15
	0x30:       "REG16",      // (48)  Register 16
	0x31:       "REG17",      // (49)  Register 17
	0x32:       "REG18",      // (50)  Register 18
	0x33:       "REG19",      // (51)  Register 19
	0x34:       "REG20",      // (52)  Register 20
	0x35:       "REG21",      // (53)  Register 21
	0x36:       "REG22",      // (54)  Register 22
	0x37:       "REG23",      // (55)  Register 23
	0x38:       "REG24",      // (56)  Register 24
	0x39:       "REG25",      // (57)  Register 25
	0x3A:       "REG26",      // (58)  Register 26
	0x3B:       "REG27",      // (59)  Register 27
	0x3C:       "REG28",      // (60)  Register 28
	0x3D:       "REG29",      // (61)  Register 29
	0x3E:       "REG30",      // (62)  Register 30
	0x3F:       "REG31",      // (63)  Register 31
	0x80000000: "RUN",        // USED with 'SKP' instruction: Skip if NOT FIRST time through program
	0x40000000: "ZRC",        // USED with 'SKP' instruction: Skip On Zero Crossing
	0x20000000: "ZRO",        // USED with 'SKP' instruction: Skip if ACC = 0
	0x10000000: "GEZ",        // USED with 'SKP' instruction: Skip if ACC is '>= 0'
	0x8000000:  "NEG",        // USED with 'SKP' instruction: Skip if ACC is Negative
}

var SymbolEquivalents = map[int][]string{
	0x00: {"SIN0_RATE",
		"SIN0", // (0)  USED with 'CHO' instruction: SINE LFO
		"SIN",  // (0) USED with 'CHO' instruction: SIN/COS from SINE LFO
		"RDA",  // (0) USED with 'CHO' instruction: ACC += (SRAM * COEFF)
	},
	0x01: {"SIN0_RANGE",
		"SIN1", // (1) USED with 'CHO' instruction: SINE LFO 1
		"COS",  // (1) USED with 'CHO' instruction: SIN/COS from SINE LFO
	},
	0x02: {"SIN1_RATE",
		"RMP0", // (2) USED with 'CHO' instruction: RAMP LFO 0
		"SOF",  // (2) USED with 'CHO' instruction: ACC = (ACC * LFO COEFF) + Constant
		"REG",  // (2) USED with 'CHO' instruction: Save LFO temp reg in LFO block
	},
	0x03: {"SIN1_RANGE",
		"RMP1", // (3) USED with 'CHO' instruction: RAMP LFO 1
		"RDAL", // (3) USED with 'CHO' instruction: Reads
		// value of selected LFO into the ACC
	},
	0x04: {"RMP0_RATE",
		"COMPC", // (4) USED with 'CHO' instruction: 2's comp
		// : Generate 1-x for interpolate
	},
	0x10: {"POT0",
		"RPTR2", // (16) USED with 'CHO' instruction: Add 1/2
		// to ramp to generate 2nd ramp for pitch shift
	},
	0x20: {"REG0",
		"NA", // (32) USED with 'CHO' instruction: Do NOT add
		// LFO to address and select cross-fade coefficient
	},
}

var Registers = map[int]string{
	0x00: "SIN0_RATE",
	0x01: "SIN0_RANGE",
	0x02: "SIN1_RATE",
	0x03: "SIN1_RANGE",
	0x04: "RMP0_RATE",
	0x05: "RMP0_RANGE",
	0x06: "RMP1_RATE",
	0x07: "RMP1_RANGE",
	0x10: "POT0",
	0x11: "POT1",
	0x12: "POT2",
	0x14: "ADCL",
	0x15: "ADCR",
	0x16: "DACL",
	0x17: "DACR",
	0x18: "ADDR_PTR",
	0x20: "REG0",
	0x21: "REG1",
	0x22: "REG2",
	0x23: "REG3",
	0x24: "REG4",
	0x25: "REG5",
	0x26: "REG6",
	0x27: "REG7",
	0x28: "REG8",
	0x29: "REG9",
	0x2a: "REG10",
	0x2b: "REG11",
	0x2c: "REG12",
	0x2d: "REG13",
	0x2e: "REG14",
	0x2f: "REG15",
	0x30: "REG16",
	0x31: "REG17",
	0x32: "REG18",
	0x33: "REG19",
	0x34: "REG20",
	0x35: "REG21",
	0x36: "REG22",
	0x37: "REG23",
	0x38: "REG24",
	0x39: "REG25",
	0x3a: "REG26",
	0x3b: "REG27",
	0x3c: "REG28",
	0x3d: "REG29",
	0x3e: "REG30",
	0x3f: "REG31",
}
