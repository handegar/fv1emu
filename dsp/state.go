package dsp

import (
	"fmt"

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

	DebugFlags *DebugFlags // Contains misc debug/error flags which will be set @ runtime

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

type DebugFlags struct {
	InvalidRamp0Values bool
	InvalidRamp1Values bool
	InvalidSin0Values  bool
	InvalidSin1Values  bool

	OutOfBoundsMemoryRead  bool // Not set yet
	OutOfBoundsMemoryWrite bool // Not set yet

	InvalidRegister int // Not set yet

	ACCOverflowCount  int
	PACCOverflowCount int
	LROverflowCount   int
	DACROverflowCount int
	DACLOverflowCount int
}

func (df *DebugFlags) SetSinLFOFlag(lfoNum int32) {
	if lfoNum == 0 {
		df.InvalidSin0Values = true
	} else {
		df.InvalidSin1Values = true
	}
}

func (df *DebugFlags) SetRampLFOFlag(lfoNum int32) {
	if lfoNum == 0 {
		df.InvalidRamp0Values = true
	} else {
		df.InvalidRamp1Values = true
	}
}

func (df *DebugFlags) Reset() {
	df.InvalidRamp0Values = false
	df.InvalidRamp1Values = false
	df.InvalidSin0Values = false
	df.InvalidSin1Values = false

	df.OutOfBoundsMemoryRead = false
	df.OutOfBoundsMemoryWrite = false

	df.InvalidRegister = 0

	df.ACCOverflowCount = 0
	df.PACCOverflowCount = 0
	df.LROverflowCount = 0

	df.DACROverflowCount = 0
	df.DACLOverflowCount = 0
}

func (df *DebugFlags) Print() {
	fmt.Printf("DebugFlags:\n"+
		" InvalidRamp0Values = %t\n"+
		" InvalidRamp1Values = %t\n"+
		" InvalidSin0Values = %t\n"+
		" InvalidSin1Values = %t\n"+
		" OutOfBoundsMemoryRead = %d\n"+
		" OutOfBoundsMemoryWrite = %d\n"+
		" InvalidRegister = %d\n"+
		" ACCOverflowCount = %d\n"+
		" PACCOverflowCount = %d\n"+
		" LROverflowCount = %d\n"+
		" DACROverflowCount = %d\n"+
		" DACLOverflowCount = %d\n",
		df.InvalidRamp0Values,
		df.InvalidRamp1Values,
		df.InvalidSin0Values,
		df.InvalidSin1Values,
		df.OutOfBoundsMemoryRead,
		df.OutOfBoundsMemoryWrite,
		df.InvalidRegister,
		df.ACCOverflowCount,
		df.PACCOverflowCount,
		df.LROverflowCount,
		df.DACROverflowCount,
		df.DACLOverflowCount)
}

type SINLFOState struct {
	Angle float64
}

type RAMPState struct {
	Value float64
}

type RegisterBank map[int]*Register

func NewState() *State {
	s := new(State)
	s.DebugFlags = new(DebugFlags)
	s.Reset()
	return s
}

func (s *State) CheckForOverflows() {
	acc := s.ACC.ToFloat64()
	if acc > 1.0 || acc < -1.0 {
		s.DebugFlags.ACCOverflowCount += 1
	}
	pacc := s.PACC.ToFloat64()
	if pacc > 1.0 || pacc < -1.0 {
		s.DebugFlags.PACCOverflowCount += 1
	}
	lr := s.LR.ToFloat64()
	if lr > 1.0 || lr < -1.0 {
		s.DebugFlags.LROverflowCount += 1
	}
	dacl := s.GetRegister(base.DACL).ToFloat64()
	if dacl > 1.0 || dacl < -1.0 {
		s.DebugFlags.DACLOverflowCount += 1
	}
	dacr := s.GetRegister(base.DACR).ToFloat64()
	if dacr > 1.0 || dacr < -1.0 {
		s.DebugFlags.DACROverflowCount += 1
	}
}

func (s *State) GetRegister(regNo int) *Register {
	err := validateRegisterNo(regNo)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		s.DebugFlags.InvalidRegister = regNo
		s.DebugFlags.Print()
		panic("Invalid register number")
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

	s.DebugFlags.Reset()
}

func validateRegisterNo(regNo int) error {
	if !(regNo >= 0 && regNo <= base.RAMP1_RANGE) &&
		!(regNo >= base.POT0 && regNo <= base.POT2) &&
		!(regNo >= base.ADCL && regNo <= base.ADDR_PTR) &&
		!(regNo >= 0x20 && regNo <= 0x3F) {
		return fmt.Errorf("RegNo = %d (0x%x) is not in use\n", regNo, regNo)
	}
	return nil
}
