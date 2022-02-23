package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	wav "github.com/youpy/go-wav"

	ui "github.com/gizak/termui/v3"

	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/disasm"
	"github.com/handegar/fv1emu/dsp"
	"github.com/handegar/fv1emu/reader"
	"github.com/handegar/fv1emu/settings"
)

type ChannelStatistics struct {
	Max     int32
	Min     int32
	Mean    int
	Clipped int
	Silent  bool
}

type WavStatistics struct {
	Left       ChannelStatistics
	Right      ChannelStatistics
	NumSamples int
}

func parseCommandLineParameters() {
	flag.StringVar(&settings.InFilename, "bin", settings.InFilename, "FV-1 binary file")
	flag.StringVar(&settings.InFilename, "hex", settings.InFilename, "SpinCAD/Intel HEX file")
	flag.StringVar(&settings.InputWav, "in", settings.InputWav, "Input wav-file")
	flag.StringVar(&settings.OutputWav, "out", settings.OutputWav, "Output wav-file")

	flag.Float64Var(&settings.ClockFrequency, "clock", settings.ClockFrequency, "Chrystal frequency")
	flag.Float64Var(&settings.TrailSeconds, "trail", settings.TrailSeconds, "Additional trail length (seconds)")

	flag.Float64Var(&settings.Pot0Value, "p0", settings.Pot0Value, "Potentiometer 0 value (0 .. 1.0)")
	flag.Float64Var(&settings.Pot1Value, "p1", settings.Pot1Value, "Potentiometer 1 value (0 .. 1.0)")
	flag.Float64Var(&settings.Pot2Value, "p2", settings.Pot2Value, "Potentiometer 2 value (0 .. 1.0)")

	flag.BoolVar(&settings.PrintCode, "print-code", settings.PrintCode, "Print program code")

	flag.BoolVar(&settings.Debugger, "debug", settings.Debugger, "Enable step-debugger user-interface")
	flag.BoolVar(&settings.PrintDebug, "print-debug", settings.PrintDebug, "Print additional info when debugging")

	var allPotsToMax bool = false
	flag.BoolVar(&allPotsToMax, "pmax", allPotsToMax, "Set all potentiometers to max")
	var allPotsToMin bool = false
	flag.BoolVar(&allPotsToMax, "pmin", allPotsToMax, "Set all potentiometers to min")

	flag.Parse()

	if allPotsToMax {
		fmt.Println("* Setting all potentiometers to max")
		settings.Pot0Value = 1.0
		settings.Pot1Value = 1.0
		settings.Pot2Value = 1.0
	} else if allPotsToMin {
		fmt.Println("* Setting all potentiometers to min")
		settings.Pot0Value = 0
		settings.Pot1Value = 0
		settings.Pot2Value = 0
	}
}

func updateWavStatistics(left int32, right int32, statistics *WavStatistics) {
	statistics.Left.Max = int32(math.Max(float64(statistics.Left.Max), float64(left)))
	statistics.Left.Min = int32(math.Min(float64(statistics.Left.Min), float64(left)))
	statistics.Left.Mean += int(left)
	if math.Abs(float64(left)) > 0.0 {
		statistics.Left.Silent = false
	}
	if math.Abs(float64(left)) > float64(0x7FFF) {
		statistics.Left.Clipped += 1
	}

	statistics.Right.Max = int32(math.Max(float64(statistics.Right.Max), float64(right)))
	statistics.Right.Min = int32(math.Min(float64(statistics.Right.Min), float64(right)))
	statistics.Right.Mean += int(right)
	if math.Abs(float64(right)) > 0.0 {
		statistics.Right.Silent = false
	}
	if math.Abs(float64(right)) > float64(0x7FFF) {
		statistics.Right.Clipped += 1
	}

	statistics.NumSamples += 1
}

func printWavStatistics(statistics *WavStatistics) {
	if statistics.Left.Silent {
		color.Cyan("* NOTE: Left channel is completely silent.")
	}
	if statistics.Right.Silent {
		color.Cyan("* NOTE: Right channel is completely silent.")
	}
	if statistics.Left.Clipped > 0 {
		color.Red("* WARNING: Left channel has %d clipped samples.",
			statistics.Left.Clipped)
	}
	if statistics.Right.Clipped > 0 {
		color.Red("* WARNING: Right channel has %d clipped samples.",
			statistics.Right.Clipped)
	}

	color.Yellow("- Left channel MinMax=<%d, %d>. Mean=%d",
		statistics.Left.Min, statistics.Left.Max, statistics.Left.Mean)
	color.Yellow("- Right channel MinMax=<%d, %d>. Mean=%d",
		statistics.Right.Min, statistics.Right.Max, statistics.Right.Mean)
}

func saveWavFile(wavFormat *wav.WavFormat, outSamples []wav.Sample) error {
	fmt.Printf("* Writing to '%s'\n", settings.OutputWav)
	outWAVFile, err := os.Create(settings.OutputWav)
	if err != nil {
		fmt.Printf("Error creating output file: %s\n", err)
		return err
	}
	writer := wav.NewWriter(outWAVFile, uint32(len(outSamples)), 2,
		wavFormat.SampleRate, wavFormat.BitsPerSample)
	defer outWAVFile.Close()

	err = writer.WriteSamples(outSamples)
	if err != nil {
		fmt.Printf("Error writing samples: %s\n", err)
		return err
	}

	return nil
}

func main() {
	fmt.Printf("* FV-1 emulator v%s\n", settings.Version)
	parseCommandLineParameters()

	if settings.Debugger {
		// Setting up TermUI
		if err := ui.Init(); err != nil {
			log.Fatalf("failed to initialize termui: %v", err)
		}
		defer ui.Close()
	}

	if settings.InFilename == "" {
		fmt.Println("No bin/hex file specified. Use the '-bin/-hex' parameter.")
		return
	}

	var buf []uint32
	var err error
	if strings.HasSuffix(settings.InFilename, ".bin") {
		buf, err = reader.ReadBin(settings.InFilename)
		if err != nil {
			fmt.Printf("Reading BIN file failed: %s\n", err)
			return
		}
	} else if strings.HasSuffix(settings.InFilename, ".hex") {
		buf, err = reader.ReadHex(settings.InFilename)
		if err != nil {
			fmt.Printf("Reading HEX file failed: %s\n", err)
			return
		}
	}

	var opCodes = dsp.DecodeOpCodes(buf)
	_ = opCodes

	if len(opCodes) == 0 {
		fmt.Println("No input instructions in BIN/HEX file...")
		return
	}

	if settings.PrintCode {
		disasm.PrintCodeListing(opCodes)
	}

	if settings.InputWav == "" {
		fmt.Println("No input WAV file specified. Use the '-in' parameter.")
		return
	}

	if settings.OutputWav == "" {
		fmt.Println("No output WAV file specified. Use the '-out' parameter.")
		return
	}

	inWAVFile, _ := os.Open(settings.InputWav)
	reader := wav.NewReader(inWAVFile)
	defer inWAVFile.Close()

	wavFormat, err := reader.Format()
	if err != nil {
		fmt.Printf("Error fetching WAV format: %s\n", err)
		return
	}
	isStereo := wavFormat.NumChannels == 2
	settings.SampleRate = float64(wavFormat.SampleRate)

	fmt.Printf("* Reading '%s': %d channels, %dHz, %dbit\n",
		settings.InputWav, wavFormat.NumChannels, wavFormat.SampleRate, wavFormat.BitsPerSample)
	fmt.Printf("* Chrystal frequency: %.2f Hz\n", settings.ClockFrequency)

	var statistics WavStatistics
	statistics.Left.Silent = true
	statistics.Right.Silent = true

	fmt.Println("  Processing...")

	start := time.Now()
	var state *dsp.State = dsp.NewState()
	sampleNum := 0

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
			outLeft, outRight, cont := processSample(left, right, state, opCodes, sampleNum)
			updateWavStatistics(outLeft, outRight, &statistics)
			outSamples = append(outSamples,
				wav.Sample{[2]int{int(outLeft), int(outRight)}})

			if !cont {
				return
			}

			sampleNum += 1
		}
	}

	if settings.Debugger {
		return
	}

	// Do trail-samples?
	numSamples := len(outSamples)
	numTrailSamples := int(settings.TrailSeconds * settings.SampleRate)
	for i := 0; i < numTrailSamples; i++ {
		outLeft, outRight, _ := processSample(0.0, 0.0, state, opCodes, i+numSamples)
		updateWavStatistics(outLeft, outRight, &statistics)
		outSamples = append(outSamples,
			wav.Sample{[2]int{int(outLeft), int(outRight)}})
	}

	duration := time.Since(start)
	fmt.Printf("   -> ..took %s (%d samples)\n", duration, len(outSamples))

	statistics.Left.Mean = statistics.Left.Mean / int(statistics.NumSamples)
	statistics.Right.Mean = statistics.Right.Mean / int(statistics.NumSamples)

	printWavStatistics(&statistics)
	saveWavFile(wavFormat, outSamples)
}

// Returns an Int-pair (16bits signed)
func processSample(inRight float64, inLeft float64, state *dsp.State, opCodes []base.Op, sampleNum int) (int32, int32, bool) {
	state.GetRegister(base.ADCL).SetFloat64(inLeft)
	state.GetRegister(base.ADCR).SetFloat64(inRight)
	cont := dsp.ProcessSample(opCodes, state, sampleNum)
	outLeft := int32(state.GetRegister(base.DACL).ToFloat64() * (1 << 15))
	outRight := int32(state.GetRegister(base.DACR).ToFloat64() * (1 << 15))
	return outLeft, outRight, cont
}
