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
		state.ACC = 1.0
		op := base.Ops[0x0D]
		op.Args[0].RawValue = 0x200
		op.Args[1].RawValue = 0x2000
		expected := state.ACC*0.5 + 0.5

		applyOp(op, state)
		if float2Compare(state.ACC, expected) {
			t.Errorf("SOF: ACC*C+D != %f. Got %f", expected, state.ACC)
		}
	})

	t.Run("AND", func(t *testing.T) {
		state := NewState()
		state.ACC = math.Float32frombits(0xFFFFFF)
		op := base.Ops[0x0E]
		op.Args[0].RawValue = 0xFFF
		expected := math.Float32frombits(0xFFF000)

		applyOp(op, state)
		if float2Compare(state.ACC, expected) {
			t.Errorf("ACC & C != %f. Got %f", expected, state.ACC)
		}
	})

	t.Run("OR", func(t *testing.T) {
		state := NewState()
		state.ACC = math.Float32frombits(0b10101010)
		op := base.Ops[0x0F]
		op.Args[0].RawValue = 0b01010101
		expected := math.Float32frombits(0b11111111)

		applyOp(op, state)
		if float2Compare(state.ACC, expected) {
			t.Errorf("ACC | C != %f. Got %f", expected, state.ACC)
		}
	})

	t.Run("XOR", func(t *testing.T) {
		state := NewState()
		state.ACC = math.Float32frombits(0b1010)
		op := base.Ops[0x10]
		op.Args[0].RawValue = 0b0101
		expected := math.Float32frombits(0b101)

		applyOp(op, state)
		if float2Compare(state.ACC, expected) {
			t.Errorf("ACC ^ C != %f. Got %f", expected, state.ACC)
		}
	})

	t.Run("LOG", func(t *testing.T) {
		state := NewState()
		state.ACC = math.Float32frombits(0x200)
		op := base.Ops[0x0B]
		op.Args[0].RawValue = 0x200
		op.Args[1].RawValue = 0x2000
		expected := float32(0.5*math.Log2(math.Abs(float64(state.ACC))) + 0.8)

		applyOp(op, state)
		// We need a much larger epsilon here as the LOG operation has such a low precision.
		// FIXME: Double check this (20220201 handegar)
		if math.Abs(float64(state.ACC)-float64(expected)) > 0.31 {
			t.Errorf("C * LOG(|ACC|) + D != %f. Got %f (%f diff)",
				expected, state.ACC, state.ACC-expected)
		}
	})

	t.Run("EXP", func(t *testing.T) {
		state := NewState()
		state.ACC = math.Float32frombits(0x200)
		op := base.Ops[0x0C]
		op.Args[0].RawValue = 0x200
		op.Args[1].RawValue = 0x3333
		expected := float32(0.8*math.Exp(float64(state.ACC)) + 0.5)

		applyOp(op, state)
		if float2Compare(state.ACC, expected) {
			t.Errorf("C * EXP(ACC) + D != %f. Got %f (%f diff)",
				expected, state.ACC, state.ACC-expected)
		}
	})

	t.Run("SKP", func(t *testing.T) {
		state := NewState()
		state.ACC = math.Float32frombits(0x200)
		op := base.Ops[0x11]
		op.Args[0].RawValue = 0x0
		op.Args[1].RawValue = 0x4

		op.Args[2].RawValue = 0x10 // RUN
		expected := state.IP + 0x4
		applyOp(op, state)
		if state.IP != expected {
			t.Errorf("Expected SKP RUN IP=%d, got %d",
				expected, state.IP)
		}

		op.Args[2].RawValue = 0x1 // NEG
		state.ACC = -1.0
		state.IP = 0
		expected = state.IP + 0x4
		applyOp(op, state)
		if state.IP != expected {
			t.Errorf("Expected SKP NEG IP=%d, got %d",
				expected, state.IP)
		}

		op.Args[2].RawValue = 0b00010 // GEZ
		state.ACC = 1.0
		state.IP = 0
		expected = state.IP + 0x4
		applyOp(op, state)
		if state.IP != expected {
			t.Errorf("Expected SKP GEZ IP=%d, got %d",
				expected, state.IP)
		}

		op.Args[2].RawValue = 0b00100 // ZRO
		state.ACC = 0.0
		state.IP = 0
		expected = state.IP + 0x4
		applyOp(op, state)
		if state.IP != expected {
			t.Errorf("Expected SKP ZRO IP=%d, got %d",
				expected, state.IP)
		}

		op.Args[2].RawValue = 0b01000 // ZRC
		state.ACC = -1.0
		state.PACC = 1.0
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
		state.ACC = 1.0
		state.Registers[base.ADCL] = 123.0
		op := base.Ops[0x04]
		op.Args[0].RawValue = 0x14
		op.Args[2].RawValue = 0x4000

		expected := state.ACC + float32(state.Registers[base.ADCL].(float64)*1.0)
		applyOp(op, state)

		// We need a much larger epsilon here it seems.
		// FIXME: Double check this (20220201 handegar)
		if math.Abs(float64(state.ACC)-float64(expected)) > 0.0039 {
			t.Errorf("Expected ACC=%f, got %f\n", state.ACC, expected)
		}
	})

	t.Run("WRAX", func(t *testing.T) {
		state := NewState()
		state.ACC = 123.0
		op := base.Ops[0x06]
		op.Args[0].RawValue = 0x16
		op.Args[2].RawValue = 0x4000

		expected := 1.0 * 123.0
		applyOp(op, state)

		if state.Registers[base.DACL].(float64) != 123.0 {
			t.Errorf("Expected DACL=%f, got %f\n", 123.0, state.Registers[base.DACL].(float64))
		}

		// We need a much larger epsilon here it seems.
		// FIXME: Double check this (20220201 handegar)
		if math.Abs(float64(state.ACC)-float64(expected)) > 0.0039 {
			t.Errorf("Expected ACC=%f, got %f\n", state.ACC, expected)
		}
	})

	t.Run("MAXX", func(t *testing.T) {
		state := NewState()
		state.ACC = 1.0
		state.Registers[0x20] = 123.0 // REG0
		op := base.Ops[0x09]
		op.Args[0].RawValue = 0x20
		op.Args[2].RawValue = 0x4000

		expected := math.Max(float64(state.Registers[0x20].(float64))*1.0,
			math.Abs(float64(state.ACC)))
		applyOp(op, state)

		// We need a much larger epsilon here it seems.
		// FIXME: Double check this (20220201 handegar)
		if math.Abs(float64(state.ACC)-float64(expected)) > 0.0039 {
			t.Errorf("Expected ACC=%f, got %f\n", state.ACC, expected)
		}
	})

	t.Run("MULX", func(t *testing.T) {
		state := NewState()
		state.ACC = 1.0
		state.Registers[0x20] = 123.0 // REG0
		op := base.Ops[0x0A]
		op.Args[0].RawValue = 0x20

		expected := float64(state.Registers[0x20].(float64)) * float64(state.ACC)
		applyOp(op, state)

		// We need a much larger epsilon here it seems.
		// FIXME: Double check this (20220201 handegar)
		if math.Abs(float64(state.ACC)-float64(expected)) > 0.0039 {
			t.Errorf("Expected ACC=%f, got %f\n", state.ACC, expected)
		}
	})

	t.Run("RDFX", func(t *testing.T) {
		state := NewState()
		state.ACC = 1.0
		state.Registers[0x20] = 123.0 // REG0
		op := base.Ops[0x05]
		op.Args[0].RawValue = 0x20
		op.Args[2].RawValue = 0x4000

		expected := (float64(state.ACC)-state.Registers[0x20].(float64))*1.0 + state.Registers[0x20].(float64)
		applyOp(op, state)

		// We need a much larger epsilon here it seems.
		// FIXME: Double check this (20220201 handegar)
		if math.Abs(float64(state.ACC)-float64(expected)) > 0.0039 {
			//if float2Compare(state.ACC, float32(expected)) {
			t.Errorf("Expected ACC=%f, got %f\n", state.ACC, expected)
		}
	})

	t.Run("WRLX", func(t *testing.T) {
		state := NewState()
		state.ACC = 1.0
		state.PACC = 2.0
		state.Registers[0x20] = 123.0 // REG0
		op := base.Ops[0x08]
		op.Args[0].RawValue = 0x20
		op.Args[2].RawValue = 0x4000

		expected := (state.PACC-state.ACC)*1.0 + state.PACC
		applyOp(op, state)

		// We need a much larger epsilon here it seems.
		// FIXME: Double check this (20220201 handegar)
		//if math.Abs(float64(state.ACC)-float64(expected)) > 0.0039 {
		if float2Compare(state.ACC, float32(expected)) {
			t.Errorf("Expected ACC=%f, got %f\n", state.ACC, expected)
		}

		if state.Registers[0x20].(float64) != 1.0 {
			t.Errorf("Expected REG0=%f, got %f\n", 1.0, state.Registers[0x20].(float64))
		}
	})

	t.Run("WRHX", func(t *testing.T) {
		state := NewState()
		state.ACC = 1.0
		state.PACC = 2.0
		state.Registers[0x20] = 123.0 // REG0
		op := base.Ops[0x07]
		op.Args[0].RawValue = 0x20
		op.Args[2].RawValue = 0x4000

		expected := state.ACC*1.0 + state.PACC
		applyOp(op, state)

		// We need a much larger epsilon here it seems.
		// FIXME: Double check this (20220201 handegar)
		//if math.Abs(float64(state.ACC)-float64(expected)) > 0.0039 {
		if float2Compare(state.ACC, float32(expected)) {
			t.Errorf("Expected ACC=%f, got %f\n", state.ACC, expected)
		}

		if state.Registers[0x20].(float64) != 1.0 {
			t.Errorf("Expected REG0=%f, got %f\n", 1.0, state.Registers[0x20].(float64))
		}
	})
}

func Test_DelayRAMOps(t *testing.T) {
	t.Run("RDA", func(t *testing.T) {
		state := NewState()
		state.ACC = 1.0
		state.DelayRAM[0x3e8] = 123.0
		op := base.Ops[0x0]
		op.Args[0].RawValue = 0x3e8
		op.Args[1].RawValue = 0x300

		expected := state.ACC + state.DelayRAM[0x3e8]*1.5
		applyOp(op, state)

		// We need a much larger epsilon here it seems.
		// FIXME: Double check this (20220201 handegar)
		if math.Abs(float64(state.ACC)-float64(expected)) > 0.2 {
			//if float2Compare(state.ACC, float32(expected)) {
			t.Errorf("Expected ACC=%f, got %f\n", state.ACC, expected)
		}
	})

	t.Run("RMPA", func(t *testing.T) {
		state := NewState()
		state.ACC = 1.0
		state.Registers[base.ADDR_PTR] = 99
		state.DelayRAM[0x3e8] = 123.0
		op := base.Ops[0x01]
		op.Args[0].RawValue = 0x0
		op.Args[1].RawValue = 0x300

		expected := state.ACC + state.DelayRAM[state.Registers[base.ADDR_PTR].(int)]*1.5
		applyOp(op, state)

		// We need a much larger epsilon here it seems.
		// FIXME: Double check this (20220201 handegar)
		if math.Abs(float64(state.ACC)-float64(expected)) > 0.2 {
			//if float2Compare(state.ACC, float32(expected)) {
			t.Errorf("Expected ACC=%f, got %f\n", state.ACC, expected)
		}
	})

	t.Run("WRA", func(t *testing.T) {
		state := NewState()
		state.ACC = 123.0
		op := base.Ops[0x02]
		op.Args[0].RawValue = 0x3e8
		op.Args[1].RawValue = 0x300

		expected := state.ACC * 1.5
		applyOp(op, state)

		// We need a much larger epsilon here it seems.
		// FIXME: Double check this (20220201 handegar)
		if state.DelayRAM[0x3e8] == state.ACC {
			t.Errorf("Expected RAM[0x3e8]=%f, got %f\n", state.ACC, state.DelayRAM[0x3e8])
		}

		if state.ACC == expected {
			t.Errorf("Expected ACC=%f, got %f\n", expected, state.ACC)
		}
	})

	t.Run("WRAP", func(t *testing.T) {
		state := NewState()
		state.ACC = 123.0
		state.LR = 2.0
		op := base.Ops[0x03]
		op.Args[0].RawValue = 0x3e8
		op.Args[1].RawValue = 0x300

		expected := (state.ACC * 1.5) + state.LR
		applyOp(op, state)

		// We need a much larger epsilon here it seems.
		// FIXME: Double check this (20220201 handegar)
		if state.DelayRAM[0x3e8] == state.ACC {
			t.Errorf("Expected RAM[0x3e8]=%f, got %f\n", state.ACC, state.DelayRAM[0x3e8])
		}

		if state.ACC == expected {
			t.Errorf("Expected ACC=%f, got %f\n", expected, state.ACC)
		}
	})
}

func Test_LFOOps(t *testing.T) {
	t.Run("WLDS", func(t *testing.T) {
		state := NewState()
		op := base.Ops[0x12]
		op.Name = "WLDS"
		op.Args[0].RawValue = 0x0a
		op.Args[1].RawValue = 0x64
		op.Args[2].RawValue = 0x0
		op.Args[3].RawValue = 0x0

		applyOp(op, state)

		// FIXME: Check SIN0-LFO (20220201 handegar)
		// FIXME: Check SIN1-LFO (20220201 handegar)

	})

	t.Run("WLDR", func(t *testing.T) {
		state := NewState()
		op := base.Ops[0x12]
		op.Name = "WLDR"
		op.Args[0].RawValue = 0x0a
		op.Args[1].RawValue = 0x64
		op.Args[2].RawValue = 0x1
		op.Args[3].RawValue = 0x0

		applyOp(op, state)

		// FIXME: Check RMP0-LFO (20220201 handegar)
		// FIXME: Check RMP1-LFO (20220201 handegar)
	})

	t.Run("JAM", func(t *testing.T) {
		state := NewState()
		op := base.Ops[0x13]
		op.Args[0].RawValue = 0x0a
		op.Args[1].RawValue = 0x0
		op.Args[2].RawValue = 0x1

		applyOp(op, state)
		// FIXME: Check RMP0-LFO (20220201 handegar)

		op.Args[1].RawValue = 0x0
		applyOp(op, state)
		// FIXME: Check RMP1-LFO (20220201 handegar)
	})

	t.Run("CHO RDA", func(t *testing.T) {
		state := NewState()
		state.ACC = 123.0
		state.LR = 2.0
		op := base.Ops[0x14]
		op.Name = "CHO RDA"
		op.Args[0].RawValue = 0x3e8
		op.Args[1].RawValue = 0x0
		op.Args[2].RawValue = 0x0
		op.Args[3].RawValue = 0x0
		op.Args[4].RawValue = 0x0

		applyOp(op, state)
		// FIXME: Verify result (20220201 handegar)
		fmt.Println("CHO RDA not verified")
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
		// SIN0
		op.Args[0].RawValue = 0x0
		op.Args[1].RawValue = 0x0
		op.Args[2].RawValue = 0x0
		op.Args[3].RawValue = 0x2
		op.Args[4].RawValue = 0x3

		// FIXME: Verify results (20220201 handegar)

		applyOp(op, state)
		fmt.Println("CHO RDAL, SIN0 not verified")

		// SIN1
		op.Args[1].RawValue = 0x1
		applyOp(op, state)
		fmt.Println("CHO RDAL, SIN1 not verified")

		// RMP0
		op.Args[1].RawValue = 0x2
		applyOp(op, state)
		fmt.Println("CHO RDAL, RMP0 not verified")

		// RMP1
		op.Args[1].RawValue = 0x3
		applyOp(op, state)
		fmt.Println("CHO RDAL, RMP1 not verified")
	})

}

func Test_PseudoOps(t *testing.T) {
	t.Run("CLR", func(t *testing.T) {
		state := NewState()
		state.ACC = 123.0
		op := base.Ops[0x0e]
		op.Name = "CLR"
		op.Args[0].RawValue = 0

		applyOp(op, state)
		if float2Compare(state.ACC, float32(0.0)) {
			t.Errorf("Expected ACC=0,0, got %f\n", state.ACC)
		}
	})

	t.Run("NOT", func(t *testing.T) {
		state := NewState()
		state.ACC = math.Float32frombits(0xFFFF)
		op := base.Ops[0x10]
		op.Name = "NOT"
		op.Args[0].RawValue = 0xFFFFFFFF

		expected := math.Float32frombits(0x10000)

		applyOp(op, state)
		if float2Compare(state.ACC, expected) {
			t.Errorf("Expected ACC=0x%x, got %f\n", expected, state.ACC)
		}
	})

	t.Run("ABSA", func(t *testing.T) {
		state := NewState()
		state.ACC = -123.0
		op := base.Ops[0x09]
		op.Name = "ABSA"

		expected := 123.0

		applyOp(op, state)
		if float2Compare(state.ACC, float32(expected)) {
			t.Errorf("Expected ACC=0x%x, got %f\n", expected, state.ACC)
		}
	})

	t.Run("LDAX", func(t *testing.T) {
		state := NewState()
		state.ACC = -1.0
		state.Registers[0x20] = 123.0 // REG0
		op := base.Ops[0x05]
		op.Name = "LDAX"
		op.Args[0].RawValue = 0x20
		expected := 123.0

		applyOp(op, state)
		if float2Compare(state.ACC, float32(expected)) {
			t.Errorf("Expected ACC=0x%x, got %f\n", expected, state.ACC)
		}
	})

}