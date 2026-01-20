package disasm

type OpDoc struct {
	Short    string
	Long     string
	Formulae string
}

var OpDocs = map[string]OpDoc{
	"SOF": {Short: "Scale and offset",
		Long: "SOF will multiply the current value in ACC with C and will then " +
			"add the constant D to the result.",
		Formulae: "C * ACC + D",
	},
	"AND": {Short: "Bit operation",
		Long: "AND will perform a bit wise \"and\" of the current ACC and the 24-bit " +
			"MASK specified within the instruction word. ",
		Formulae: "ACC & MASK",
	},
	"OR": {Short: "Bit operation",
		Long: "OR will perform a bit wise \"or\" of the current ACC and the 24-bit " +
			"MASK specified within the instruction word",
		Formulae: "ACC | MASK",
	},
	"XOR": {Short: "Bit operation",
		Long: "XOR will perform a bit wise \"xor\" of the current ACC and the 24-bit " +
			"MASK specified within the instruction word.",
		Formulae: "ACC ^ MASK",
	},
	"LOG": {Short: "Mathematical operation",
		Long: "LOG will multiply the Base2 LOG of the current absolute value in " +
			"ACC with C and add the constant D to the result.",
		Formulae: "C * LOG(|ACC|) + D",
	},
	"EXP": {Short: "Mathematical operation",
		Long:     "EXP will multiply 2^ACC with C and add the constant D to the result",
		Formulae: "C * EXP(ACC) + D",
	},
	"SKP": {Short: "Conditional skip",
		Long:     "The SKP instruction allows conditional program execution",
		Formulae: "CMASK N",
	},
	"RDAX": {Short: "Read from register",
		Long: "RDAX will fetch the value contained in [ADDR] from the register file, " +
			"multiply it with C and add the result to the previous content of ACC",
		Formulae: "C * REG[ADDR] + ACC",
	},
	"WRAX": {Short: "Write to register",
		Long:     "WRAX will save the current value in ACC to [ADDR] and then multiply ACC by C",
		Formulae: "ACC->REG[ADDR], C * ACC",
	},
	"MAXX": {Short: "Get max value of Reg*C or ACC",
		Long: "MAXX will compare the absolute value of ACC versus C times the absolute value " +
			"of the register pointed to by ADDR. If the absolute value of ACC is larger ACC " +
			"will be loaded with |ACC|, otherwise the accumulator becomes overwritten by " +
			"|REG[ADDR] * C|.",
		Formulae: "MAX(|REG[ADDR] * C|, |ACC|)",
	},
	"MULX": {Short: "Multiply ACC with register",
		Long:     "MULX will multiply ACC by the value of the register pointed to by ADDR.",
		Formulae: "ACC * REG[ADDR]",
	},
	"RDFX": {Short: "Multi-op instruction",
		Long: "RDFX will subtract the value of the register pointed to by ADDR from ACC, " +
			"multiply the result by C and then add the value of the register pointed to by ADDR.",
		Formulae: "(ACC-REG[ADDR])*C + REG[ADDR]",
	},
	"WRLX": {Short: "Multi-op instruction",
		Long: "First the current ACC value is stored into the register pointed to by ADDR, then " +
			"ACC is subtracted from the previous content of ACC (PACC). The difference is then " +
			"multiplied by C and finally PACC is added to the result.",
		Formulae: "ACC->REG[ADDR], (PACC-ACC)*C + PACC",
	},
	"WRHX": {Short: "Multi-op instruction",
		Long: "The current ACC value is stored in the register pointed to by ADDR, " +
			"then ACC is multiplied by C. Finally the previous content of ACC (PACC) is added to " +
			"the product",
		Formulae: "ACC->REG[ADDR], (ACC*C) + PACC",
	},
	"RDA": {Short: "Read from RAM",
		Long: "RDA will fetch the sample [ADDR] from the delay ram, multiply it by C and add " +
			"the result to the previous content of ACC.",
		Formulae: "SRAM[ADDR] * C + ACC",
	},
	"RMPA": {Short: "Indirect RAM read",
		Long: "RMPA provides indirect delay line addressing in that the delay line address of " +
			"the sample to be multiplied by C is not explicitly given in the instruction itself " +
			"but contained within the pointer register ADDR_PTR (absolute address 24 within the " +
			"internal register file.) ",
		Formulae: "SRAM[PNTR[N]] * C + ACC",
	},
	"WRA": {Short: "Write to RAM",
		Long: "WRA will store ACC to the delay ram location addressed by ADDR and then " +
			"multiply ACC by C.",
		Formulae: "ACC->SRAM[ADDR], ACC * C",
	},
	"WRAP": {Short: "Write to RAM and update ACC",
		Long: "WRAP will store ACC to the delay ram location addressed by ADDR then multiply ACC " +
			"by C and finally add the content of the LR register to the product",
		Formulae: "ACC->SRAM[ADDR], (ACC*C) + LR",
	},
	"WLDS": {Short: "Config sine LFO",
		Long: "WLDS will load frequency and amplitude control values into the selected " +
			"SIN LFO (0 or 1).",
		Formulae: "see datasheet",
	},
	"WLDR": {Short: "Config ramp LFO",
		Long: "WLDR will load frequency and amplitude control values into the selected " +
			"RAMP LFO. (0 or 1)",
		Formulae: "see datasheet",
	},
	"JAM": {Short: "Reset ramp LFO",
		Long:     "JAM will reset the selected RAMP LFO to its starting point",
		Formulae: "0 -> RAMP LFO N",
	},
	"CHO RDA": {Short: "Chorus read from MEM",
		Long: "Like the RDA instruction, CHO RDA will read a sample from the delay ram, " +
			"multiply it by a coefficient and an LFO and add the product to the previous content of ACC.",
		Formulae: "See datasheet",
	},
	"CHO SOF": {Short: "Chorus scale and offset",
		Long: "Like the SOF instruction, CHO SOF will multiply ACC by a coefficient and add " +
			"the constant D to the result. ", /*However, in contrast to SOF the coefficient is not " +
		"explicitly embedded within the instruction. Instead, based on the selected LFO and " +
		"the 6 bit vector C, the coefficient is picked from a list of possible coefficients " +
		"available within the LFO block of the FV-1"*/
		Formulae: "See datasheet",
	},
	"CHO RDAL": {Short: "Write the LFO value to ACC",
		Long:     "CHO RDAL will read the current value of the selected LFO into ACC.",
		Formulae: "LFO*1 -> ACC",
	},
	"CLR": {Short: "Clear ACC",
		Long:     "Clear the ACC register",
		Formulae: "0->ACC",
	},
	"NOT": {Short: "Bit operation",
		Long:     "NOT will negate all bit positions within accumulator thus performing a 1’s complement.",
		Formulae: "/ACC -> ACC",
	},
	"ABSA": {Short: "Absolute value of ACC",
		Long:     "Loads the accumulator with the absolute value of the accumulator",
		Formulae: "|ACC| -> ACC",
	},
	"LDAX": {Short: "Load register into ACC",
		Long:     "Loads the accumulator with the contents of the addressed register.",
		Formulae: "REG[ADDR] -> ACC",
	},
	"NOP": {Short: "No-Operation",
		Long:     "Does nothing. Same as 'SKP 0, 0'",
		Formulae: "No operation",
	},
}
