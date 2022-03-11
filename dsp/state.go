package dsp

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/settings"
)

const DELAY_RAM_SIZE = 0x7FFF * 14

type State struct {
	IP          uint                  // Instruction pointer
	DelayRAM    [DELAY_RAM_SIZE]int32 // Internal memory
	DelayRAMPtr int                   // Moving delay-ram pointer. Decreased each run-through
	ACC         *Register             // Accumulator (S1.14 or S.23?)
	PACC        *Register             // Same as ACC but from the previous run-through
	LR          *Register             // The last sample read from the DelayRAM
	RUN_FLAG    bool                  // Only TRUE the first run of the program

	Sin0State  SINLFOState
	Sin1State  SINLFOState
	Ramp0State RAMPState
	Ramp1State RAMPState

	DebugFlags *DebugFlags // Contains misc debug/error flags which will be set @ runtime

	// Holdes
	sin0LFOReg  *Register // Holds the frozen sine LFO value
	sin1LFOReg  *Register // Holds the frozen sine LFO value
	ramp0LFOReg *Register // Holds the frozen ramp LFO value
	ramp1LFOReg *Register // Holds the frozen ramp LFO value

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

	OutOfBoundsMemoryRead  int
	OutOfBoundsMemoryWrite int

	InvalidRegister int // Not set yet

	ACCOverflowCount  int
	PACCOverflowCount int
	LROverflowCount   int
	DACROverflowCount int
	DACLOverflowCount int

	// Used when debugging
	Ramp0Min float64
	Ramp0Max float64
	Ramp1Min float64
	Ramp1Max float64
	Sin0Min  float64
	Sin0Max  float64
	Sin1Min  float64
	Sin1Max  float64
	XFadeMin float64
	XFadeMax float64
}

func (df *DebugFlags) SetInvalidSinLFOFlag(lfoNum int32) {
	if lfoNum == 0 {
		df.InvalidSin0Values = true
	} else {
		df.InvalidSin1Values = true
	}
}

func (df *DebugFlags) SetInvalidRampLFOFlag(lfoNum int32) {
	if lfoNum == 0 {
		df.InvalidRamp0Values = true
	} else {
		df.InvalidRamp1Values = true
	}
}

func (df *DebugFlags) IncreaseOutOfBoundsMemoryRead() error {
	df.OutOfBoundsMemoryRead += 1

	msg := "Out-of-bounds memory read: "
	_, file, no, ok := runtime.Caller(1)
	if ok {
		msg += fmt.Sprintf("%s:%d", file, no)
	}
	return errors.New(msg)
}

func (df *DebugFlags) IncreaseOutOfBoundsMemoryWrite() error {
	df.OutOfBoundsMemoryWrite += 1
	msg := "Out-of-bounds memory write: "
	_, file, no, ok := runtime.Caller(1)
	if ok {
		msg += fmt.Sprintf("%s:%d", file, no)
	}
	return errors.New(msg)
}

func (df *DebugFlags) Reset() {
	df.InvalidRamp0Values = false
	df.InvalidRamp1Values = false
	df.InvalidSin0Values = false
	df.InvalidSin1Values = false

	df.OutOfBoundsMemoryRead = 0
	df.OutOfBoundsMemoryWrite = 0

	df.InvalidRegister = 0

	df.ACCOverflowCount = 0
	df.PACCOverflowCount = 0
	df.LROverflowCount = 0

	df.DACROverflowCount = 0
	df.DACLOverflowCount = 0

	// Internal stuff
	df.Ramp0Min = 999.0
	df.Ramp0Max = -999.0
	df.Ramp1Min = 999.0
	df.Ramp1Max = -999.0
	df.XFadeMax = -999.0
	df.XFadeMin = 999.0
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

func (s *State) Copy(in *State) {
	s.IP = in.IP
	s.RUN_FLAG = in.RUN_FLAG
	s.ACC.Copy(in.ACC)
	s.PACC.Copy(in.PACC)
	s.LR.Copy(in.LR)
	s.Sin0State.Angle = in.Sin0State.Angle
	s.Sin1State.Angle = in.Sin1State.Angle
	s.Ramp0State.Value = in.Ramp0State.Value
	s.Ramp1State.Value = in.Ramp1State.Value
	s.DelayRAMPtr = in.DelayRAMPtr

	for i := 0; i < 64; i++ {
		if (i >= 8 && i <= 15) || i == 19 || (i >= 25 && i <= 31) {
			s.Registers[i] = nil
			continue // Unused registers
		}
		s.Registers[i] = NewRegister(0) // All registers are S.23 as default
		s.Registers[i].Copy(in.Registers[i])
	}
}

func (s *State) Duplicate() *State {
	new := NewState()
	new.Copy(s)
	return new
}

func (s *State) Reset() {
	s.IP = 0
	s.RUN_FLAG = false
	s.ACC = NewRegister(0)
	s.PACC = NewRegister(0)
	s.LR = NewRegister(0)

	s.sin0LFOReg = NewRegister(0)
	s.sin1LFOReg = NewRegister(0)
	s.ramp0LFOReg = NewRegister(0)
	s.ramp1LFOReg = NewRegister(0)
	s.Sin0State.Angle = 0.0
	s.Sin1State.Angle = 0.0
	s.Ramp0State.Value = 0.0
	s.Ramp1State.Value = 0.0

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

	s.GetRegister(base.POT0).SetFloat64(settings.Pot0Value) // POT0,       (16)  Pot 0 input register
	s.GetRegister(base.POT1).SetFloat64(settings.Pot1Value) // POT1,       (17)  Pot 1 input register
	s.GetRegister(base.POT2).SetFloat64(settings.Pot2Value) // POT2,       (18)  Pot 2 input register

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
