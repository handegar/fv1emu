package dsp

import (
	"fmt"

	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/settings"
)

const DELAY_RAM_SIZE = 0x8000 // 32k of delay RAM -> 1sec with a 32768Hz clock

const RAMP_START = 0.0
const RAMP_END = 0.5

type State struct {
	IP          uint                  // Instruction pointer
	DelayRAM    [DELAY_RAM_SIZE]int32 // Internal memory
	DelayRAMPtr int                   // Moving delay-ram pointer. Decreased each run-through
	ACC         *Register             // Accumulator (S1.14 or S.23?)
	PACC        *Register             // Same as ACC but from the previous run-through
	LR          *Register             // The last sample read from the DelayRAM
	RUN_FLAG    bool                  // Only TRUE the first run of the program

	Sin0Osc  SineOscillator
	Sin1Osc  SineOscillator
	Ramp0Osc RampOscillator
	Ramp1Osc RampOscillator

	DebugFlags *DebugFlags // Contains misc debug/error flags which will be set @ runtime

	// Holdes
	sin0LFOReg  *Register // Holds the frozen sine LFO value
	sin1LFOReg  *Register // Holds the frozen sine LFO value
	ramp0LFOReg *Register // Holds the frozen ramp LFO value
	ramp1LFOReg *Register // Holds the frozen ramp LFO value

	Registers RegisterBank

	workRegA  *Register // Register used interally for certain operations
	workRegB  *Register // Register used interally for certain operations
	scaleReg  *Register // Register used interally for certain operations
	offsetReg *Register // Register used interally for certain operations

	workReg0_23 *Register // S.23
	workReg0_10 *Register // S.10
	workReg1_14 *Register // S1.14
	workReg1_9  *Register // S1.9
	workReg4_6  *Register // S4.6
}

func (s *State) UpdateSineLFOs() {
	s.Sin0Osc.Update()
	s.Sin1Osc.Update()
}

func (s *State) UpdateRampLFOs() {
	s.Ramp0Osc.Update()
	s.Ramp1Osc.Update()
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
	s.Sin0Osc.value = in.Sin0Osc.value
	s.Sin1Osc.value = in.Sin1Osc.value
	s.Ramp0Osc.value = in.Ramp0Osc.value
	s.Ramp1Osc.value = in.Ramp1Osc.value
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
	s.Sin0Osc.value = 0
	s.Sin1Osc.value = 0
	s.Ramp0Osc.Reset()
	s.Ramp1Osc.Reset()

	s.workRegA = NewRegister(0)
	s.workRegB = NewRegister(0)
	s.scaleReg = NewRegister(0)
	s.offsetReg = NewRegister(0)

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

	s.GetRegister(base.POT0).SetClampedFloat64(settings.Pot0Value) // POT0, (16) Pot 0 input register
	s.GetRegister(base.POT1).SetClampedFloat64(settings.Pot1Value) // POT1, (17) Pot 1 input register
	s.GetRegister(base.POT2).SetClampedFloat64(settings.Pot2Value) // POT2, (18) Pot 2 input register

	s.GetRegister(base.RAMP0_RANGE).Value = 512
	s.GetRegister(base.RAMP1_RANGE).Value = 512

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
