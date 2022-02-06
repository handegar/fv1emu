package utils

import (
	"math"
	"testing"
)

func floatTest(t *testing.T, intbits int, fracbits int, in uint32, expected float64, epsilon float64) {
	f := QFormatToFloat64(in, intbits, fracbits)
	if math.Abs(float64(f)-expected) > epsilon {
		if intbits == 0 {
			t.Fatalf("S.%d of '0x%x'/0b%b != %f (got %f)\n",
				fracbits, in, in, expected, f)
		} else {
			t.Fatalf("S%d.%d of '0x%x'/0b%b != %f (got %f)\n",
				intbits, fracbits, in, in, expected, f)
		}
	}
}

func Test_Real1ToFloat_S_10(t *testing.T) {
	var s10_epsilon float64 = 0.0009765625
	floatTest(t, 0, 10, 0b00000000000, 0.0, s10_epsilon)
	floatTest(t, 0, 10, 0b10000000000, -1.0, s10_epsilon)
	floatTest(t, 0, 10, 0b01111111111, 0.9990234375, s10_epsilon)
	floatTest(t, 0, 10, 0b00000000001, s10_epsilon, s10_epsilon)
	floatTest(t, 0, 10, 0x200, 0.5, s10_epsilon)
	floatTest(t, 0, 10, 0x440, -0.9375, s10_epsilon)
	floatTest(t, 0, 10, 0x780, -0.125, s10_epsilon)
}

func Test_Real2ToFloat_S1_14(t *testing.T) {
	var s1_14_epsilon float64 = 0.00006103516
	floatTest(t, 1, 14, 0x8000, -2, s1_14_epsilon)
	floatTest(t, 1, 14, 0xFFFF, -0.00006103516, s1_14_epsilon)
	floatTest(t, 1, 14, 0x0, 0.0, s1_14_epsilon)
	floatTest(t, 1, 14, 0x1, 0.00006103516, s1_14_epsilon)
	floatTest(t, 1, 14, 0x7FFF, 1.99993896484, s1_14_epsilon)

	floatTest(t, 1, 14, 0xe000, -0.5, s1_14_epsilon)
}

func Test_Real2ToFloat_S1_9(t *testing.T) {
	var s1_9_epsilon float64 = 0.001953125

	floatTest(t, 1, 9, 0x0, 0.0, s1_9_epsilon)
	floatTest(t, 1, 9, 0x1, 0.001953125, s1_9_epsilon)
	floatTest(t, 1, 9, 0x7FF-1, -s1_9_epsilon, s1_9_epsilon)
	floatTest(t, 1, 9, 0x7FF, -0.001953125, s1_9_epsilon)
	floatTest(t, 1, 9, 0x3FF, 1.998046875, s1_9_epsilon)

	floatTest(t, 1, 9, 0x700, -0.5, s1_9_epsilon)
	floatTest(t, 1, 9, 0x6c0, -0.625, s1_9_epsilon)
	floatTest(t, 1, 9, 0x400, -2.0, s1_9_epsilon)
	floatTest(t, 1, 9, 0x140, 0.625, s1_9_epsilon)
	floatTest(t, 1, 9, 0x200, 1.0, s1_9_epsilon)
}

func Test_Real4ToFloat_S4_6(t *testing.T) {
	var s4_6_epsilon float64 = 1.0 / float64((int(1) << 6))

	floatTest(t, 4, 6, 0x0, 0.0, s4_6_epsilon)
	floatTest(t, 4, 6, 0x7FF, -16.0, s4_6_epsilon)
	floatTest(t, 4, 6, 0x3FF, 15.999998, s4_6_epsilon)
	floatTest(t, 4, 6, 0x400, -s4_6_epsilon, s4_6_epsilon)
	floatTest(t, 4, 6, 0x1, s4_6_epsilon, s4_6_epsilon)
}
