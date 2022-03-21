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


## Dependencies

All dependencies are listed in "go.mod"


## Compiling

    $ go mod download
    $ go build


## Usage

    $ ./fv1emu --help
     * FV-1 emulator v0.1
     Usage of ./fv1emu:
     -bin string
    	FV-1 binary file
     -debug
    	Enable step-debugger user-interface
     -hex string
    	SpinCAD/Intel HEX file
     -in string
    	Input wav-file (default "input.wav")
     -out string
    	Output wav-file (default "output.wav")
     -p0 float
    	Potentiometer 0 value (0 .. 1.0) (default 0.5)
     -p1 float
    	Potentiometer 1 value (0 .. 1.0) (default 0.5)
     -p2 float
    	Potentiometer 2 value (0 .. 1.0) (default 0.5)
     -pmax
    	Set all potentiometers to maximum
     -pmin
    	Set all potentiometers to minimum
     -print-code
    	Print program code (default true)
     -print-debug
    	Print additional info when debugging
     -skip-to int
    	Skip to sample number (when debugging) (default -1)
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
   the debugger
 - Test on MacOS and Windows.
 - Better streaming, preferably realtime streaming.
 - Realtime processing of an input stream.
   - Add functionality for changing the POT-values at runtime.
 - Export CSV/Excel tables with register values for each sample
   - Nice to visualize in external graphing programs. LFO shapes etc.
 - Let the user set the external clock-speed to other frequencies than
   the default.
 - Add scripted change-patterns to the realtime clock (like the EQD
   Afterneath).


## Links

 - A test-suite for the FV-1: https://github.com/ndf-zz/fv1testing

 - A Python based FV-1 assembler: https://github.com/ndf-zz/asfv1


## Notes

- Converting UTF-16 .spn files to UTF-8:
 
        $ iconv -f UTF-16LE -t UTF-8 <infile> -o <outfile>
    
