package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"syscall"
	"time"

	wav "github.com/youpy/go-wav"

	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/disasm"
	"github.com/handegar/fv1emu/dsp"
	"github.com/handegar/fv1emu/reader"
	"github.com/handegar/fv1emu/settings"
)

func parseCommandLineParameters() {
	flag.StringVar(&settings.InFilename, "bin", settings.InFilename, "FV-1 binary file")
	flag.StringVar(&settings.InFilename, "hex", settings.InFilename, "SpinCAD/Intel HEX file")
	flag.StringVar(&settings.InputWav, "in", settings.InputWav, "Input wav-file")
	flag.StringVar(&settings.OutputWav, "out", settings.OutputWav, "Output wav-file")

	flag.IntVar(&settings.ProgramNo, "n", settings.ProgramNo, "Program number")
	flag.Float64Var(&settings.Pot0Value, "p0", settings.Pot0Value, "Potetiometer 0 value (0 .. 1.0)")
	flag.Float64Var(&settings.Pot1Value, "p1", settings.Pot1Value, "Potetiometer 1 value (0 .. 1.0)")
	flag.Float64Var(&settings.Pot2Value, "p2", settings.Pot2Value, "Potetiometer 2 value (0 .. 1.0)")

	flag.BoolVar(&settings.PrintStats, "print-stats", settings.PrintStats, "Print program stats")
	flag.BoolVar(&settings.PrintCode, "print-code", settings.PrintCode, "Print program code")
	flag.BoolVar(&settings.PrintDebug, "print-debug", settings.PrintDebug, "Print debug info with program code")
	flag.Parse()
}

func main() {
	fmt.Printf("* FV-1 emulator v%s\n", settings.Version)
	parseCommandLineParameters()

	if settings.InFilename == "" {
		fmt.Println("No bin/hex file specified. Use the '-bin/-hex' parameter.")
		syscall.Exit(-1)
	}

	var buf []uint32
	var err error
	if strings.HasSuffix(settings.InFilename, ".bin") {
		buf, err = reader.ReadBin(settings.InFilename)
		if err != nil {
			fmt.Printf("Reading BIN file failed: %s\n", err)
			syscall.Exit(-1)
		}
	} else if strings.HasSuffix(settings.InFilename, ".hex") {
		buf, err = reader.ReadHex(settings.InFilename)
		if err != nil {
			fmt.Printf("Reading HEX file failed: %s\n", err)
			syscall.Exit(-1)
		}
	}

	var opCodes = dsp.DecodeOpCodes(buf)
	_ = opCodes

	if settings.PrintCode {
		disasm.PrintCodeListing(opCodes)
	}

	if settings.InputWav == "" {
		fmt.Println("No input WAV file specified. Use the '-in' parameter.")
		syscall.Exit(-1)
	}

	if settings.OutputWav == "" {
		fmt.Println("No output WAV file specified. Use the '-out' parameter.")
		syscall.Exit(-1)
	}

	inWAVFile, _ := os.Open(settings.InputWav)
	reader := wav.NewReader(inWAVFile)
	defer inWAVFile.Close()

	wavFormat, err := reader.Format()
	if err != nil {
		fmt.Printf("Error fetching WAV format: %s\n", err)
		syscall.Exit(-1)
	}
	isStereo := wavFormat.NumChannels == 2

	fmt.Printf("* Reading '%s': %d channels, %dHz, %dbit\n",
		settings.InputWav, wavFormat.NumChannels, wavFormat.SampleRate, wavFormat.BitsPerSample)

	floatToIntScaler := 2 * math.Pow(2, float64(wavFormat.BitsPerSample)-1)

	nonZeroSample := false

	fmt.Println("  Processing...")
	// FIXME: Do some timing here (20220131 handegar)
	start := time.Now()
	state := dsp.NewState()
	var outSamples []wav.Sample
	for {
		samples, err := reader.ReadSamples()
		if err == io.EOF {
			break
		}

		for _, sample := range samples {
			var left float64 = reader.FloatValue(sample, 0)
			var right float64 = left
			if isStereo {
				right = reader.FloatValue(sample, 1)
			}

			state.Registers[base.ADCL] = left
			state.Registers[base.ADCR] = right

			dsp.ProcessSample(opCodes, state)

			outLeft := state.Registers[base.DACL].(float64) * floatToIntScaler
			outRight := state.Registers[base.DACR].(float64) * floatToIntScaler

			if outLeft > 2*math.SmallestNonzeroFloat64 ||
				outRight > 2*math.SmallestNonzeroFloat64 {
				nonZeroSample = true
			}

			outSamples = append(outSamples,
				wav.Sample{[2]int{int(outLeft), int(outRight)}})
		}
	}
	duration := time.Since(start)
	fmt.Printf("   -> ..took %s\n", duration)

	if !nonZeroSample {
		fmt.Println("* NOTE: All samples were zero, ie. no sound was produced")
	}

	fmt.Printf("* Writing to '%s'\n", settings.OutputWav)
	outWAVFile, err := os.Create(settings.OutputWav)
	if err != nil {
		fmt.Printf("Error creating output file: %s\n", err)
		syscall.Exit(-1)
	}
	writer := wav.NewWriter(outWAVFile, uint32(len(outSamples)), 2,
		wavFormat.SampleRate, wavFormat.BitsPerSample)
	defer outWAVFile.Close()

	err = writer.WriteSamples(outSamples)
	if err != nil {
		fmt.Printf("Error writing samples: %s\n", err)
		syscall.Exit(-1)
	}

}
