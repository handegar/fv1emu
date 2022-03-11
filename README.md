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
    $ ./fv1emu --in INPUT.WAV --out OUTPUT.WAV --bin ALGO.BIN 



## Debugger

It is possible to step-debug an FV-1 program by using the *'-debug'*
parameter. The debugger can be used to inspect the internal state and
registers of the system. The debugger is terminal-based.

![Debugger](/debugger-screenshot.png)



## TODOs

 - Calibrate the LFO with an actual FV-1 dsp.
 - Get the Ramp-LFO right.
 - Test on MacOS and Windows.
 - Better streaming, preferably realtime streaming.
 - Realtime processing of an input stream.



## Links

 - A test-suite for the FV-1
   https://github.com/ndf-zz/fv1testing

 - A Python based FV-1 assembler
   https://github.com/ndf-zz/asfv1



## Notes

 - UTF-16 spn files to UTF-8
 
        $ iconv -f UTF-16LE -t UTF-8 <infile> -o <outfile>
    
