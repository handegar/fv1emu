package dsp

import (
	"math"
	"testing"
)

func Test_Fundamentals(t *testing.T) {
	r1 := NewRegisterWithIntsAndFracs(0x200, 0, 10)    // 0.5
	r2 := NewRegisterWithIntsAndFracs(0x2000, 1, 14)   // 0.5
	r3 := NewRegisterWithIntsAndFracs(0x400000, 0, 23) // 0.5
	rf := NewRegisterWithFloat64(0.5)

	if r1.Value != r2.Value ||
		r1.Value != r3.Value ||
		r1.Value != rf.Value {
		t.Fatalf("Q-Format transponse-result not similar for all formats.")
	}

	if r1.IsSigned() {
		t.Fatalf("Register says is is signed, but it is not")
	}

	r1 = NewRegisterWithIntsAndFracs(0x600, 0, 10)    // -0.5
	r2 = NewRegisterWithIntsAndFracs(0xe000, 1, 14)   // -0.5
	r3 = NewRegisterWithIntsAndFracs(0xC00000, 0, 23) // -0.5
	rf = NewRegisterWithFloat64(-0.5)

	if r1.Value != r2.Value ||
		r1.Value != r3.Value ||
		r1.Value != rf.Value {
		t.Fatalf("Q-Format transponse-result not similar for all formats.")
	}

	if !r1.IsSigned() {
		t.Fatalf("Register says is is unsigned, but it is not")
	}

}

func Test_ToQFormat(t *testing.T) {
	r1 := NewRegisterWithIntsAndFracs(0x200, 0, 10) // 0.5
	r1q := r1.ToQFormat(0, 10)
	if r1q != 0x200 {
		t.Fatalf("S.10 -> reg -> S.10 did not yield the same result")
	}

	r2 := NewRegisterWithIntsAndFracs(0x2000, 1, 14) // 0.5
	r2q := r2.ToQFormat(1, 14)
	if r2q != 0x2000 {
		t.Fatalf("S1.14 -> reg -> S1.14 did not yield the same result")
	}

	r3 := NewRegisterWithIntsAndFracs(0x400000, 0, 23) // 0.5
	r3q := r3.ToQFormat(0, 23)
	if r3q != 0x400000 {
		t.Fatalf("S.23 -> reg -> S.23 did not yield the same result")
	}

	r1 = NewRegisterWithIntsAndFracs(0x600, 0, 10) // -0.5
	r1q = r1.ToQFormat(0, 10)
	if r1q != 0x600 {
		t.Fatalf("S.10 -> reg -> S.10 did not yield the same result")
	}

	r2 = NewRegisterWithIntsAndFracs(0xe000, 1, 14) // -0.5
	r2q = r2.ToQFormat(1, 14)
	if r2q != 0xe000 {
		t.Fatalf("S1.14 -> reg -> S1.14 did not yield the same result")
	}

	r3 = NewRegisterWithIntsAndFracs(0xc00000, 0, 23) // -0.5
	r3q = r3.ToQFormat(0, 23)
	if r3q != 0xc00000 {
		t.Fatalf("S.23 -> reg -> S.23 did not yield the same result")
	}
}

func Test_ToFromFloat64(t *testing.T) {
	r := NewRegister(0)
	expected := 0.1
	r.SetFloat64(expected)
	asFloat := r.ToFloat64()
	if math.Abs(asFloat-expected) > (1.0 / (1 << 23)) {
		t.Errorf("FAILED: Got %f , expected %f", asFloat, expected)
	}

	expected = 1.123456789
	r.SetFloat64(expected)
	asFloat = r.ToFloat64()
	if math.Abs(asFloat-expected) > (1.0 / (1 << 23)) {
		t.Errorf("FAILED: Got %f , expected %f", asFloat, expected)
	}

}

func Test_FromQFormat_S0_23(t *testing.T) {
	// Using
	// https://www.venea.net/web/q_format_conversion#q_format_form
	// to get the qformat integers.
	r1 := NewRegisterWithIntsAndFracs(0b100_0000_0000_0000_0000_0000, 0, 23) // 0.5
	r2 := NewRegisterWithFloat64(0.5)
	if r1.ToFloat64() != r2.ToFloat64() {
		t.Fatalf("%f != %f\n", r1.ToFloat64(), r2.ToFloat64())
	}

	r1 = NewRegisterWithIntsAndFracs(0x7FDF3B, 0, 23)
	r2 = NewRegisterWithFloat64(0.999)
	if r1.ToFloat64() != r2.ToFloat64() {
		t.Fatalf("%f != %f\n", r1.ToFloat64(), r2.ToFloat64())
	}

	r1 = NewRegisterWithIntsAndFracs(0x800000, 0, 23)
	r2 = NewRegisterWithFloat64(-1.0)
	if r1.ToFloat64() != r2.ToFloat64() {
		t.Fatalf("%f != %f\n", r1.ToFloat64(), r2.ToFloat64())
	}
}

func Test_FromQFormat_S1_14(t *testing.T) {
	// S1.14
	r1 := NewRegisterWithIntsAndFracs(0x8000, 1, 14) // -2
	if r1.ToFloat64() != -2.0 {
		t.Errorf("FAILED: Could not get 0x8000 as S1.14=-2.0. Got %f",
			r1.ToFloat64())
	}

	r2 := NewRegisterWithFloat64(-2.0)
	if !r1.Equal(r2) {
		t.Errorf("FAILED: Could not get -2.0 as S1.14: expected %f, got %f",
			r2.ToFloat64(), r1.ToFloat64())
	}

	r1 = NewRegisterWithIntsAndFracs(0x7FFF, 1, 14)
	if math.Abs(r1.ToFloat64()-1.99993896484) > 0.00006 {
		t.Errorf("FAILED: Could not get 0x7FFF as S1.14=1.99993896484. Got %f",
			r1.ToFloat64())
	}

	r2 = NewRegisterWithFloat64(1.99993896484)
	if !r1.Equal(r2) {
		t.Errorf("FAILED: Could not get 1.99993896484 as S1.14: %f != %f",
			r1.ToFloat64(), r2.ToFloat64())
	}
}

func Test_FromQFormat_S0_10(t *testing.T) {
	r1 := NewRegisterWithIntsAndFracs(0x200, 0, 10) // 0.5
	r2 := NewRegisterWithFloat64(0.5)
	if r1.ToFloat64() != r2.ToFloat64() {
		t.Fatalf("%f != %f\n", r1.ToFloat64(), r2.ToFloat64())
	}

	r1 = NewRegisterWithIntsAndFracs(0x3FF, 0, 10)
	r2 = NewRegisterWithFloat64(0.999)
	if !r1.EqualWithEpsilon(r2, 10) {
		t.Fatalf("%f != %f\n", r1.ToFloat64(), r2.ToFloat64())
	}

	r1 = NewRegisterWithIntsAndFracs(0x400, 0, 10)
	r2 = NewRegisterWithFloat64(-1.0)
	if r1.ToFloat64() != r2.ToFloat64() {
		t.Fatalf("%f != %f\n", r1.ToFloat64(), r2.ToFloat64())
	}
}

func Test_Add_SimilarQFormats(t *testing.T) {
	r1 := NewRegisterWithFloat64(-1.25)
	r2 := NewRegisterWithFloat64(0)
	r1.Add(r2)

	expected := NewRegister(0)
	expected.SetFloat64(-1.25)

	if !r1.Equal(expected) {
		t.Errorf("FAILED: Got %d, expected %d\n"+
			"got: [0b%32b]\nexp: [0b%32b]\n",
			r1.Value, expected.Value,
			uint32(r1.Value), uint32(expected.Value))
	}

	r1.SetFloat64(1.234)
	r2.SetFloat64(-1.322)
	r1.Add(r2)
	expected.SetFloat64(1.234 - 1.322)

	if !r1.Equal(expected) {
		t.Errorf("FAILED: Got %d, expected %d\n"+
			"got: [0b%32b]\nexp: [0b%32b]\n",
			r1.Value, expected.Value,
			uint32(r1.Value), uint32(expected.Value))
	}

	r1.SetFloat64(1.234)
	r2.SetFloat64(1.322)
	r1.Add(r2)
	expected.SetFloat64(1.234 + 1.322)

	if !r1.Equal(expected) {
		t.Errorf("FAILED: Got %f, expected %f\n"+
			"got: %d\nexp: %d\n",
			r1.ToFloat64(), expected.ToFloat64(),
			uint32(r1.Value), uint32(expected.Value))
	}
}

func Test_Add_DifferentQFormats(t *testing.T) {
	r1 := NewRegisterWithIntsAndFracs(0x800000, 0, 23) // -1
	r2 := NewRegisterWithIntsAndFracs(0xc000, 1, 14)   // -1
	r3 := NewRegisterWithFloat64(-1.0)                 // -1

	if !r1.Equal(r2) {
		t.Fatalf("Expected both to be -1, got %f and %f",
			r1.ToFloat64(), r2.ToFloat64())
	}

	// Different Q-Formats
	r1 = NewRegisterWithIntsAndFracs(0xCD, 0, 10)     // 0.2
	r2 = NewRegisterWithIntsAndFracs(0xCCD, 1, 14)    // 0.2
	r3 = NewRegisterWithIntsAndFracs(0x400000, 0, 23) // 0.5

	expected := NewRegisterWithFloat64(0.2 + 0.2)

	r1.Add(r2)
	if !r1.EqualWithEpsilon(expected, 10) {
		t.Errorf("0.2+0.2 != %f\nGot %f\n",
			expected.ToFloat64(), r1.ToFloat64())
	}

	r1 = NewRegisterWithIntsAndFracs(0x133, 0, 10) // 0.3
	r1.Add(r3)
	expected = NewRegisterWithFloat64(0.3 + 0.5)
	if !r1.EqualWithEpsilon(expected, 10) {
		t.Errorf("0.3+0.2 != %f\nGot %f\n",
			expected.ToFloat64(), r1.ToFloat64())
	}

}

func Test_Mult(t *testing.T) {
	r1 := NewRegisterWithFloat64(0.3)
	r2 := NewRegisterWithFloat64(0.4)
	expected := NewRegisterWithFloat64(0.3 * 0.4)

	r1.Mult(r2)
	if !r1.Equal(expected) {
		t.Errorf("Got %f, expected %f",
			r1.ToFloat64(), expected.ToFloat64())
	}

	r1 = NewRegisterWithFloat64(0.3)
	r2 = NewRegisterWithFloat64(-0.4)
	expected = NewRegisterWithFloat64(0.3 * -0.4)

	r1.Mult(r2)
	if !r1.Equal(expected) {
		t.Errorf("Got %f, expected %f",
			r1.ToFloat64(), expected.ToFloat64())
	}

	// Different Q-Formats
	r1 = NewRegisterWithIntsAndFracs(0b100111001, 0, 10)    // 0.3
	r2 = NewRegisterWithIntsAndFracs(0b110011001100, 1, 14) // 0.2
	expected = NewRegisterWithFloat64(0.3 * 0.2)

	r1.Mult(r2)
	if !r1.EqualWithEpsilon(expected, 10) {
		t.Errorf("0.3*0.2 != %f\nGot %f\n",
			expected.ToFloat64(), r1.ToFloat64())
	}
}

func Test_Abs(t *testing.T) {
	r := NewRegister(-123)
	r.Abs()
	expected := int32(123)
	if r.Value != expected {
		t.Errorf("Register.Abs() failed: Got %d, expected %d",
			r.Value, expected)
	}
}

func Test_BitLogic(t *testing.T) {
	r := NewRegister(0b1111)
	r.And(0b1010)
	expected := int32(0b1010)
	if r.Value != expected {
		t.Errorf("Register.And() failed: Got 0b%b, expected 0b%b",
			r.Value, expected)
	}

	r.Value = 0b0
	r.Or(0b1010)
	expected = int32(0b1010)
	if r.Value != expected {
		t.Errorf("Register.Or() failed: Got 0b%b, expected 0b%b",
			r.Value, expected)
	}

	r.Value = 0b1010
	r.Xor(0b1010)
	expected = int32(0)
	if r.Value != expected {
		t.Errorf("Register.Xor() failed: Got 0b%b, expected 0b%b",
			r.Value, expected)
	}

	r.Value = 0b1010
	r.Not(0b1010)
	expected = int32(0b1010 &^ 0b1010)
	if r.Value != expected {
		t.Errorf("Register.Not() failed: Got 0b%b, expected 0b%b",
			r.Value, expected)
	}
}

func Test_ChainedOperations(t *testing.T) {
	// Use floats
	r1 := NewRegisterWithFloat64(0.5)
	r2 := NewRegisterWithFloat64(0.3)
	r3 := NewRegisterWithFloat64(0.2)
	expected := NewRegisterWithFloat64((0.5 * 0.3) + 0.2)

	r1.Mult(r2).Add(r3)
	if !r1.Equal(expected) {
		t.Errorf("Chain operation (0.5*0.3)+0.2 != %f\nGot %f\n",
			expected.ToFloat64(), r1.ToFloat64())
	}

	// Try again, but with reduces precision
	r1.SetFloat64(0.5)
	r2q := r2.ToQFormat(1, 14)
	r3q := r3.ToQFormat(0, 10)
	x2 := NewRegisterWithIntsAndFracs(r2q, 1, 14)
	x3 := NewRegisterWithIntsAndFracs(r3q, 0, 10)
	r1.Mult(x2).Add(x3)
	if !r1.EqualWithEpsilon(expected, 10) {
		t.Errorf("Chain operation (0.5*0.3)+0.2 != %f\nGot %f\n",
			expected.ToFloat64(), r1.ToFloat64())
	}
}
