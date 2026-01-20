package dsp

import (
	"errors"
	"fmt"
	"runtime"
)

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
