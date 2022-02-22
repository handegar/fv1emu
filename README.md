# fv1emu
A simple Spin FV-1 DSP emulator written in GOLang.

NOTE: This is work in progress and not usable (yet)

## Building

### Dependencies
    FIXME

### Compiling

    $ go build


## Usage

    $ ./fv1emu --help
    $ ./fv1emu --in INPUT.WAV --out OUTPUT.WAV --bin ALGO.BIN 

## Debugger

    It is possible to step-debug a FV-1 program by adding the '-debug' paramenter.
    One can then inspect the internal state and registers of the system for each operation.


## Todos

- Change to "beep" for Wav handling. This will enable streaming as well as doing wav stuff

## Links
    - A test-suite for the FV-1
      https://github.com/ndf-zz/fv1testing
