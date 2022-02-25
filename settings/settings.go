package settings

var Version = "0.1"

var InputWav = "input.wav"
var OutputWav = "output.wav"
var InFilename = ""

// Do a code printout
var PrintCode = true

// Max number of operations allowed in a FV-1 program
var MaxNumberOfOps = 128

// Potentiometer values
var Pot0Value = 0.5
var Pot1Value = 0.5
var Pot2Value = 0.5

// Step debugger
var Debugger = false

// Current samplerate
var SampleRate = 44100.0

// Std. clock speed
var ClockFrequency = 32768.0

// Trail samples
var TrailSeconds = 0.0

// Print extra debug info when disassembling
var PrintDebug = false

//
// Debug stuff -- might disapear
//
var CHO_RDAL_is_NA = false
var CHO_RDAL_is_RPTR2 = false
var CHO_RDAL_is_COMPA = false
var CHO_RDAL_is_COS = false
