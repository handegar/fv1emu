package dsp

import (
	"fmt"
	"math"
	"testing"

	"github.com/handegar/fv1emu/base"
)

var s1_14_epsilon float64 = 0.00006103516
var s10_epsilon float64 = 0.0009765625

func float2Compare(a float32, b float32) bool {
	return math.Abs(float64(a-b)) > s1_14_epsilon
}

func float1Compare(a float32, b float32) bool {
	return math.Abs(float64(a-b)) > s10_epsilon
}

func Test_AccumulatorOps(t *testing.T) {
	t.Run("SOF", func(t *testing.T) {
		state := NewState()
		c := NewRegisterWithFloat64(0.3)
		d := NewRegisterWithFloat64(0.2)
		state.ACC.SetFloat64(0.5)

		op := base.Ops[0x0D]
		op.Args[0].RawValue = d.ToQFormat(0, 10)
		op.Args[1].RawValue = c.ToQFormat(1, 14)
		expected := NewRegisterWithFloat64((0.5 * 0.3) + 0.2)

		// C * ACC + D
		applyOp(op, state)

		if !state.ACC.EqualWithEpsilon(expected, 10) {
			t.Errorf("SOF: ACC*C+D != 0x%x. Got 0x%x\n"+
				"expected=%f\n     got=%f",
				expected.Value, state.ACC.Value,
				expected.ToFloat64(), state.ACC.ToFloat64())
		}
	})

	t.Run("AND", func(t *testing.T) {
		state := NewState()
		state.ACC.SetWithIntsAndFracs(0b1111, 0, 23)
		op := base.Ops[0x0E]

		op.Args[1].RawValue = 0b1
		expected := state.ACC.Value & op.Args[1].RawValue
		applyOp(op, state)

		if state.ACC.Value != expected {
			t.Errorf("ACC & C != 0b%b. Got 0b%b", expected, state.ACC.Value)
		}
	})

	t.Run("OR", func(t *testing.T) {
		state := NewState()
		state.ACC.Clear()
		op := base.Ops[0x0F]
		// Set LSB
		op.Args[1].RawValue = 0b1
		expected := state.ACC.Value | op.Args[1].RawValue
		applyOp(op, state)

		if state.ACC.Value != expected {
			fmt.Printf("b%b | 0b%b\n", state.ACC.Value, op.Args[1].RawValue)
			t.Errorf("ACC | C != 0b%b. Got 0b%b", expected, state.ACC.Value)
		}

		// Set MSB
		state.ACC.Clear()
		op.Args[1].RawValue = 0b1 << 23
		expected = state.ACC.Value | op.Args[1].RawValue

		applyOp(op, state)
		if state.ACC.Value != expected {
			t.Errorf("ACC | C != 0b%b. Got 0b%b", expected, state.ACC.Value)
		}
	})

	t.Run("XOR", func(t *testing.T) {
		state := NewState()
		state.ACC.SetWithIntsAndFracs(0b1111, 0, 23)
		op := base.Ops[0x10]

		op.Args[1].RawValue = 0b1
		expected := state.ACC.Value ^ op.Args[1].RawValue
		applyOp(op, state)

		if state.ACC.Value != expected {
			t.Errorf("ACC ^ C != 0b%b. Got 0b%b", expected, state.ACC.Value)
		}
	})

	t.Run("LOG", func(t *testing.T) {
		state := NewState()
		state.ACC.SetFloat64(0.3)
		op := base.Ops[0x0B]
		c := NewRegisterWithFloat64(0.5)
		d := NewRegisterWithFloat64(0.8)

		op.Args[0].RawValue = d.ToQFormat(4, 6)
		op.Args[1].RawValue = c.ToQFormat(1, 14)

		expectedF := 0.5*(math.Log10(0.3)/math.Log10(2.0)/16.0) + 0.8
		expected := NewRegisterWithFloat64(expectedF)

		applyOp(op, state)
		// We need a much larger epsilon here as the LOG operation has such a low precision.
		// FIXME: Double check this (20220201 handegar)
		if !state.ACC.EqualWithEpsilon(expected, 10) {
			t.Errorf("C * LOG(|ACC|) + D != %f. Got %f",
				expected.ToFloat64(), state.ACC.ToFloat64())
		}
	})

	t.Run("EXP", func(t *testing.T) {
		state := NewState()
		state.ACC.SetFloat64(0.7)
		op := base.Ops[0x0C]
		c := NewRegisterWithFloat64(0.5)
		d := NewRegisterWithFloat64(0.8)
		op.Args[0].RawValue = d.ToQFormat(0, 10)
		op.Args[1].RawValue = c.ToQFormat(1, 14)
		expectedF := 0.5*math.Exp(0.7) + 0.8
		expected := NewRegisterWithFloat64(expectedF)

		applyOp(op, state)

		// FIXME: Due to precision issues the result might
		// vary quite a bit when dealing with
		// exponentials. How can we properly test this?
		// (20220216 handegar)
		if !state.ACC.EqualWithEpsilon(expected, 10) {
			t.Errorf("C * EXP(ACC) + D != %f. Got %f",
				expected.ToFloat64(), state.ACC.ToFloat64())
		}
	})

	t.Run("SKP", func(t *testing.T) {
		state := NewState()
		state.ACC.Value = 0x200
		op := base.Ops[0x11]
		op.Args[0].RawValue = 0x0
		op.Args[1].RawValue = 0x4

		op.Args[2].RawValue = base.SKP_RUN
		expected := state.IP
		applyOp(op, state)
		if state.IP != expected {
			t.Errorf("Expected SKP RUN IP=%d, got %d",
				expected, state.IP)
		}

		state.RUN_FLAG = true
		op.Args[2].RawValue = base.SKP_RUN
		expected = state.IP + 4
		applyOp(op, state)
		if state.IP != expected {
			t.Errorf("Expected SKP RUN IP=%d, got %d",
				expected, state.IP)
		}

		op.Args[2].RawValue = base.SKP_NEG
		state.ACC.SetFloat64(-0.5)
		if !state.ACC.IsSigned() {
			t.Errorf("ACC.IsSigned() does not work")
		}
		state.IP = 0
		expected = state.IP + 0x4
		applyOp(op, state)
		if state.IP != expected {
			t.Errorf("Expected SKP NEG IP=%d, got %d",
				expected, state.IP)
		}

		op.Args[2].RawValue = base.SKP_GEZ
		state.ACC.SetFloat64(0.5)
		state.IP = 0
		expected = state.IP + 0x4
		applyOp(op, state)
		if state.IP != expected {
			t.Errorf("Expected SKP GEZ IP=%d, got %d",
				expected, state.IP)
		}

		op.Args[2].RawValue = base.SKP_ZRO
		state.ACC.Clear()
		state.IP = 0
		expected = state.IP + 0x4
		applyOp(op, state)
		if state.IP != expected {
			t.Errorf("Expected SKP ZRO IP=%d, got %d",
				expected, state.IP)
		}

		op.Args[2].RawValue = base.SKP_ZRC
		state.ACC.SetFloat64(-0.5)
		state.PACC.SetFloat64(0.3)
		state.IP = 0
		expected = state.IP + 0x4
		applyOp(op, state)
		if state.IP != expected {
			t.Errorf("Expected SKP ZRC IP=%d, got %d",
				expected, state.IP)
		}

	})
}

func Test_RegisterOps(t *testing.T) {
	t.Run("RDAX", func(t *testing.T) {
		state := NewState()
		state.ACC.SetFloat64(0.5)

		state.GetRegister(base.ADCL).SetFloat64(0.4)
		op := base.Ops[0x04]
		op.Args[0].RawValue = base.ADCL // addr, 6bit
		state.workReg1_14.SetFloat64(0.3)
		op.Args[2].RawValue = state.workReg1_14.ToQFormat(1, 14) // C, s1.14

		expected := NewRegisterWithFloat64(0.3*0.4 + 0.5)
		applyOp(op, state)

		if !state.ACC.EqualWithEpsilon(expected, 14) {
			t.Errorf("Expected ACC=%f, got %f\n",
				state.ACC.ToFloat64(), expected.ToFloat64())
		}
	})

	t.Run("WRAX", func(t *testing.T) {
		state := NewState()
		a := NewRegisterWithFloat64(1.0)
		state.ACC.SetFloat64(1.0)
		op := base.Ops[0x06]
		op.Args[0].RawValue = base.DACL
		c := NewRegisterWithFloat64(0.5)
		op.Args[2].RawValue = c.ToQFormat(1, 14)

		expected := NewRegisterWithFloat64(1.0 * 0.5)
		applyOp(op, state)

		if !state.GetRegister(base.DACL).Equal(a) {
			t.Errorf("Expected DACL=%f, got %f\n",
				a.ToFloat64(), state.GetRegister(base.DACL).ToFloat64())
		}

		if !state.ACC.Equal(expected) {
			t.Errorf("Expected ACC=%f, got %f\n",
				state.ACC.ToFloat64(), expected.ToFloat64())
		}
	})

	t.Run("MAXX", func(t *testing.T) {
		state := NewState()
		state.ACC.Value = 1
		state.GetRegister(0x20).Value = 123 // REG0
		op := base.Ops[0x09]
		op.Args[0].RawValue = 0x20
		op.Args[2].RawValue = 0x4000

		expected := int32(math.Max(float64(state.GetRegister(0x20).Value),
			math.Abs(float64(state.ACC.Value))))
		applyOp(op, state)

		if state.ACC.Value != expected {
			t.Errorf("Expected ACC=0x%x, got 0x%x\n", state.ACC.Value, expected)
		}
	})

	t.Run("MULX", func(t *testing.T) {
		state := NewState()
		state.ACC.SetFloat64(0.5)
		state.GetRegister(0x20).SetFloat64(0.2) // REG0
		op := base.Ops[0x0A]
		op.Args[0].RawValue = 0x20

		expected := NewRegisterWithFloat64(0.5 * 0.2)

		applyOp(op, state)

		if !state.ACC.Equal(expected) {
			t.Errorf("Expected ACC=%f, got %f\n", state.ACC.ToFloat64(), expected.ToFloat64())
		}
	})

	t.Run("RDFX", func(t *testing.T) {
		state := NewState()
		state.ACC.SetFloat64(1.0)
		state.GetRegister(0x20).SetFloat64(0.1) // REG0
		op := base.Ops[0x05]
		op.Args[0].RawValue = 0x20 // REG0
		c := NewRegisterWithFloat64(0.7)
		op.Args[2].RawValue = c.ToQFormat(1, 14)

		expected := NewRegisterWithFloat64((1.0-0.1)*0.7 + 0.1)
		applyOp(op, state)

		if !state.ACC.EqualWithEpsilon(expected, 14) {
			t.Errorf("Expected ACC=%f, got %f\n", state.ACC.ToFloat64(), expected.ToFloat64())
		}
	})

	t.Run("WRLX", func(t *testing.T) {
		state := NewState()

		a := NewRegisterWithFloat64(0.7)
		state.ACC.Copy(a)
		state.PACC.SetFloat64(0.6)
		state.GetRegister(0x20).SetFloat64(0.1) // REG0
		op := base.Ops[0x08]
		op.Args[0].RawValue = 0x20
		c := NewRegisterWithFloat64(0.5)
		op.Args[2].RawValue = c.ToQFormat(1, 14)

		expected := NewRegisterWithFloat64((0.6-0.7)*0.5 + 0.6)
		applyOp(op, state)

		if !state.ACC.Equal(expected) {
			t.Errorf("Expected ACC=0x%x, got 0x%x\n", state.ACC.Value, expected)
		}

		if !state.GetRegister(0x20).Equal(a) {
			t.Errorf("Expected REG0=%f, got %f\n", a.ToFloat64(),
				state.GetRegister(0x20).ToFloat64())
		}
	})

	t.Run("WRHX", func(t *testing.T) {
		state := NewState()
		a := NewRegisterWithFloat64(0.7)
		state.ACC.Copy(a)
		state.PACC.SetFloat64(0.6)
		state.GetRegister(0x20).SetFloat64(0.1) // REG0
		op := base.Ops[0x07]
		op.Args[0].RawValue = 0x20
		c := NewRegisterWithFloat64(0.5)
		op.Args[2].RawValue = c.ToQFormat(1, 14)

		expected := NewRegisterWithFloat64(0.7*0.5 + 0.6)
		applyOp(op, state)

		if !state.ACC.Equal(expected) {
			t.Errorf("Expected ACC=%f, got %f\n",
				state.ACC.ToFloat64(), expected.ToFloat64())
		}

		if !state.GetRegister(0x20).Equal(a) {
			t.Errorf("Expected REG0=%f, got %f\n",
				a.ToFloat64(),
				state.GetRegister(0x20).ToFloat64())
		}
	})
}

func Test_DelayRAMOps(t *testing.T) {
	t.Run("RDA", func(t *testing.T) {
		state := NewState()
		state.ACC.SetFloat64(0.6)
		s := NewRegisterWithFloat64(0.4)
		state.DelayRAM[0x3e8] = s.ToQFormat(0, 23)
		op := base.Ops[0x0]
		op.Args[0].RawValue = 0x3e8
		c := NewRegisterWithFloat64(0.5)
		op.Args[1].RawValue = c.ToQFormat(1, 9)

		expected := NewRegisterWithFloat64(0.4*0.5 + 0.6)
		applyOp(op, state)

		if !state.ACC.Equal(expected) {
			t.Errorf("Expected ACC=%f, got %f\n",
				state.ACC.ToFloat64(), expected.ToFloat64())
		}
	})

	t.Run("RMPA", func(t *testing.T) {
		state := NewState()
		state.ACC.Value = 1
		state.GetRegister(base.ADDR_PTR).Value = 99
		state.DelayRAM[0x3e8] = 123
		op := base.Ops[0x01]
		op.Args[0].RawValue = 0x0
		op.Args[1].RawValue = 0x300

		expected :=
			state.ACC.Value +
				state.DelayRAM[state.GetRegister(base.ADDR_PTR).Value]*op.Args[1].RawValue
		applyOp(op, state)

		// We need a much larger epsilon here it seems.
		// FIXME: Double check this (20220201 handegar)
		if state.ACC.Value != expected {
			t.Errorf("Expected ACC=0x%x, got 0x%x\n", state.ACC.Value, expected)
		}

		op.Args[0].RawValue = 0x0
		op.Args[1].RawValue = 0x200 // 1.0

		expected =
			state.ACC.Value +
				state.DelayRAM[state.GetRegister(base.ADDR_PTR).Value]*op.Args[1].RawValue
		applyOp(op, state)

		// We need a much larger epsilon here it seems.
		// FIXME: Double check this (20220201 handegar)
		if state.ACC.Value != expected {
			t.Errorf("Expected ACC=0x%x, got 0x%x\n", state.ACC.Value, expected)
		}
	})

	t.Run("WRA", func(t *testing.T) {
		state := NewState()
		preACC := NewRegisterWithFloat64(0.5)
		state.ACC.Copy(preACC)
		op := base.Ops[0x02]
		s := NewRegisterWithFloat64(0.6)
		state.DelayRAM[0x3e8] = s.ToQFormat(0, 23)
		op.Args[0].RawValue = 0x3e8 // ram addr
		c := NewRegisterWithFloat64(0.7)
		op.Args[1].RawValue = c.ToQFormat(1, 9)

		expected := NewRegisterWithFloat64(0.5 * 0.7)
		applyOp(op, state)

		if state.DelayRAM[0x3e8] != preACC.ToQFormat(0, 23) {
			t.Errorf("Expected RAM[0x3e8]=%b, got %b\n",
				preACC.ToQFormat(0, 23), state.DelayRAM[0x3e8])
		}

		if !state.ACC.EqualWithEpsilon(expected, 9) {
			t.Errorf("Expected ACC=%f, got %f\n",
				expected.ToFloat64(), state.ACC.ToFloat64())
		}
	})

	t.Run("WRAP", func(t *testing.T) {
		state := NewState()
		state.ACC.SetFloat64(0.5)
		state.LR.SetFloat64(0.3)
		op := base.Ops[0x03]
		op.Args[0].RawValue = 0x3e8

		c := NewRegisterWithFloat64(0.2)
		op.Args[1].RawValue = c.ToQFormat(1, 9)

		expected := NewRegisterWithFloat64(0.5*0.2 + 0.3)

		applyOp(op, state)

		if state.DelayRAM[0x3e8] == state.ACC.ToQFormat(0, 23) {
			t.Errorf("Expected RAM[0x3e8]=0x%x, got 0x%x\n",
				state.ACC.Value, state.DelayRAM[0x3e8])
		}

		if state.ACC.Equal(expected) {
			t.Errorf("Expected ACC=%f, got %f\n",
				expected.ToFloat64(), state.ACC.ToFloat64())
		}
	})
}

func Test_LFOOps(t *testing.T) {
	t.Run("WLDS", func(t *testing.T) {
		state := NewState()
		op := base.Ops[0x12]
		op.Name = "WLDS"
		op.Args[0].RawValue = 0x0a // amp=10
		op.Args[1].RawValue = 0x64 // freq=100
		op.Args[2].RawValue = 0x0  // Sin0

		applyOp(op, state)
		if state.GetRegister(base.SIN0_RATE).Value != op.Args[1].RawValue {
			t.Fatalf("Expected Sin0.Freq=0x%x, got 0x%x",
				op.Args[1].RawValue, state.GetRegister(base.SIN0_RATE).Value)
		}
		if state.GetRegister(base.SIN0_RANGE).Value != op.Args[0].RawValue {
			t.Fatalf("Expected Sin0.Amplitude=0x%x, got 0x%x",
				op.Args[0].RawValue, state.GetRegister(base.SIN0_RANGE).Value)
		}

		op.Args[0].RawValue = 0x0a // amp=10
		op.Args[1].RawValue = 0x64 // freq=100
		op.Args[2].RawValue = 0x1  // Sin1

		applyOp(op, state)
		if state.GetRegister(base.SIN1_RATE).Value != op.Args[1].RawValue {
			t.Fatalf("Expected Sin1.Freq=0x%x, got 0x%x",
				op.Args[1].RawValue, state.GetRegister(base.SIN1_RATE).Value)
		}
		if state.GetRegister(base.SIN1_RANGE).Value != op.Args[0].RawValue {
			t.Fatalf("Expected Sin1.Amplitude=0x%x, got 0x%x",
				op.Args[0].RawValue, state.GetRegister(base.SIN1_RANGE).Value)
		}
	})

	t.Run("WLDR", func(t *testing.T) {
		state := NewState()
		op := base.Ops[0x12]
		op.Name = "WLDR"
		op.Args[0].RawValue = 0x0 // 0b00 -> 4096
		op.Args[1].RawValue = 0x0
		op.Args[2].RawValue = 0x64 // 100
		op.Args[3].RawValue = 0x0  // Rmp0

		applyOp(op, state)

		if state.GetRegister(base.RAMP0_RATE).Value != op.Args[2].RawValue {
			t.Fatalf("Expected Ramp0.Freq=0x%x, got 0x%x",
				op.Args[2].RawValue, state.GetRegister(base.RAMP0_RATE).Value)
		}
		if state.GetRegister(base.RAMP0_RANGE).Value != base.RampAmpValues[int(op.Args[0].RawValue)] {
			t.Fatalf("Expected Ramp0.Amplitude=%d, got %d",
				base.RampAmpValues[int(op.Args[0].RawValue)],
				state.GetRegister(base.RAMP0_RANGE).Value)
		}

		op.Args[0].RawValue = 0x1 // 0b00 -> 2048
		op.Args[1].RawValue = 0x0
		op.Args[2].RawValue = 0x64 // 100
		op.Args[3].RawValue = 0x1  // Rmp1
		applyOp(op, state)

		if state.GetRegister(base.RAMP1_RATE).Value != op.Args[2].RawValue {
			t.Fatalf("Expected Ramp1.Freq=0x%x, got 0x%x",
				op.Args[2].RawValue, state.GetRegister(base.RAMP1_RATE).Value)
		}
		if state.GetRegister(base.RAMP1_RANGE).Value != base.RampAmpValues[int(op.Args[0].RawValue)] {
			t.Fatalf("Expected Ramp1.Amplitude=%d, got %d",
				base.RampAmpValues[int(op.Args[0].RawValue)],
				state.GetRegister(base.RAMP1_RANGE).Value)
		}
	})

	t.Run("JAM", func(t *testing.T) {
		state := NewState()
		state.Ramp0State.Value = 123.0
		state.Ramp1State.Value = 234.0

		op := base.Ops[0x13]
		op.Args[0].RawValue = 0x0
		op.Args[1].RawValue = 0x0
		op.Args[2].RawValue = 0x0

		applyOp(op, state)
		if float2Compare(float32(state.Ramp0State.Value), float32(0.0)) {
			t.Errorf("Expected Ramo0State to be 0.0, got %f\n", state.Ramp0State.Value)
		}

		op.Args[1].RawValue = 0x1
		applyOp(op, state)
		if float2Compare(float32(state.Ramp1State.Value), float32(0.0)) {
			t.Errorf("Expected Ramp1State to be 0.0, got %f\n", state.Ramp1State.Value)
		}
	})

	t.Run("CHO RDA", func(t *testing.T) {
		state := NewState()
		state.ACC.Clear()
		state.GetRegister(base.SIN0_RANGE).Value = 125
		state.GetRegister(base.SIN1_RANGE).Value = 125
		for i := 0; i < 1000; i++ {
			state.DelayRAM[1000+i] = int32(1 + i)
		}
		op := base.Ops[0x14]
		op.Name = "CHO RDA"
		op.Args[0].RawValue = 0x3e8 // 1000
		op.Args[1].RawValue = 0x00  // Type (SIN0)
		op.Args[3].RawValue = 0x0   // Flags

		applyOp(op, state)
		if state.ACC.Value != state.DelayRAM[1000+int(GetLFOValue(0, state))] {
			t.Errorf("Expected ACC to be 0x%x, got 0x%x\n",
				state.DelayRAM[1000+int(GetLFOValue(0, state))],
				state.ACC.Value)
		}

		state.GetRegister(base.SIN0_RANGE).Value = 2
		state.ACC.Clear()
		state.Sin0State.Angle = 3.1415 / 2.0
		applyOp(op, state)
		if state.ACC.Value != state.DelayRAM[1000+int(GetLFOValue(0, state))] {
			t.Errorf("Expected ACC to be 0x%x, got 0x%x\n",
				state.DelayRAM[1000+int(GetLFOValue(0, state))],
				state.ACC.Value)
		}

		state.ACC.Clear()
		op.Args[1].RawValue = 0x01 // Type (SIN1)
		state.Sin1State.Angle = 0.0
		applyOp(op, state)
		if state.ACC.Value != state.DelayRAM[1000+int(GetLFOValue(1, state))] {
			t.Errorf("Expected ACC to be 0x%x, got 0x%x\n",
				state.DelayRAM[1000+int(GetLFOValue(1, state))],
				state.ACC.Value)
		}

		state.GetRegister(base.SIN0_RANGE).SetFloat64(2.0)
		state.ACC.Clear()
		state.Sin1State.Angle = 3.1415 / 2.0
		applyOp(op, state)
		if state.ACC.Value != state.DelayRAM[1000+int(GetLFOValue(1, state))] {
			t.Errorf("Expected ACC to be 0x%x, got 0x%x\n",
				state.DelayRAM[1000+int(GetLFOValue(1, state))],
				state.ACC.Value)
		}

		// FXIME: Ramp 0/1
		fmt.Println("CHO RDA  RMPx, ... not verified")
	})

	t.Run("CHO SOF", func(t *testing.T) {
		state := NewState()
		op := base.Ops[0x14]
		op.Name = "CHO SOF"
		op.Args[0].RawValue = 0x3e8
		op.Args[1].RawValue = 0x0
		op.Args[2].RawValue = 0x0
		op.Args[3].RawValue = 0x0
		op.Args[4].RawValue = 0x2

		applyOp(op, state)
		// FIXME: Verify result (20220201 handegar)
		fmt.Println("CHO SOF not verified")
	})

	t.Run("CHO RDAL", func(t *testing.T) {
		state := NewState()
		op := base.Ops[0x14]
		op.Name = "CHO RDAL"

		// Load SIN0 into ACC
		op.Args[0].RawValue = 0x0
		op.Args[1].RawValue = 0x0
		op.Args[2].RawValue = 0x0
		op.Args[3].RawValue = 0x2
		op.Args[4].RawValue = 0x3

		state.Sin0State.Angle = 3.14 / 2.0
		state.GetRegister(base.SIN0_RANGE).Value = 1
		applyOp(op, state)
		if float2Compare(float32(state.ACC.Value),
			float32(math.Sin(state.Sin0State.Angle)*float64(state.GetRegister(base.SIN0_RANGE).Value))) {
			t.Errorf("Expected ACC=0, got 0x%x\n", state.ACC.Value)
		}

		// SIN1
		op.Args[1].RawValue = 0x1
		state.Sin1State.Angle = 3.14 / 2.0
		state.GetRegister(base.SIN1_RANGE).Value = 1
		applyOp(op, state)
		if float2Compare(float32(state.ACC.Value),
			float32(math.Sin(state.Sin1State.Angle)*float64(state.GetRegister(base.SIN1_RANGE).Value))) {
			t.Errorf("Expected ACC=0, got 0x%x\n", state.ACC.Value)
		}

		// RMP0
		op.Args[1].RawValue = 0x2
		state.Ramp0State.Value = 1.23
		state.GetRegister(base.RAMP0_RANGE).Value = 2
		applyOp(op, state)
		if float2Compare(float32(state.ACC.Value),
			float32(state.Ramp0State.Value*float64(state.GetRegister(base.RAMP0_RANGE).Value))) {
			t.Errorf("Expected ACC=0, got 0x%x\n", state.ACC.Value)
		}

		// RMP1
		op.Args[1].RawValue = 0x3
		state.Ramp1State.Value = 1.23
		state.GetRegister(base.RAMP1_RANGE).Value = 2
		applyOp(op, state)
		if float2Compare(float32(state.ACC.Value),
			float32(state.Ramp1State.Value*float64(state.GetRegister(base.RAMP1_RANGE).Value))) {
			t.Errorf("Expected ACC=0, got 0x%x\n", state.ACC.Value)
		}
	})

}

func Test_PseudoOps(t *testing.T) {
	t.Run("CLR", func(t *testing.T) {
		state := NewState()
		state.ACC.Value = 123
		op := base.Ops[0x0e]
		op.Name = "CLR"
		op.Args[0].RawValue = 0

		applyOp(op, state)
		if state.ACC.Value != 0 {
			t.Errorf("Expected ACC=0, got 0x%x\n",
				state.ACC.Value)
		}
	})

	t.Run("NOT", func(t *testing.T) {
		state := NewState()
		state.ACC.Value = 0xFFFF
		op := base.Ops[0x10]
		op.Name = "NOT"
		op.Args[1].RawValue = 0xFFFF

		expected := int32(0xFFFF &^ 0xFFFF)

		applyOp(op, state)
		if state.ACC.Value != expected {
			t.Errorf("Expected ACC=0x%x, got 0x%x\n",
				expected, state.ACC.Value)
		}
	})

	t.Run("ABSA", func(t *testing.T) {
		state := NewState()
		state.ACC.SetFloat64(-123.0)
		op := base.Ops[0x09]
		op.Name = "ABSA"

		expected := NewRegisterWithFloat64(123.0)

		applyOp(op, state)
		if !state.ACC.Equal(expected) {
			t.Errorf("Expected ACC=0x%x, got 0x%x\n",
				expected.Value, state.ACC.Value)
		}
	})

	t.Run("LDAX", func(t *testing.T) {
		state := NewState()
		state.ACC.Value = -1
		state.GetRegister(0x20).Value = 123 // REG0
		op := base.Ops[0x05]
		op.Name = "LDAX"
		op.Args[0].RawValue = 0x20
		expected := int32(123)

		applyOp(op, state)
		if state.ACC.Value != expected {
			t.Errorf("Expected ACC=0x%x, got 0x%x\n",
				expected, state.ACC.Value)
		}
	})

}
