package dsp

import (
	"fmt"
	"math"
	"testing"

	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/utils"
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
		state.ACC = float32(utils.QFormatToFloat64(0b1111, 1, 23))
		op := base.Ops[0x0E]

		op.Args[1].RawValue = 0b1
		expected := uint32(op.Args[1].RawValue)
		applyOp(op, state)

		ACC_bits := utils.Float64ToQFormat(float64(state.ACC), 1, 23)
		if ACC_bits != expected {
			t.Errorf("ACC & C != 0b%b. Got 0b%b", expected, ACC_bits)
		}
	})

	t.Run("OR", func(t *testing.T) {
		state := NewState()
		state.ACC = 0
		op := base.Ops[0x0F]
		// Set LSB
		op.Args[1].RawValue = 0b1
		expected := uint32(op.Args[1].RawValue)
		applyOp(op, state)

		ACC_bits := utils.Float64ToQFormat(float64(state.ACC), 1, 23)
		if ACC_bits != expected {
			t.Errorf("ACC | C != 0b%b. Got 0b%b", expected, ACC_bits)
		}

		// Set MSB
		state.ACC = 0
		op.Args[1].RawValue = 0b1 << 23
		expected = uint32(op.Args[1].RawValue)

		applyOp(op, state)
		ACC_bits = utils.Float64ToQFormat(float64(state.ACC), 1, 23)
		if ACC_bits != expected {
			t.Errorf("ACC | C != 0b%b. Got 0b%b", expected, ACC_bits)
		}
	})

	t.Run("XOR", func(t *testing.T) {
		state := NewState()
		state.ACC = float32(utils.QFormatToFloat64(0b1111, 1, 23))
		op := base.Ops[0x10]

		op.Args[1].RawValue = 0b1
		expected := 0b1111 ^ uint32(op.Args[1].RawValue)
		applyOp(op, state)

		ACC_bits := utils.Float64ToQFormat(float64(state.ACC), 1, 23)
		if ACC_bits != expected {
			t.Errorf("ACC ^ C != 0b%b. Got 0b%b", expected, ACC_bits)
		}
	})

	t.Run("LOG", func(t *testing.T) {
		state := NewState()
		state.ACC = math.Float32frombits(0x200)
		op := base.Ops[0x0B]
		op.Args[0].RawValue = 0x200
		op.Args[1].RawValue = 0x2000
		expected := float32(0.5*(math.Log10(float64(state.ACC))/math.Log10(2.0)/16.0) + 0.8)

		applyOp(op, state)
		// We need a much larger epsilon here as the LOG operation has such a low precision.
		// FIXME: Double check this (20220201 handegar)
		if math.Abs(float64(state.ACC)-float64(expected)) > 8.0 {
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
		op.Args[0].RawValue = base.ADCL // addr, 6bit
		op.Args[2].RawValue = 0x4000    // C, s1.14

		expected := state.ACC + float32(state.Registers[base.ADCL].(float64)*1.0)
		applyOp(op, state)

		// We need a much larger epsilon here it seems.
		// FIXME: Double check this (20220201 handegar)
		if math.Abs(float64(state.ACC)-float64(expected)) > s1_14_epsilon {
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
			t.Errorf("Expected ACC=%f, got %f\n", state.ACC, expected)
		}

		op.Args[0].RawValue = 0x0
		op.Args[1].RawValue = 0x200 // 1.0

		expected = state.ACC + state.DelayRAM[state.Registers[base.ADDR_PTR].(int)]*1.0
		applyOp(op, state)

		// We need a much larger epsilon here it seems.
		// FIXME: Double check this (20220201 handegar)
		if math.Abs(float64(state.ACC)-float64(expected)) > 0.2 {
			t.Errorf("Expected ACC=%f, got %f\n", state.ACC, expected)
		}
	})

	t.Run("WRA", func(t *testing.T) {
		state := NewState()
		preACC := float32(123.0)
		state.ACC = preACC
		op := base.Ops[0x02]
		state.DelayRAM[0x3e8] = 0.0
		op.Args[0].RawValue = 0x3e8 // ram addr
		op.Args[1].RawValue = utils.Float64ToQFormat(1.5, 1, 9)

		expected := state.ACC * 1.5
		applyOp(op, state)

		// We need a much larger epsilon here it seems.
		// FIXME: Double check this (20220201 handegar)
		if math.Abs(float64(preACC-state.DelayRAM[0x3e8])) > s1_14_epsilon {
			t.Errorf("Expected RAM[0x3e8]=%f, got %f\n", preACC, state.DelayRAM[0x3e8])
		}

		if math.Abs(float64(state.ACC)-float64(expected)) > s1_14_epsilon {
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

		if math.Abs(float64(state.ACC)-float64(expected)) > s1_14_epsilon {
			t.Errorf("Expected ACC=%f, got %f\n", expected, state.ACC)
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
		if state.Registers[base.SIN0_RATE].(int) != 100 {
			t.Fatalf("Expected Sin0.Freq=100, got %d", state.Registers[base.SIN0_RATE].(int))
		}
		if state.Registers[base.SIN0_RANGE].(int) != 10 {
			t.Fatalf("Expected Sin0.Amplitude=10, got %d", state.Registers[base.SIN0_RANGE].(int))
		}

		op.Args[0].RawValue = 0x0a // amp=10
		op.Args[1].RawValue = 0x64 // freq=100
		op.Args[2].RawValue = 0x1  // Sin1

		applyOp(op, state)
		if state.Registers[base.SIN1_RATE].(int) != 100 {
			t.Fatalf("Expected Sin1.Freq=100, got %d", state.Registers[base.SIN1_RATE].(int))
		}
		if state.Registers[base.SIN1_RANGE].(int) != 10 {
			t.Fatalf("Expected Sin1.Amplitude=10, got %d", state.Registers[base.SIN1_RANGE].(int))
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

		if state.Registers[base.RAMP0_RATE].(int) != 100 {
			t.Fatalf("Expected Ramp0.Freq=100, got %d", state.Registers[base.RAMP0_RATE].(int))
		}
		if state.Registers[base.RAMP0_RANGE].(int) != 4096 {
			t.Fatalf("Expected Ramp0.Amplitude=4096, got %d", state.Registers[base.RAMP0_RANGE].(int))
		}

		op.Args[0].RawValue = 0x1 // 0b00 -> 2048
		op.Args[1].RawValue = 0x0
		op.Args[2].RawValue = 0x64 // 100
		op.Args[3].RawValue = 0x1  // Rmp1
		applyOp(op, state)

		if state.Registers[base.RAMP1_RATE].(int) != 100 {
			t.Fatalf("Expected Ramp1.Freq=100, got %d", state.Registers[base.RAMP1_RATE].(int))
		}
		if state.Registers[base.RAMP1_RANGE].(int) != 2048 {
			t.Fatalf("Expected Ramp1.Amplitude=4096, got %d", state.Registers[base.RAMP1_RANGE].(int))
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
		state.ACC = 0.0
		state.Registers[base.SIN0_RANGE] = 1
		state.Registers[base.SIN1_RANGE] = 1
		for i := 0; i < 1000; i++ {
			state.DelayRAM[1000+i] = 1.0 + float32(i)
		}
		op := base.Ops[0x14]
		op.Name = "CHO RDA"
		op.Args[0].RawValue = 0x3e8 // 1000
		op.Args[1].RawValue = 0x00  // Type (SIN0)
		op.Args[3].RawValue = 0x0   // Flags

		applyOp(op, state)
		if float2Compare(float32(state.ACC), state.DelayRAM[1000+int(GetLFOValue(0, state))]) {
			t.Errorf("Expected ACC to be %f, got %f\n", state.DelayRAM[1000+int(GetLFOValue(0, state))], state.ACC)
		}

		state.Registers[base.SIN0_RANGE] = 2
		state.ACC = 0.0
		state.Sin0State.Angle = 3.1415 / 2.0
		applyOp(op, state)
		if float2Compare(float32(state.ACC), state.DelayRAM[1000+int(GetLFOValue(0, state))]) {
			t.Errorf("Expected ACC to be %f, got %f\n", state.DelayRAM[1000+int(GetLFOValue(0, state))], state.ACC)
		}

		state.ACC = 0.0
		op.Args[1].RawValue = 0x01 // Type (SIN1)
		state.Sin1State.Angle = 0.0
		applyOp(op, state)
		if float2Compare(float32(state.ACC), state.DelayRAM[1000+int(GetLFOValue(1, state))]) {
			t.Errorf("Expected ACC to be %f, got %f\n", state.DelayRAM[1000+int(GetLFOValue(1, state))], state.ACC)
		}

		state.Registers[base.SIN0_RANGE] = 2
		state.ACC = 0.0
		state.Sin1State.Angle = 3.1415 / 2.0
		applyOp(op, state)
		if float2Compare(float32(state.ACC), state.DelayRAM[1000+int(GetLFOValue(1, state))]) {
			t.Errorf("Expected ACC to be %f, got %f\n", state.DelayRAM[1000+int(GetLFOValue(1, state))], state.ACC)
		}

		// Ramp 0/1

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
		state.Registers[base.SIN0_RANGE] = int(1)
		applyOp(op, state)
		if float2Compare(state.ACC, float32(math.Sin(state.Sin0State.Angle)*float64(state.Registers[base.SIN0_RANGE].(int)))) {
			t.Errorf("Expected ACC=0,0, got %f\n", state.ACC)
		}

		// SIN1
		op.Args[1].RawValue = 0x1
		state.Sin1State.Angle = 3.14 / 2.0
		state.Registers[base.SIN1_RANGE] = int(1)
		applyOp(op, state)
		if float2Compare(state.ACC, float32(math.Sin(state.Sin1State.Angle)*float64(state.Registers[base.SIN1_RANGE].(int)))) {
			t.Errorf("Expected ACC=0,0, got %f\n", state.ACC)
		}

		// RMP0
		op.Args[1].RawValue = 0x2
		state.Ramp0State.Value = 1.23
		state.Registers[base.RAMP0_RANGE] = int(2)
		applyOp(op, state)
		if float2Compare(state.ACC, float32(state.Ramp0State.Value*float64(state.Registers[base.RAMP0_RANGE].(int)))) {
			t.Errorf("Expected ACC=0,0, got %f\n", state.ACC)
		}

		// RMP1
		op.Args[1].RawValue = 0x3
		state.Ramp1State.Value = 1.23
		state.Registers[base.RAMP1_RANGE] = int(2)
		applyOp(op, state)
		if float2Compare(state.ACC, float32(state.Ramp1State.Value*float64(state.Registers[base.RAMP1_RANGE].(int)))) {
			t.Errorf("Expected ACC=0,0, got %f\n", state.ACC)
		}
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
