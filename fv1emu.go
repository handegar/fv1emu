package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime/pprof"
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

func parseCommandLineParameters() bool {
	flag.StringVar(&settings.InFilename, "bin",
		settings.InFilename, "FV-1 binary file")
	flag.StringVar(&settings.InFilename, "hex",
		settings.InFilename, "SpinCAD/Intel HEX file (alias for \"-bin\")")
	flag.StringVar(&settings.InputWav, "in",
		settings.InputWav, "Input wav-file")
	flag.StringVar(&settings.OutputWav, "out",
		settings.OutputWav, "Output wav-file")
	flag.BoolVar(&settings.Stream, "stream",
		settings.Stream, "Stream output to sound device")

	//
	// Potentimeters settings
	//
	var allPotsToMax bool = false
	flag.BoolVar(&allPotsToMax, "pmax",
		allPotsToMax, "Set all potentiometers to maximum")
	var allPotsToMin bool = false
	flag.BoolVar(&allPotsToMin, "pmin",
		allPotsToMin, "Set all potentiometers to minimum")

	flag.Float64Var(&settings.Pot0Value, "p0", settings.Pot0Value,
		"Potentiometer 0 value (0 .. 1.0)")
	flag.Float64Var(&settings.Pot1Value, "p1", settings.Pot1Value,
		"Potentiometer 1 value (0 .. 1.0)")
	flag.Float64Var(&settings.Pot2Value, "p2", settings.Pot2Value,
		"Potentiometer 2 value (0 .. 1.0)")

	// FIXME: Not fully implemented yet (20220305 handegar)
	/*
		flag.Float64Var(&settings.ClockFrequency, "clock", settings.ClockFrequency,
			"Chrystal frequency")
	*/

	flag.Float64Var(&settings.TrailSeconds, "trail", settings.TrailSeconds,
		"Additional trail length (seconds)")

	flag.BoolVar(&settings.Debugger, "debug",
		settings.Debugger,
		"Enable step-debugger user-interface")

	flag.IntVar(&settings.ProgramNumber, "prog",
		settings.ProgramNumber,
		"Which program to load for multiprogram BIN/HEX files")

	flag.IntVar(&settings.SkipToSample, "skip-to",
		settings.SkipToSample,
		"Skip to sample number (when debugging)")

	flag.IntVar(&settings.StopAtSample, "stop-at",
		settings.StopAtSample,
		"Stop at sample number")

	flag.IntVar(&settings.WriteRegisterToCSV, "reg-to-csv",
		settings.WriteRegisterToCSV,
		"Write register values to 'reg-<NUM>.csv'. One value per sample.")

	flag.BoolVar(&settings.Disable24BitsClamping, "disable-24bits-clamping",
		settings.Disable24BitsClamping,
		"Disable clamping of register values to 24-bits but use the entire 32-bits range.")

	flag.BoolVar(&settings.PrintCode, "print-code",
		settings.PrintCode, "Print program code")

	flag.BoolVar(&settings.PrintDebug, "print-debug",
		settings.PrintDebug, "Print additional info when debugging")

	//
	// Debug stuff. Mostly relevant when bugfixing fv1emu itself.
	//
	flag.BoolVar(&settings.CHO_RDAL_is_NA, "cho-rdal-is-NA", settings.CHO_RDAL_is_NA,
		"DEBUG: The 'CHO RDAL' op will output the NA crossfade envelope for RMP0")
	flag.BoolVar(&settings.CHO_RDAL_is_NA_COMPC, "cho-rdal-is-NA-COMPC", settings.CHO_RDAL_is_NA_COMPC,
		"DEBUG: The 'CHO RDAL' op will output the NA|COMPC crossfade envelope for RMP0")
	flag.BoolVar(&settings.CHO_RDAL_is_RPTR2, "cho-rdal-is-RPTR2", settings.CHO_RDAL_is_RPTR2,
		"DEBUG: The 'CHO RDAL' op will output the RPTR2 envelope for RMP0")
	flag.BoolVar(&settings.CHO_RDAL_is_RPTR2_COMPC, "cho-rdal-is-RPTR2-COMPC", settings.CHO_RDAL_is_RPTR2_COMPC,
		"DEBUG: The 'CHO RDAL' op will output the RPTR2|COMPC envelope for RMP0")
	flag.BoolVar(&settings.CHO_RDAL_is_COMPA, "cho-rdal-is-COMPA", settings.CHO_RDAL_is_COMPA,
		"DEBUG: The 'CHO RDAL' op will output the COMPA envelope for SIN0/RMP0")
	flag.BoolVar(&settings.CHO_RDAL_is_COMPC, "cho-rdal-is-COMPC", settings.CHO_RDAL_is_COMPC,
		"DEBUG: The 'CHO RDAL' op will output the COMPC envelope for SIN0/RMP0")
	flag.BoolVar(&settings.CHO_RDAL_is_COS, "cho-rdal-is-COS", settings.CHO_RDAL_is_COS,
		"DEBUG: The 'CHO RDAL' op will output the COS envelope for SIN0")
	flag.StringVar(&settings.ProfilerFilename, "profile", settings.ProfilerFilename,
		"DEBUG: Activate the GOLang CPU profiler by specifying the output file")

	flag.Parse()

	if flag.NFlag() == 0 {
		fmt.Printf("  Type \"./%s -help\" for more info.\n", filepath.Base(os.Args[0]))
		return false
	}

	if settings.InFilename == "" {
		fmt.Println("  No bin/hex file specified. Use the '-bin/-hex' parameter.")
		return false
	}

	if settings.InputWav == "" {
		fmt.Println("  No input WAV file specified. Use the '-in' parameter.")
		return false
	}

	if settings.OutputWav == "" {
		fmt.Println("  No output WAV file specified. Use the '-out' parameter.")
		return false
	}

	if settings.ProgramNumber < 0 || settings.ProgramNumber > 7 {
		fmt.Println("  Program number must be between 0 and 7.")
		return false
	}

	if allPotsToMax {
		settings.Pot0Value = 1.0
		settings.Pot1Value = 1.0
		settings.Pot2Value = 1.0
	} else if allPotsToMin {
		settings.Pot0Value = 0
		settings.Pot1Value = 0
		settings.Pot2Value = 0
	}

	return true
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
	if !parseCommandLineParameters() {
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

	if len(buf) <= settings.ProgramNumber*settings.InstructionsPerSample {
		fmt.Printf("Number of program(s) in BIN/HEX is only %d.\n",
			len(buf)/settings.InstructionsPerSample)
		return
	}

	var opCodes = dsp.DecodeOpCodes(buf[settings.ProgramNumber*settings.InstructionsPerSample:])
	_ = opCodes

	if len(opCodes) == 0 {
		fmt.Println("No instructions in the BIN/HEX file...")
		return
	}

	if settings.PrintCode {
		disasm.PrintCodeListing(opCodes)
	}

	printPotentiometersInUse(opCodes)
	printDACsAndADCsInUse(opCodes)

	var regCSVWriter *csv.Writer = nil
	if settings.WriteRegisterToCSV >= 0 && !settings.Debugger {
		filename := fmt.Sprintf("./reg-%d.csv", settings.WriteRegisterToCSV)
		csvfile, err := os.Create(filename)
		if err != nil {
			fmt.Printf("Could not create '%s'", filename)
			return
		}
		defer csvfile.Close()

		color.Yellow("* All valued for REG%d will be written to '%s'\n",
			settings.WriteRegisterToCSV, filename)

		regCSVWriter = csv.NewWriter(csvfile)
		regCSVWriter.Write([]string{fmt.Sprintf(";; fv1emu v%s |  Register%d dump",
			settings.Version,
			settings.WriteRegisterToCSV)})
	}

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

	if settings.ProfilerFilename != "" {
		f, err := os.Create(settings.ProfilerFilename)
		if err != nil {
			fmt.Printf("* ERROR: Could not create profile file: %s\n", err)
			return
		}
		fmt.Printf("* Writing profile info to '%s'\n", settings.ProfilerFilename)
		fmt.Printf("  - To analyze result: \"$ go tool pprof %s\"\n", settings.ProfilerFilename)
		err = pprof.StartCPUProfile(f)
		if err != nil {
			fmt.Printf("* ERROR: Could not start profiler: %s\n", err)
			return
		}
		defer f.Close()
		defer pprof.StopCPUProfile()
	}

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

			if regCSVWriter != nil {
				s := fmt.Sprintf("%f",
					state.GetRegister(base.REG0+settings.WriteRegisterToCSV).ToFloat64())
				regCSVWriter.Write([]string{s})
			}

			if !cont || (settings.StopAtSample > 0 && sampleNum >= settings.StopAtSample) {
				letsContinue = false
				break
			}
		}

		if !letsContinue {
			break
		}
	}

	/*
		if settings.Profiler {
			runtime.GC() // get up-to-date statistics
			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Fatal("Could not write memory profile: ", err)
			}
		}
	*/
	if settings.Debugger {
		ui.Close()
		color.Yellow("* No more samples to process.")
	} else {
		// Do trail-samples?
		numSamples := len(outSamples)

		if settings.TrailSeconds > 0.0 &&
			(settings.StopAtSample > 0 && sampleNum < settings.StopAtSample) {
			numTrailSamples := int(settings.TrailSeconds * settings.SampleRate)
			fmt.Printf("* Adding a %.2f second(s) trail (%d samples)\n",
				settings.TrailSeconds, numTrailSamples)
			for i := 0; i < numTrailSamples; i++ {
				outLeft, outRight, ok := processSample(0.0, 0.0, state, opCodes, numSamples+i)
				updateWavStatistics(numSamples+i, 0.0, 0.0, &statistics)

				if regCSVWriter != nil {
					s := fmt.Sprintf("%f",
						state.GetRegister(base.REG0+settings.WriteRegisterToCSV).ToFloat64())
					regCSVWriter.Write([]string{s})
				}

				if !ok {
					break
				}

				outSamples = append(outSamples, [2]float64{outLeft, outRight})
			}
		}
		duration := time.Since(start).Seconds()
		fmt.Printf("   -> ..took %fs to process %d samples (%.2f%% of realtime)\n",
			duration, len(outSamples),
			(float64(duration) / (float64(len(outSamples)) / settings.SampleRate) * 100.0))
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
	pot0used, pot1used, pot2used := dsp.PotentiometersInUse(opCodes)

	var pots []string
	if pot0used {
		pots = append(pots, fmt.Sprintf("POT0=%.3f", settings.Pot0Value))
	}
	if pot1used {
		pots = append(pots, fmt.Sprintf("POT1=%.3f", settings.Pot1Value))
	}
	if pot2used {
		pots = append(pots, fmt.Sprintf("POT2=%.3f", settings.Pot2Value))
	}

	if len(pots) == 0 {
		color.Cyan("* No potentiometers in use.\n")
	} else {
		color.Cyan("* Potentiometers in use: %s\n", strings.Join(pots, ", "))

	}
}

func printDACsAndADCsInUse(opCodes []base.Op) {
	dacr, dacl := dsp.DACsInUse(opCodes)
	adcr, adcl := dsp.ADCsInUse(opCodes)

	var lst []string
	if dacl {
		lst = append(lst, "DACL")
	}
	if dacr {
		lst = append(lst, "DACR")
	}
	if len(lst) == 0 {
		color.Red("* No DACs in use. The program will generate no sound.")
	} else {
		color.Cyan("* DACs in use: %s\n", strings.Join(lst, ", "))
	}

	lst = []string{}
	if adcl {
		lst = append(lst, "ADCL")
	}
	if adcr {
		lst = append(lst, "ADCR")
	}
	if len(lst) == 0 {
		color.Yellow("* No ADCs in use. The program takes no input.")
	} else {
		color.Cyan("* ADCs in use: %s\n", strings.Join(lst, ", "))
	}
}
