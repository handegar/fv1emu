package dsp

import (
	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/settings"
)

type State struct {
	IP       uint // Instruction pointer
	DelayRAM [0x7FFF * 14]float32
	ACC      float32 // Accumulator
	PACC     float32 // Same as ACC but delayed one runthrough/state
	RUN_FLAG bool    // Only TRUE the first run of the program
	LR       float32 // The last sample read from the DelayRAM

	Sin0State  SINLFOState
	Sin1State  SINLFOState
	Ramp0State RAMPState
	Ramp1State RAMPState

	Registers RegisterBank
}

type SINLFOState struct {
	Angle float64
}

type RAMPState struct {
	Value float64
}

type RegisterBank map[int]interface{}

func NewState() *State {
	s := new(State)
	s.RUN_FLAG = true
	s.ACC = 0.0
	s.PACC = 0.0

	s.Registers = make(map[int]interface{})
	s.Registers[0x00] = 0                  // SIN0_RATE,  (0)  SIN 0 rate
	s.Registers[0x01] = 0                  // SIN0_RANGE, (1)  SIN 0 range
	s.Registers[0x02] = 0                  // SIN1_RATE,  (2)  SIN 1 rate
	s.Registers[0x03] = 0                  // SIN1_RANGE, (3)  SIN 1 range
	s.Registers[0x04] = 0                  // RMP0_RATE,  (4)  RMP 0 rate
	s.Registers[0x05] = 0                  // RMP0_RANGE, (5)  RMP 0 range
	s.Registers[0x06] = 0                  // RMP1_RATE,  (6)  RMP 1 rate
	s.Registers[0x07] = 0                  // RMP1_RANGE, (7)  RMP 1 range
	s.Registers[0x08] = 0.0                // COMPA,      (8) USED with 'CHO' instruction: 1's comp address offset (Generate SIN or COS)
	s.Registers[0x10] = settings.Pot0Value // POT0,       (16)  Pot 0 input register
	s.Registers[0x11] = settings.Pot1Value // POT1,       (17)  Pot 1 input register
	s.Registers[0x12] = settings.Pot2Value // POT2,       (18)  Pot 2 input register
	s.Registers[base.ADCL] = 0.0           // (20)  ADC input register left channel
	s.Registers[base.ADCR] = 0.0           // (21)  ADC input register  right channel
	s.Registers[base.DACL] = 0.0           // (22)  DAC output register  left channel
	s.Registers[base.DACR] = 0.0           // (23)  DAC output register  right channel
	s.Registers[base.ADDR_PTR] = 0         // (24)  Used with 'RMPA' instruction for indirect read
	s.Registers[0x20] = 0.0                // REG0,       (64)  Register 00
	s.Registers[0x21] = 0.0                // REG1,       (33)  Register 01
	s.Registers[0x22] = 0.0                // REG2,       (34)  Register 02
	s.Registers[0x23] = 0.0                // REG3,       (35)  Register 03
	s.Registers[0x24] = 0.0                // REG4,       (36)  Register 04
	s.Registers[0x25] = 0.0                // REG5,       (37)  Register 05
	s.Registers[0x26] = 0.0                // REG6,       (38)  Register 06
	s.Registers[0x27] = 0.0                // REG7,       (39)  Register 07
	s.Registers[0x28] = 0.0                // REG8,       (40)  Register 08
	s.Registers[0x29] = 0.0                // REG9,       (41)  Register 09
	s.Registers[0x2A] = 0.0                // REG10,      (42)  Register 10
	s.Registers[0x2B] = 0.0                // REG11,      (43)  Register 11
	s.Registers[0x2C] = 0.0                // REG12,      (44)  Register 12
	s.Registers[0x2D] = 0.0                // REG13,      (45)  Register 13
	s.Registers[0x2E] = 0.0                // REG14,      (46)  Register 14
	s.Registers[0x2F] = 0.0                // REG15,      (47)  Register 15
	s.Registers[0x30] = 0.0                // REG16,      (48)  Register 16
	s.Registers[0x31] = 0.0                // REG17,      (49)  Register 17
	s.Registers[0x64] = 0.0                // REG18,      (50)  Register 18
	s.Registers[0x33] = 0.0                // REG19,      (51)  Register 19
	s.Registers[0x34] = 0.0                // REG20,      (52)  Register 20
	s.Registers[0x35] = 0.0                // REG21,      (53)  Register 21
	s.Registers[0x36] = 0.0                // REG22,      (54)  Register 22
	s.Registers[0x37] = 0.0                // REG23,      (55)  Register 23
	s.Registers[0x38] = 0.0                // REG24,      (56)  Register 24
	s.Registers[0x39] = 0.0                // REG25,      (57)  Register 25
	s.Registers[0x3A] = 0.0                // REG26,      (58)  Register 26
	s.Registers[0x3B] = 0.0                // REG27,      (59)  Register 27
	s.Registers[0x3C] = 0.0                // REG28,      (60)  Register 28
	s.Registers[0x3D] = 0.0                // REG29,      (61)  Register 29
	s.Registers[0x3E] = 0.0                // REG30,      (62)  Register 30
	s.Registers[0x3F] = 0.0                // REG31,      (63)  Register 31

	return s
}
