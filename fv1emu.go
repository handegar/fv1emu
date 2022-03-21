package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/fatih/color"
	ui "github.com/gizak/termui/v3"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"

	"github.com/handegar/fv1emu/base"
	"github.com/handegar/fv1emu/debugger"
	"github.com/handegar/fv1emu/disasm"
	"github.com/handegar/fv1emu/dsp"
	"github.com/handegar/fv1emu/reader"
	"github.com/handegar/fv1emu/settings"
	"github.com/handegar/fv1emu/writer"
)

type ChannelStatistics struct {
	Max             float64
	Min             float64
	Mean            float64
	Clipped         int
	FirstClipSample int
	Silent          bool
}

type WavStatistics struct {
	Left       ChannelStatistics
	Right      ChannelStatistics
	NumSamples int
}

func parseCommandLineParameters() {
	flag.StringVar(&settings.InFilename, "bin", settings.InFilename,
		"FV-1 binary file")
	flag.StringVar(&settings.InFilename, "hex", settings.InFilename,
		"SpinCAD/Intel HEX file")
	flag.StringVar(&settings.InputWav, "in", settings.InputWav,
		"Input wav-file")
	flag.StringVar(&settings.OutputWav, "out", settings.OutputWav,
		"Output wav-file")
	flag.BoolVar(&settings.Stream, "stream", settings.Stream,
		"Stream output to sound device")

	// FIXME: Not fully implemented yet (20220305 handegar)
	/*
		flag.Float64Var(&settings.ClockFrequency, "clock", settings.ClockFrequency,
			"Chrystal frequency")
	*/

	flag.Float64Var(&settings.TrailSeconds, "trail", settings.TrailSeconds,
		"Additional trail length (seconds)")

	flag.Float64Var(&settings.Pot0Value, "p0", settings.Pot0Value,
		"Potentiometer 0 value (0 .. 1.0)")
	flag.Float64Var(&settings.Pot1Value, "p1", settings.Pot1Value,
		"Potentiometer 1 value (0 .. 1.0)")
	flag.Float64Var(&settings.Pot2Value, "p2", settings.Pot2Value,
		"Potentiometer 2 value (0 .. 1.0)")

	flag.BoolVar(&settings.Debugger, "debug", settings.Debugger,
		"Enable step-debugger user-interface")
	flag.IntVar(&settings.SkipToSample, "skip-to", settings.SkipToSample,
		"Skip to sample number (when debugging)")

	flag.BoolVar(&settings.PrintCode, "print-code", settings.PrintCode,
		"Print program code")

	flag.BoolVar(&settings.PrintDebug, "print-debug", settings.PrintDebug,
		"Print additional info when debugging")

	var allPotsToMax bool = false
	flag.BoolVar(&allPotsToMax, "pmax", allPotsToMax,
		"Set all potentiometers to maximum")
	var allPotsToMin bool = false
	flag.BoolVar(&allPotsToMin, "pmin", allPotsToMin,
		"Set all potentiometers to minimum")

	// Debug stuff
	flag.BoolVar(&settings.CHO_RDAL_is_NA, "cho-rdal-is-NA", settings.CHO_RDAL_is_NA,
		"DEBUG: The 'CHO RDAL' op will output the NA crossfade envelope for RMP0")
	flag.BoolVar(&settings.CHO_RDAL_is_RPTR2, "cho-rdal-is-RPTR2", settings.CHO_RDAL_is_RPTR2,
		"DEBUG: The 'CHO RDAL' op will output the RPTR2 envelope for RMP0")
	flag.BoolVar(&settings.CHO_RDAL_is_COMPA, "cho-rdal-is-COMPA", settings.CHO_RDAL_is_COMPA,
		"DEBUG: The 'CHO RDAL' op will output the COMPA envelope for SIN0/RMP0")
	flag.BoolVar(&settings.CHO_RDAL_is_COS, "cho-rdal-is-COS", settings.CHO_RDAL_is_COS,
		"DEBUG: The 'CHO RDAL' op will output the COS envelope for SIN0")

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

func updateWavStatistics(sampleNum int, left float64, right float64, statistics *WavStatistics) {
	statistics.Left.Max = math.Max(statistics.Left.Max, left)
	statistics.Left.Min = math.Min(statistics.Left.Min, left)
	statistics.Left.Mean += left
	if math.Abs(left) > 0.0 {
		statistics.Left.Silent = false
	}
	if math.Abs(left) > 1.0 {
		statistics.Left.Clipped += 1
		if statistics.Left.FirstClipSample == 0 {
			statistics.Left.FirstClipSample = sampleNum
		}
	}

	statistics.Right.Max = math.Max(statistics.Right.Max, right)
	statistics.Right.Min = math.Min(statistics.Right.Min, right)
	statistics.Right.Mean += right
	if math.Abs(right) > 0.0 {
		statistics.Right.Silent = false
	}
	if math.Abs(right) > 1.0 {
		statistics.Right.Clipped += 1
		if statistics.Right.FirstClipSample == 0 {
			statistics.Right.FirstClipSample = sampleNum
		}
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
		color.Red("* WARNING: Left channel had %d clipped samples (First clip @ sample %d).",
			statistics.Left.Clipped, statistics.Left.FirstClipSample)
	}
	if statistics.Right.Clipped > 0 {
		color.Red("* WARNING: Right channel had %d clipped samples (First clip @ sample %d).",
			statistics.Right.Clipped, statistics.Right.FirstClipSample)
	}

	color.Yellow("- Left channel MinMax=<%f, %f>. Mean=%f",
		statistics.Left.Min, statistics.Left.Max, statistics.Left.Mean)
	color.Yellow("- Right channel MinMax=<%f, %f>. Mean=%f",
		statistics.Right.Min, statistics.Right.Max, statistics.Right.Mean)
}

func main() {
	fmt.Printf("* FV-1 emulator v%s\n", settings.Version)
	parseCommandLineParameters()

	if settings.InFilename == "" {
		fmt.Println("No bin/hex file specified. Use the '-bin/-hex' parameter.")
		return
	}

	if settings.InputWav == "" {
		fmt.Println("No input WAV file specified. Use the '-in' parameter.")
		return
	}

	if settings.OutputWav == "" {
		fmt.Println("No output WAV file specified. Use the '-out' parameter.")
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

	printPotentiometersInUse(opCodes)

	f, stream, wavFormat, err := reader.ReadWAV(settings.InputWav)
	defer f.Close()

	isStereo := wavFormat.NumChannels == 2
	settings.SampleRate = float64(wavFormat.SampleRate)

	fmt.Printf("* Reading '%s': %d channels, %dHz, %dbit\n",
		settings.InputWav, wavFormat.NumChannels, wavFormat.SampleRate, wavFormat.Precision)
	fmt.Printf("* Chrystal frequency: %.2f Hz\n", settings.ClockFrequency)

	var statistics WavStatistics
	statistics.Left.Silent = true
	statistics.Right.Silent = true

	fmt.Printf("* Processing...\n")

	start := time.Now()
	var state *dsp.State = dsp.NewState()
	sampleNum := 0

	if settings.Debugger {
		// Setting up TermUI
		if err := ui.Init(); err != nil {
			log.Fatalf("failed to initialize termui: %v", err)
		}
		debugger.Reset()
		if settings.SkipToSample > 0 {
			dsp.SkipNumSamples(settings.SkipToSample)
		}
	}

	//
	// Main processing loop
	//

	var outSamples [][2]float64
	for {
		var samples [][2]float64 = make([][2]float64, 1024)
		_, err := stream.Stream(samples)
		if !err {
			break
		}

		letsContinue := true
		for _, sample := range samples {
			var left float64 = sample[0]
			var right float64 = sample[1]
			if isStereo {
				right = sample[0]
			}
			outLeft, outRight, cont := processSample(left, right, state, opCodes, sampleNum)
			updateWavStatistics(sampleNum, outLeft, outRight, &statistics)
			outSamples = append(outSamples, [2]float64{outLeft, outRight})
			sampleNum += 1

			if !cont {
				letsContinue = false
				break
			}
		}

		if !letsContinue {
			break
		}
	}

	if settings.Debugger {
		ui.Close()
		color.Yellow("* No more samples to process.")
	} else {
		// Do trail-samples?
		numSamples := len(outSamples)

		if settings.TrailSeconds > 0.0 {
			numTrailSamples := int(settings.TrailSeconds * settings.SampleRate)
			fmt.Printf("* Adding a %.2f second(s) trail (%d samples)\n",
				settings.TrailSeconds, numTrailSamples)
			for i := 0; i < numTrailSamples; i++ {
				outLeft, outRight, ok := processSample(0.0, 0.0, state, opCodes, numSamples+i)
				updateWavStatistics(numSamples+i, 0.0, 0.0, &statistics)

				if !ok {
					break
				}

				outSamples = append(outSamples, [2]float64{outLeft, outRight})
			}
		}
		duration := time.Since(start)
		fmt.Printf("   -> ..took %s (%d samples)\n", duration, len(outSamples))
	}

	statistics.Left.Mean = statistics.Left.Mean / float64(statistics.NumSamples)
	statistics.Right.Mean = statistics.Right.Mean / float64(statistics.NumSamples)

	printWavStatistics(&statistics)

	if settings.PrintDebug {
		fmt.Printf("DEBUG: Sin0-range used: <%f, %f>\n",
			state.DebugFlags.Sin0Min, state.DebugFlags.Sin0Max)
		fmt.Printf("DEBUG: Sin1-range used: <%f, %f>\n",
			state.DebugFlags.Sin1Min, state.DebugFlags.Sin1Max)
		fmt.Printf("DEBUG: Ramp0-range used: <%f, %f>\n",
			state.DebugFlags.Ramp0Min, state.DebugFlags.Ramp0Max)
		fmt.Printf("DEBUG: Ramp1-range used: <%f, %f>\n",
			state.DebugFlags.Ramp1Min, state.DebugFlags.Ramp1Max)
		fmt.Printf("DEBUG: XFade-range used: <%f, %f>\n",
			state.DebugFlags.XFadeMin, state.DebugFlags.XFadeMax)
	}

	if settings.Stream {
		color.Cyan("* Forwarding buffer to sound-device (CTRL-C to quit)\n")
		s := new(writer.WriteStreamer)
		s.Data = outSamples
		err = speaker.Init(wavFormat.SampleRate, len(outSamples))
		if err != nil {
			log.Fatalf("ERROR: Failed to initialize 'beep.Speaker': %v", err)
		}

		// FIXME: The "done" signal does not work properly
		// yet. Investigate. (20220310 handegar)
		done := make(chan bool, 1)
		speaker.Play(beep.Seq(s, beep.Callback(func() {
			done <- true
		})))
		<-done

	} else {
		if !settings.Debugger {
			writer.SaveAsWAV(settings.OutputWav, wavFormat, outSamples)
		}
	}
}

func DebugPreFn(opCodes []base.Op, state *dsp.State, sampleNum int) int {
	debugger.RegisterState(state)
	debugger.UpdateScreen(opCodes, state, sampleNum)
	return dsp.Ok
}

func DebugPostFn(opCodes []base.Op, state *dsp.State, sampleNum int) int {
	e := debugger.WaitForInput(state)
	switch e {
	case "quit":
		return dsp.Quit
	case "next op":
		break
	case "previous op":
		// Rewind until we find the closes valid state. Might
		// be longer than -1 due to SKPs.
		var prevState *dsp.State = nil
		for x := state.IP - 1; x >= 0; x-- {
			ok := false
			prevState, ok = debugger.GetRegisteredState(x)
			if ok {
				break
			}
		}

		if prevState != nil {
			state.Copy(prevState)
			return dsp.NextInstruction
		} else {
			fmt.Println("FATAL: No previous state could be found!")
			return dsp.Fatal
		}
	case "next sample":
		dsp.SkipNumSamples(1)
		break
	case "next 100 samples":
		dsp.SkipNumSamples(100)
		break
	case "next 1000 samples":
		dsp.SkipNumSamples(1000)
		break
	case "next 10000 samples":
		dsp.SkipNumSamples(10000)
		break
	case "next 100000 samples":
		dsp.SkipNumSamples(100000)
		break
	}
	return dsp.Ok
}

func NoDebugFn(opCodes []base.Op, state *dsp.State, sampleNum int) int {
	// Do nothing
	return dsp.Ok
}

// Returns an Int-pair (16bits signed)
func processSample(inRight float64, inLeft float64, state *dsp.State, opCodes []base.Op, sampleNum int) (float64, float64, bool) {
	state.GetRegister(base.ADCL).SetFloat64(inLeft)
	state.GetRegister(base.ADCR).SetFloat64(inRight)
	cont := true
	if settings.Debugger && dsp.GetSkipNumSamples() == 0 {
		cont = dsp.ProcessSample(opCodes, state, sampleNum, DebugPreFn, DebugPostFn)
	} else {
		cont = dsp.ProcessSample(opCodes, state, sampleNum, NoDebugFn, NoDebugFn)
	}
	outLeft := state.GetRegister(base.DACL).ToFloat64()
	outRight := state.GetRegister(base.DACR).ToFloat64()
	return outLeft, outRight, cont
}

func printPotentiometersInUse(opCodes []base.Op) {
	pot0used, pot1used, pot2used := dsp.UsesPotentiometers(opCodes)

	var pots []string
	if pot0used {
		pots = append(pots, "POT0")
	}
	if pot1used {
		pots = append(pots, "POT1")
	}
	if pot2used {
		pots = append(pots, "POT2")
	}

	if len(pots) == 0 {
		color.Cyan("* No potentiometers in use.\n")
	} else {
		color.Cyan("* Potentiometers in use: %s\n", strings.Join(pots, ", "))

	}

}
