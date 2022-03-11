# fv1emu

A simple Spin FV-1 DSP emulator written in GOLang. It has a built-in
debugger for step-by-step evaluating and state inspection.


## Current state

Most of the reverb/delay related programs work as expected. The main issues currently is related to the LFO behaviour and the *CHO RDA* instruction.
All programs used when testing has been compiled with the *asfv1.py* assembler.



## Dependencies

All dependencies are listed in "go.mod"


## Compiling

    $ go mod download
    $ go build


## Usage

    $ ./fv1emu --help
    $ ./fv1emu --in INPUT.WAV --out OUTPUT.WAV --bin ALGO.BIN 


## Debugger

It is possible to step-debug a FV-1 program by adding the '-debug' paramenter.
One can then inspect the internal state and registers of the system for each operation.

![Debugger](/debugger-screenshot.png)


## TODOs

 - Calibrate the LFO with an actual FV-1 dsp.
 - Get the Ramp-LFO right.
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
    
