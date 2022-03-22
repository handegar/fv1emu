package dsp

import (
	"fmt"
	"math"

	"github.com/handegar/fv1emu/settings"
)

/**
  This is a register object which can hold any Q-Format number as long
  as it fits into the internal 32-bit S8.23 format (sig, 7bit integer,
  24bit fraction).

  The value is always considered as a potentially signed.
*/

type Register struct {
	Value int32 // FV-1 only uses 24 bits
	// This is the "public" Q-format, internally it's always S8.23
	IntBits      int
	FractionBits int
}

func NewRegister(value int32) *Register {
	r := new(Register)
	r.Value = value
	r.IntBits = 8
	r.FractionBits = 23
	return r
}

// Create a new register with a specified Q-Format
func NewRegisterWithIntsAndFracs(value int32, intbits int, fractionbits int) *Register {
	r := NewRegister(0)
	r.SetWithIntsAndFracs(value, intbits, fractionbits)
	return r
}

func NewRegisterWithFloat64(value float64) *Register {
	r := NewRegister(0)
	r.SetFloat64(value)
	return r
}

func (r *Register) Clear() *Register {
	r.Value = 0.0
	return r
}

func (r *Register) SetToMax24Bit() *Register {
	r.Value = 0x7fffff
	return r
}

func (r *Register) SetToMin24Bit() *Register {
	r.Value = -0x800000
	return r
}

// Returns TRUE on second value if value were clamped
func (r *Register) Clamp24Bit() (*Register, bool) {
	if settings.Disable24BitsClamping { // Shall we not perform clamping?
		return r, false
	}

	overfloweth := false
	if r.Value > 0x7fffff {
		r.Value = 0x7fffff
		overfloweth = true
	} else if r.Value < -0x800000 {
		r.Value = -0x800000
		overfloweth = true
	}
	return r, overfloweth
}

func (r *Register) extendSign() {
	// Shift all the way to the left to make the sign-bit "stick"
	shl := (31 - (r.IntBits + r.FractionBits))
	r.Value = r.Value << shl
	// Shift back right to the 8th position. The sign bit will then
	// be set for each position (this won't work for a uint32 value)
	r.Value = r.Value >> (8 - r.IntBits)
}

// Set a different Q-Format value than S8.23
func (r *Register) SetWithIntsAndFracs(value int32, intbits int, fractionbits int) *Register {
	// Will the actual number fit?
	if value > ((1 << (intbits + fractionbits + 1)) - 1) {
		fmt.Errorf("The value 0x%x won't fit int a S%d.%d", value, intbits, fractionbits)
	}

	r.IntBits = intbits
	r.FractionBits = fractionbits
	r.Value = value

	r.extendSign()
	return r
}

func (r *Register) SetFloat64(value float64) *Register {
	r.Value = int32(value * math.Pow(2, 23))
	r.IntBits = 8
	r.FractionBits = 23
	return r
}

func (r *Register) SetInt32(value int32) *Register {
	r.Value = value
	r.IntBits = 32
	r.FractionBits = 0
	return r
}

func (r *Register) ToInt32() int32 {
	return r.Value
}

func (r *Register) ToQFormat(intBits int, fractionBits int) int32 {
	shl := 23 - fractionBits
	x := r.Value >> shl
	y := x & ((1 << (intBits + fractionBits + 1)) - 1)
	return y
}

func (r *Register) ToFloat64() float64 {
	return float64(r.Value) / (1 << 23)
}

func (r *Register) Copy(reg *Register) *Register {
	r.Value = reg.Value
	r.IntBits = reg.IntBits
	r.FractionBits = reg.FractionBits
	return r
}

func (r *Register) Add(reg *Register) *Register {
	r.Value += reg.Value
	r.Clamp24Bit()
	return r
}

func (r *Register) Sub(reg *Register) *Register {
	r.Value -= reg.Value
	r.Clamp24Bit()
	return r
}

func (r *Register) Mult(reg *Register) *Register {
	v1 := int64(r.Value)
	v2 := int64(reg.Value)
	r.Value = int32((v1 * v2) >> 23)
	r.Clamp24Bit()
	return r
}

func (r *Register) Abs() *Register {
	// FIXME: Do this using bit shifts instead? (20220209 handegar)
	if r.IsSigned() {
		r.Value = -r.Value
	}
	return r
}

func (r *Register) And(value int32) *Register {
	r.Value = r.Value & value
	return r
}

func (r *Register) Or(value int32) *Register {
	r.Value = r.Value | value
	return r
}

func (r *Register) Xor(value int32) *Register {
	r.Value = r.Value ^ value
	return r
}

func (r *Register) Not(value int32) *Register {
	// Could we just do "value = -value"?
	r.Value = r.Value &^ value
	return r
}

func (r *Register) IsSigned() bool {
	return r.Value < 0
}

/*
   bitsFraction specifies the required precision for the compare.
   E.g: bitsFraction=3 -> epsilon=1/(2^3) (or 2^3 as an integer-value)
*/
func (r *Register) EqualWithEpsilon(reg *Register, lowestBitsFraction int) bool {
	a := r.Value >> (23 - (lowestBitsFraction - 1))
	b := reg.Value >> (23 - (lowestBitsFraction - 1))

	if a > b {
		return a-b <= 1
	} else {
		return b-a <= 1
	}
}

/*
   Performs a compare with a precision matched to the least precise
   register.
*/
func (r *Register) Equal(reg *Register) bool {
	a := r.Value
	b := reg.Value

	// Reduce precision to the least-precise value
	leastFractionBits := int(math.Min(float64(r.FractionBits), float64(reg.FractionBits)))

	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	diff = diff - 1 // To eliminate rounding errors
	return diff < (1 << (23 - leastFractionBits))
}

func (r *Register) GreaterThan(reg *Register) bool {
	return r.Value > reg.Value
}

func (r *Register) LessThan(reg *Register) bool {
	return r.Value < reg.Value
}

func (r *Register) DebugPrint() {
	fmt.Printf("\n")
	fmt.Printf("        |MSB     24      16      8    LSB| S%d.%d\n", r.IntBits, r.FractionBits)
	fmt.Printf("        |........v.......v.......v.......|\n")
	fmt.Printf("      [0b%32b], 0x%x, signed=%t\n", uint32(r.Value), uint32(r.Value), r.IsSigned())
	fmt.Printf("         ^^^^^^^^,^^^^^^^^^^^^^^^^^^^^^^^\n")
}
