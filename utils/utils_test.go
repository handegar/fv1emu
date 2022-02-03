package utils

import (
	"math"
	"testing"
)

func floatTest1(t *testing.T, bits int, in uint32, expected float64, epsilon float64) {
	f := Real1ToFloat(bits, in)
	if math.Abs(float64(f)-expected) > epsilon {
		t.Fatalf("S.%d of '0x%x'/0b%b != %f (got %f)\n",
			bits-2, in, in, expected, f)
	}
}

func floatTest2(t *testing.T, bits int, in uint32, expected float64, epsilon float64) {
	f := Real2ToFloat(bits, in)
	if math.Abs(float64(f)-expected) > epsilon {
		//PrintUInt32AsBinary(in)
		//PrintUInt32AsQFormat(1, bits-2, in)
		t.Fatalf("S1.%d of '0x%x'/0b%b != %f (got %f)\n",
			bits-2, in, in, expected, f)
	}
}

func floatTest4(t *testing.T, bits int, in uint32, expected float64, epsilon float64) {
	f := Real4ToFloat(bits, in)
	if math.Abs(float64(f)-expected) > epsilon {
		t.Fatalf("S4.%d of '0x%x'/0b%b != %f (got %f)\n",
			bits-4-1, in, in, expected, f)
	}
}

func Test_Real1ToFloat_S_10(t *testing.T) {
	var s10_epsilon float64 = 0.0009765625
	floatTest1(t, 11, 0b00000000000, 0.0, s10_epsilon)
	floatTest1(t, 11, 0b10000000000, -1.0, s10_epsilon)
	floatTest1(t, 11, 0b01111111111, 0.9990234375, s10_epsilon)
	floatTest1(t, 11, 0b00000000001, s10_epsilon, s10_epsilon)
	floatTest1(t, 11, 0x200, 0.5, s10_epsilon)
	floatTest1(t, 11, 0x440, -0.9375, s10_epsilon)
}

func Test_Real2ToFloat_S1_14(t *testing.T) {
	var s1_14_epsilon float64 = 0.00006103516
	floatTest2(t, 16, 0x8000, -2, s1_14_epsilon)
	floatTest2(t, 16, 0xFFFF, -0.00006103516, s1_14_epsilon)
	floatTest2(t, 16, 0x0, 0.0, s1_14_epsilon)
	floatTest2(t, 16, 0x1, 0.00006103516, s1_14_epsilon)
	floatTest2(t, 16, 0x7FFF, 1.99993896484, s1_14_epsilon)
}

func Test_Real2ToFloat_S1_9(t *testing.T) {
	var s1_9_epsilon float64 = 0.001953125

	// disfy-1 returns -0.625, but that has to be wrong as the
	// integer bit is set.
	//floatTest2(t, 11, 0x6c0, -1.376344, s1_9_epsilon) // 110'1100'0000

	floatTest2(t, 11, 0x700, -0.5, s1_9_epsilon)   //
	floatTest2(t, 11, 0x6c0, -0.625, s1_9_epsilon) // 110'1100'0000
	floatTest2(t, 11, 0x400, -2.0, s1_9_epsilon)   // 100'0000'0000

	floatTest2(t, 11, 0x0, 0.0, s1_9_epsilon)               //             0
	floatTest2(t, 11, 0x1, 0.001953125, s1_9_epsilon)       //             1
	floatTest2(t, 11, 0x7FF-1, -s1_9_epsilon, s1_9_epsilon) // 111'1111'1110
	floatTest2(t, 11, 0x7FF, -0.001953125, s1_9_epsilon)    // 111'1111'1111
	floatTest2(t, 11, 0x3FF, 1.998046875, s1_9_epsilon)     //  11'1111'1111

	floatTest2(t, 11, 0x140, 0.625, s1_9_epsilon) // 110'1100'0000
}

func Test_Real4ToFloat_S4_6(t *testing.T) {
	var s4_6_epsilon float64 = 1.0 / float64((int(1) << 6))
	floatTest4(t, 11, 0x0, 0.0, s4_6_epsilon)
	floatTest4(t, 11, 0x7FF, -0.015625, s4_6_epsilon)
	floatTest4(t, 11, 0x3FF, 15.999998, s4_6_epsilon)
	floatTest4(t, 11, 0x400, -16.0, s4_6_epsilon)
	floatTest4(t, 11, 0x1, s4_6_epsilon, s4_6_epsilon)
}

func Test_StringToQFormat(t *testing.T) {

	//str := "-0.625"
	/*
		str := "-0.5"
		fmt.Printf("INPUT: '%s'\n", str)
		v := StringToQFormat(str, 1, 9)
		PrintUInt32AsQFormat(1, 9, v)
		f := Real2ToFloat(11, v)
		fmt.Printf("OUTPUT: %f (alt: %f)\n", f, QFormatToFloat64(v, 1, 9))
	*/
	/*
		str = "0.625"
		fmt.Printf("INPUT: '%s'\n", str)
		v = StringToQFormat(str, 1, 9)
		PrintUInt32AsQFormat(1, 9, v)
		f = Real2ToFloat(11, v)
		fmt.Printf("OUTPUT: %f (alt: %f)\n", f, QFormatToFloat64(v, 1, 9))
	*/
}
