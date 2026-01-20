package base

var Ops = map[uint32]Op{
	0x0B: {"LOG",
		[]OpArg{{11, Real_4_6, 0}, {16, Real_1_14, 0}},
		0},
	0x0C: {"EXP",
		[]OpArg{{11, Real_10, 0}, {16, Real_1_14, 0}},
		0},
	0x0D: {"SOF",
		[]OpArg{{11, Real_10, 0}, {16, Real_1_14, 0}},
		0},
	0x0E: {"AND", // Also CLR if arg=0
		[]OpArg{{4, Blank, 0}, {24, Bin, 0}},
		0},
	0x0F: {"OR",
		[]OpArg{{4, Blank, 0}, {24, Bin, 0}},
		0},

	0x10: {"XOR", // Also NOT if arg=0xFFFFF8
		[]OpArg{{4, Blank, 0}, {24, Bin, 0}},
		0},
	0x11: {"SKP", // Also NOP if all args are 0
		[]OpArg{{16, Blank, 0}, {6, UInt, 0}, {5, Flag, 0}},
		0},
	0x00: {"RDA",
		[]OpArg{{16, UInt, 0}, {11, Real_1_9, 0}},
		0},
	0x01: {"RMPA",
		[]OpArg{{16, Const, 24}, {11, Real_1_9, 0}},
		0},
	0x02: {"WRA",
		[]OpArg{{16, UInt, 0}, {11, Real_1_9, 0}},
		0},
	0x03: {"WRAP",
		[]OpArg{{16, UInt, 0}, {11, Real_1_9, 0}},
		0},
	0x04: {"RDAX",
		[]OpArg{{6, UInt, 0}, {5, Blank, 0}, {16, Real_1_14, 0}},
		0},
	0x06: {"WRAX",
		[]OpArg{{6, UInt, 0}, {5, Blank, 0}, {16, Real_1_14, 0}},
		0},
	0x09: {"MAXX", // Also ABSA if all args=0
		[]OpArg{{6, UInt, 0}, {5, Blank, 0}, {16, Real_1_14, 0}},
		0},
	0x0A: {"MULX",
		[]OpArg{{6, UInt, 0}, {21, Blank, 0}},
		0},
	0x05: {"RDFX", // Also LDAX if all args 1&2 is 0
		[]OpArg{{6, UInt, 0}, {5, Blank, 0}, {16, Real_1_14, 0}},
		0},
	0x08: {"WRLX",
		[]OpArg{{6, UInt, 0}, {5, Blank, 0}, {16, Real_1_14, 0}},
		0},
	0x07: {"WRHX",
		[]OpArg{{6, UInt, 0}, {5, Blank, 0}, {16, Real_1_14, 0}},
		0},
	0x12: {"WLDR", // WLDS and WLDR (last flag: 00->WLDS, 01->WLDR)
		[]OpArg{{2, Flag, 0}, {6, Blank, 0}, {16, Int, 0}, {1, Flag, 0}, {2, Const, 1}},
		// WLDS
		//  []OpArg{{15, UInt, 0}, {9, UInt, 0}, {1, Flag, 0}, {2, Const, 0}},
		0},
	0x13: {"JAM",
		[]OpArg{{1, Flag, 0}, {25, Const, 1}},
		0},
	0x14: {"CHO", //  SubCmd: 0b00=RDA, 0b10=SOF, 0b11=RDAL,
		//       ADDR           N             0              FLAGS         SubCmd
		[]OpArg{{16, UInt, 0}, {2, Flag, 0}, {1, Const, 0}, {6, Flag, 0}, {2, UInt, 0}},
		0},
}

// Symbols as strings
var Symbols = map[int]string{
	0x00:       "SIN0_RATE",  // (0)  SIN 0 rate
	0x01:       "SIN0_RANGE", // (1)  SIN 0 range
	0x02:       "SIN1_RATE",  // (2)  SIN 1 rate
	0x03:       "SIN1_RANGE", // (3)  SIN 1 range
	0x04:       "RMP0_RATE",  // (4)  RMP 0 rate
	0x05:       "RMP0_RANGE", // (5)  RMP 0 range
	0x06:       "RMP1_RATE",  // (6)  RMP 1 rate
	0x07:       "RMP1_RANGE", // (7)  RMP 1 range
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

var SkpFlagSymbols = map[int]string{
	SKP_NEG: "NEG",
	SKP_GEZ: "GEZ",
	SKP_ZRO: "ZRO",
	SKP_ZRC: "ZRC",
	SKP_RUN: "RUN",
}

var ChoFlagSymbols = map[int]string{
	CHO_SIN:   "SIN",
	CHO_COS:   "COS",
	CHO_REG:   "REG",
	CHO_COMPC: "COMPC",
	CHO_COMPA: "COMPA",
	CHO_RPTR2: "RPTR2",
	CHO_NA:    "NA",
}

var LFOTypeNames = []string{
	"SIN0", "SIN1", "RMP0", "RMP1", "COS0", "COS1",
}

var RampAmpValuesMap = map[int16]int32{
	4096: 0,
	2048: 1,
	1024: 2,
	512:  3,
}

// FIXME: Correct order? Same as in disfv1.py, though. (20240713 handegar)
var RampAmpValues = []int16{4096, 2048, 1024, 512}

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
