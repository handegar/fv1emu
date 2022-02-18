package dsp

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/settings"
)

const DELAY_RAM_SIZE = 0x7FFF * 14

type State struct {
	IP          uint                  // Instruction pointer
	DelayRAM    [DELAY_RAM_SIZE]int32 // Internal memory
	DelayRAMPtr int                   // Moving delay-ram pointer. Decreased each run-through
	ACC         *Register             // Accumulator
	PACC        *Register             // Same as ACC but from the previous run-through
	LR          *Register             // The last sample read from the DelayRAM
	RUN_FLAG    bool                  // Only TRUE the first run of the program

	Sin0State  SINLFOState
	Sin1State  SINLFOState
	Ramp0State RAMPState
	Ramp1State RAMPState
	sinLFOReg  *Register // Holds the frozen sine LFO value
	rampLFOReg *Register // Holds the frozen ramp LFO value

	Registers RegisterBank

	workRegA *Register // Register used interally for certain operations
	workRegB *Register // Register used interally for certain operations

	workReg0_23 *Register // S.23
	workReg0_10 *Register // S.10
	workReg1_14 *Register // S1.14
	workReg1_9  *Register // S1.9
	workReg4_6  *Register // S4.6
}

type SINLFOState struct {
	Angle float64
}

type RAMPState struct {
	Value float64
}

type RegisterBank map[int]*Register

func (s *State) GetRegister(regNo int) *Register {
	if validateRegisterNo(regNo) != nil {
		panic(true)
	}

	return s.Registers[regNo]
}

func (s *State) Reset() {
	s.IP = 0
	s.RUN_FLAG = false
	s.ACC = NewRegister(0)
	s.PACC = NewRegister(0)
	s.LR = NewRegister(0)

	s.sinLFOReg = NewRegister(0)
	s.rampLFOReg = NewRegister(0)

	s.workRegA = NewRegister(0)
	s.workRegB = NewRegister(0)

	s.workReg0_23 = NewRegisterWithIntsAndFracs(0, 0, 23)
	s.workReg0_10 = NewRegisterWithIntsAndFracs(0, 0, 10)
	s.workReg1_14 = NewRegisterWithIntsAndFracs(0, 1, 14)
	s.workReg1_9 = NewRegisterWithIntsAndFracs(0, 1, 9)
	s.workReg4_6 = NewRegisterWithIntsAndFracs(0, 4, 6)

	s.DelayRAMPtr = 0

	s.Registers = make(map[int]*Register)
	for i := 0; i < 64; i++ {
		if (i >= 8 && i <= 15) || i == 19 || (i >= 25 && i <= 31) {
			continue // Unused registers
		}
		s.Registers[i] = NewRegister(0) // All registers are S.23 as default
	}

	s.GetRegister(base.SIN0_RATE).Clear()   // SIN0_RATE,  (0)  SIN 0 rate
	s.GetRegister(base.SIN0_RANGE).Clear()  // SIN0_RANGE, (1)  SIN 0 range
	s.GetRegister(base.SIN1_RATE).Clear()   // SIN1_RATE,  (2)  SIN 1 rate
	s.GetRegister(base.SIN1_RANGE).Clear()  // SIN1_RANGE, (3)  SIN 1 range
	s.GetRegister(base.RAMP0_RATE).Clear()  // RMP0_RATE,  (4)  RMP 0 rate
	s.GetRegister(base.RAMP0_RANGE).Clear() // RMP0_RANGE, (5)  RMP 0 range
	s.GetRegister(base.RAMP1_RATE).Clear()  // RMP1_RATE,  (6)  RMP 1 rate
	s.GetRegister(base.RAMP1_RANGE).Clear() // RMP1_RANGE, (7)  RMP 1 range

	s.GetRegister(base.POT0).SetFloat64(settings.Pot0Value) // POT0,       (16)  Pot 0 input register
	s.GetRegister(base.POT1).SetFloat64(settings.Pot1Value) // POT1,       (17)  Pot 1 input register
	s.GetRegister(base.POT2).SetFloat64(settings.Pot2Value) // POT2,       (18)  Pot 2 input register

	s.GetRegister(base.ADCL).Clear() // (20)  ADC input register left channel
	s.GetRegister(base.ADCR).Clear() // (21)  ADC input register  right channel
	s.GetRegister(base.DACL).Clear() // (22)  DAC output register  left channel
	s.GetRegister(base.DACR).Clear() // (23)  DAC output register  right channel

	s.GetRegister(base.ADDR_PTR).Clear() // (24)  Used with 'RMPA' instruction for indirect read

	s.GetRegister(0x20).Clear() // REG0,       (64)  Register 00
	s.GetRegister(0x21).Clear() // REG1,       (33)  Register 01
	s.GetRegister(0x22).Clear() // REG2,       (34)  Register 02
	s.GetRegister(0x23).Clear() // REG3,       (35)  Register 03
	s.GetRegister(0x24).Clear() // REG4,       (36)  Register 04
	s.GetRegister(0x25).Clear() // REG5,       (37)  Register 05
	s.GetRegister(0x26).Clear() // REG6,       (38)  Register 06
	s.GetRegister(0x27).Clear() // REG7,       (39)  Register 07
	s.GetRegister(0x28).Clear() // REG8,       (40)  Register 08
	s.GetRegister(0x29).Clear() // REG9,       (41)  Register 09
	s.GetRegister(0x2A).Clear() // REG10,      (42)  Register 10
	s.GetRegister(0x2B).Clear() // REG11,      (43)  Register 11
	s.GetRegister(0x2C).Clear() // REG12,      (44)  Register 12
	s.GetRegister(0x2D).Clear() // REG13,      (45)  Register 13
	s.GetRegister(0x2E).Clear() // REG14,      (46)  Register 14
	s.GetRegister(0x2F).Clear() // REG15,      (47)  Register 15
	s.GetRegister(0x30).Clear() // REG16,      (48)  Register 16
	s.GetRegister(0x31).Clear() // REG17,      (49)  Register 17
	s.GetRegister(0x32).Clear() // REG18,      (50)  Register 18
	s.GetRegister(0x33).Clear() // REG19,      (51)  Register 19
	s.GetRegister(0x34).Clear() // REG20,      (52)  Register 20
	s.GetRegister(0x35).Clear() // REG21,      (53)  Register 21
	s.GetRegister(0x36).Clear() // REG22,      (54)  Register 22
	s.GetRegister(0x37).Clear() // REG23,      (55)  Register 23
	s.GetRegister(0x38).Clear() // REG24,      (56)  Register 24
	s.GetRegister(0x39).Clear() // REG25,      (57)  Register 25
	s.GetRegister(0x3A).Clear() // REG26,      (58)  Register 26
	s.GetRegister(0x3B).Clear() // REG27,      (59)  Register 27
	s.GetRegister(0x3C).Clear() // REG28,      (60)  Register 28
	s.GetRegister(0x3D).Clear() // REG29,      (61)  Register 29
	s.GetRegister(0x3E).Clear() // REG30,      (62)  Register 30
	s.GetRegister(0x3F).Clear() // REG31,      (63)  Register 31
}

func NewState() *State {
	s := new(State)
	s.Reset()
	return s
}

func (s *State) Print() {
	color.Blue("IP=%d, ACC=%d, ADDR_PTR=%d, DelayRAMPtr=%d, RUN_FLAG=%t",
		s.IP, s.ACC.Value,
		s.Registers[base.ADDR_PTR].Value, s.DelayRAMPtr,
		s.RUN_FLAG)
	color.Cyan("ADCL=%d, ADCR=%d, DACL=%d, DACR=%d\n",
		s.Registers[base.ADCL].Value, s.Registers[base.ADCR].Value,
		s.Registers[base.DACL].Value, s.Registers[base.DACR].Value)
	color.Cyan("POT0=%f, POT1=%f, POT2=%f\n",
		s.Registers[base.POT0].ToFloat64(),
		s.Registers[base.POT1].ToFloat64(),
		s.Registers[base.POT2].ToFloat64())

	fmt.Printf(`REG0-7:   [%f, %f, %f, %f, %f, %f, %f, %f]
REG8-15:  [%f, %f, %f, %f, %f, %f, %f, %f]
REG16-23: [%f, %f, %f, %f, %f, %f, %f, %f]
REG24-31: [%f, %f, %f, %f, %f, %f, %f, %f]
`,
		s.Registers[0x20].ToFloat64(), s.Registers[0x21].ToFloat64(),
		s.Registers[0x22].ToFloat64(), s.Registers[0x23].ToFloat64(),
		s.Registers[0x24].ToFloat64(), s.Registers[0x25].ToFloat64(),
		s.Registers[0x26].ToFloat64(), s.Registers[0x27].ToFloat64(),
		s.Registers[0x28].ToFloat64(), s.Registers[0x29].ToFloat64(),
		s.Registers[0x2a].ToFloat64(), s.Registers[0x2b].ToFloat64(),
		s.Registers[0x2c].ToFloat64(), s.Registers[0x2d].ToFloat64(),
		s.Registers[0x2e].ToFloat64(), s.Registers[0x2f].ToFloat64(),
		s.Registers[0x30].ToFloat64(), s.Registers[0x31].ToFloat64(),
		s.Registers[0x32].ToFloat64(), s.Registers[0x33].ToFloat64(),
		s.Registers[0x34].ToFloat64(), s.Registers[0x35].ToFloat64(),
		s.Registers[0x36].ToFloat64(), s.Registers[0x37].ToFloat64(),
		s.Registers[0x38].ToFloat64(), s.Registers[0x39].ToFloat64(),
		s.Registers[0x3a].ToFloat64(), s.Registers[0x3b].ToFloat64(),
		s.Registers[0x3c].ToFloat64(), s.Registers[0x3d].ToFloat64(),
		s.Registers[0x3e].ToFloat64(), s.Registers[0x3f].ToFloat64(),
	)
	color.Cyan("SIN0 Rate/range=[%d(%f), %d(%f)]\nSIN1 Rate/range=[%d(%f), %d(%f)]\n",
		s.Registers[base.SIN0_RATE].Value, s.Registers[base.SIN0_RATE].ToFloat64(),
		s.Registers[base.SIN0_RANGE].Value, s.Registers[base.SIN0_RANGE].ToFloat64(),
		s.Registers[base.SIN1_RATE].Value, s.Registers[base.SIN1_RATE].ToFloat64(),
		s.Registers[base.SIN1_RANGE].Value, s.Registers[base.SIN1_RANGE].ToFloat64())
	color.White("SIN0 Angle=%f, SIN1 Angle=%f\n", s.Sin0State.Angle, s.Sin1State.Angle)
	color.Cyan("RAMP0 Rate/range=[%d, %d], RAMP1 Rate/range=[%d, %d]\n",
		s.Registers[base.RAMP0_RATE].Value, s.Registers[base.RAMP0_RANGE].Value,
		s.Registers[base.RAMP1_RATE].Value, s.Registers[base.RAMP1_RANGE].Value)
	color.White("RAMP0 Value=%f, RAMP1 Value=%f\n", s.Ramp0State.Value, s.Ramp1State.Value)
}

func validateRegisterNo(regNo int) error {
	if !(regNo >= 0 && regNo <= base.RAMP1_RANGE) &&
		!(regNo >= base.POT0 && regNo <= base.POT2) &&
		!(regNo >= base.ADCL && regNo <= base.ADDR_PTR) &&
		!(regNo >= 0x20 && regNo <= 0x3F) {
		return fmt.Errorf("RegNo = %d (0x%x) is not in use\n", regNo, regNo)
	}
	/*
		if regNo < 0 || regNo > 63 {
			return fmt.Errorf("RegNo = %d (0x%x) is out of bounds\n", regNo, regNo)
		}
	*/
	return nil
}
