# fv1emu

A simple Spin FV-1 DSP emulator written in GOLang. It has a built-in
debugger for step-by-step evaluating and state inspection.


## License

    MIT License


## Current state

Most of the reverb/delay related programs works as expected. The main
issues currently is related to the LFO behaviour and the *CHO RDA/SOF*
instruction, especially the RAMP LFO.


All programs used when testing has been compiled using the *asfv1.py*
assembler.


The emulator has currently only been compiled and tested on
Ubuntu/Linux.

No optimizations has been done so this emulator is not yet ready
for robust realtime processing, yet.


## Dependencies

All dependencies are listed in "go.mod"


## Compiling

    $ go mod download
    $ go build


## Usage

    $ ./fv1emu -help

    -bin string
    	FV-1 binary file
    -debug
    	Enable step-debugger user-interface
    -disable-24bits-clamping
    	Disable clamping of register values to 24-bits but use the entire 32-bits range.
    -hex string
    	SpinCAD/Intel HEX file
    -in string
    	Input wav-file (default "input.wav")
    -out string
    	Output wav-file (default "output.wav")
    -p0 float
    	Potensiometer 0 value (0 .. 1.0) (default 0.5)
    -p1 float
    	Potensiometer 1 value (0 .. 1.0) (default 0.5)
    -p2 float
    	Potensiometer 2 value (0 .. 1.0) (default 0.5)
    -pmax
    	Set all potensiometers to maximum
    -pmin
    	Set all potensiometers to minimum
    -print-code
    	Print program code (default true)
    -print-debug
    	Print additional info when debugging
    -prog int
    	Which program to load for multiprogram BIN/HEX files
    -reg-to-csv int
    	Write register values to 'reg-<NUM>.csv'. One value per sample. (default -1)
    -skip-to int
    	Skip to sample number (when debugging) (default -1)
    -stop-at int
    	Stop at sample number (default -1)
    -stream
    	Stream output to sound device
    -trail float
    	Additional trail length (seconds)

    $ ./fv1emu --in INPUT.WAV --out OUTPUT.WAV --bin ALGO.BIN


## Debugger

It is possible to step-debug an FV-1 program by using the *'-debug'*
parameter. The debugger can be used to inspect the internal state and
registers of the system. The debugger is terminal-based.

![Debugger](/debugger-screenshot.png)

The debugger also has a simple Delay Memory inspector.


## TODOs

 - Calibrate the LFO with an actual FV-1 DSP.
 - Get the Ramp-LFO right.
 - Catch overflows within operations (the register.Clamp24Bit() function) and show warnings in
   the debugger.
 - Test on MacOS and Windows.
 - Better streaming, preferably realtime streaming.
 - Realtime processing of an input-stream like another app or
   a microphone.
   - Add functionality for changing the POT-values at runtime.
 - Export CSV/Excel tables with register values for each sample
   - Nice to visualize in external graphing programs. LFO shapes etc.
 - Let the user set the external clock-speed to other frequencies than
   the default.
 - Add scripted change-patterns to the realtime clock (like the EQD
   Afterneath).
 - The FV-1 has internal filters. These should be emulated.
   - The AN-0001, page 5 mentions "high-pass filtering in the DAC" in
     the Ramp LFO program.
 - Better memory-visualization
 - Keep track and visualize allocated memory chunks in addition to the
   entire memory-map.
 - Add a reset function in the debugger (or go-to)

## Links

 - A test-suite for the FV-1: https://github.com/ndf-zz/fv1testing

 - A Python based FV-1 assembler: https://github.com/ndf-zz/asfv1

 - A search on Github for SpinASM programs: https://github.com/search?q=extension%3Aspcd

 - A collection of programs: https://github.com/mstratman/fv1-programs/tree/master/docs/files

 - A VST plugin with a FV-1 simulator: https://github.com/p-kai-n/spnsim

 - A FV-1 VM running on a STM32: https://github.com/patrickdowling/fv1vm

 - A FV-1 asm to C converter: https://github.com/expertsleepersltd/spn_to_c

 - General pitch-shift tutorial: https://www.youtube.com/watch?v=fJUmmcGKZMI


## Notes

  - Converting UTF-16 .spn files to UTF-8:

        $ iconv -f UTF-16LE -t UTF-8 <infile> -o <outfile>

  - Plotting the content of a CSV register dump:
    - Execute the following script available in this repo:

        $ ./csv-plot.sh CSVFILE

  - Converting a Intel HEX file to BIN:

        $ objcopy -I ihex original.hex -O binary newfile.bin

  - On linux Ocenaudio is a nice editor with auto-reload
