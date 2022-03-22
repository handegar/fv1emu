package settings

var Version = "0.1"

var InputWav = "input.wav"
var OutputWav = "output.wav"
var InFilename = ""

// Stream result to speaker?
var Stream = false

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

// Print extra debug info
var PrintDebug = false

// The number of instructions the FV-1 will process each sample.
var InstructionsPerSample = 128

// Skip to sample @ startup
var SkipToSample = -1

// The simulator uses 32bits fixed floats but the FV-1 uses 24bits
// floats. We will therefore clamp all values to 24 bits. However one
// might want to detect when a register or DAC reaches it's limits to
// catch whatever might cause clipping. Disabling the clamping will
// then allow the values to go all the way to 32bits. Overflows will
// be highlighted in the debugger.
var Disable24BitsClamping = false

//
// Debug stuff -- might disapear
//
var CHO_RDAL_is_NA = false
var CHO_RDAL_is_RPTR2 = false
var CHO_RDAL_is_COMPA = false
var CHO_RDAL_is_COS = false
