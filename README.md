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
 - Adjust the POT values in the debugger
 - See ramp NA, COMPA and COMPC values in debugger

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

  - List of pitch coefficients:
    From https://www.diystompboxes.com/smfforum/index.php?topic=131801.0

    EQU DOWN24 -0.375
	EQU DOWN23 -0.3675671132
	EQU DOWN22 -0.359692244
	EQU DOWN21 -0.3513491106
	EQU DOWN20 -0.3425098688
	EQU DOWN19 -0.3331450182
	EQU DOWN18 -0.3232233047
	EQU DOWN17 -0.3127116154
	EQU DOWN16 -0.3015748685
	EQU DOWN15 -0.2897758962
	EQU DOWN14 -0.2772753205
	EQU DOWN13 -0.2640314218
	EQU DOWN12 -0.25
	EQU DOWN11 -0.235134226
	EQU DOWN10 -0.219384488
	EQU DOWN9 -0.202698221
	EQU DOWN8 -0.185019738
	EQU DOWN7 -0.166290036
	EQU DOWN6 -0.146446609
	EQU DOWN5 -0.125423231
	EQU DOWN4 -0.103149737
	EQU DOWN3 -0.079551792
	EQU DOWN2 -0.054550641
	EQU DOWN1 -0.028062844
	EQU ROOT  0
	EQU UP1	0.029731547
	EQU UP2	0.061231024
	EQU UP3	0.094603558
	EQU UP4	0.129960525
	EQU UP5	0.167419927
	EQU UP6	0.207106781
	EQU UP7	0.249153538
	EQU UP8	0.293700526
	EQU UP9	0.340896415
	EQU UP10 0.390898718
	EQU UP11 0.443874313
	EQU UP12 0.5
	EQU UP13 0.5594630944
	EQU UP14 0.6224620483
	EQU UP15 0.689207115
	EQU UP16 0.7599210499
	EQU UP17 0.8348398542
	EQU UP18 0.9142135624
	EQU UP19 0.9983070769
