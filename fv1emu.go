package main

import (
	"flag"
	"fmt"
	"syscall"

	"github.com/handegar/fv1emu/dsp"
	"github.com/handegar/fv1emu/reader"
	"github.com/handegar/fv1emu/settings"
	"github.com/handegar/fv1emu/utils"
)

func parseCommandLineParameters() {
	flag.StringVar(&settings.HexFilename, "hex", settings.HexFilename, "FV-1 binary file")
	flag.StringVar(&settings.InputWav, "in", settings.InputWav, "Input wav-file")
	flag.StringVar(&settings.OutputWav, "out", settings.OutputWav, "Output wav-file")
	flag.IntVar(&settings.ProgramNo, "n", settings.ProgramNo, "Program number")
	flag.BoolVar(&settings.PrintStats, "print-stats", settings.PrintStats, "Print program stats")
	flag.BoolVar(&settings.PrintCode, "print-code", settings.PrintCode, "Print program code")
	flag.Parse()
}

func main() {
	fmt.Printf("* FV-1 emulator v%s\n", settings.Version)
	parseCommandLineParameters()

	if settings.HexFilename == "" {
		fmt.Println("No HEX/BIN file specified. Use the '-hex' parameter.")
		syscall.Exit(-1)
	}

	buf, err := reader.ReadHEX(settings.HexFilename)
	if err != nil {
		fmt.Printf("Reading HEX file failed: %s\n", err)
		syscall.Exit(-1)
	}

	var opCodes = dsp.ParseBuffer(buf)
	_ = opCodes

	if settings.PrintCode {
		utils.PrintCodeListing(opCodes)
	}

}
