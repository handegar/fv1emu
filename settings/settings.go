package settings

var Version = "0.1"

var InputWav = "input.wav"
var OutputWav = "output.wav"
var InFilename = ""

// Stream result to speaker?
var Stream = false

// Do a code printout
var PrintCode = true

// Potentiometer values
var Pot0Value = 0.5
var Pot1Value = 0.5
var Pot2Value = 0.5

// Step debugger
var Debugger = false

// Activate profiler
var Profiler = false

// Samplerate of the output result
var SampleRate = 44100.0

// Internal clock speed of the "chip". Usually 32768.0 but we'll match
// the samplerate as this is more convenient.
var ClockFrequency = 44100.0

// Trail samples
var TrailSeconds = 0.0

// Print extra debug info
var PrintDebug = false

// The number of instructions the FV-1 will process each sample.
var InstructionsPerSample = 128

// Skip to sample @ startup
var SkipToSample = -1

// Only process N-samples
var StopAtSample = -1

//
var ProgramNumber = 0

// The simulator uses 32bits fixed floats but the FV-1 uses 24bits
// floats. We will therefore clamp all values to 24 bits. However one
// might want to detect when a register or DAC reaches it's limits to
// catch whatever might cause clipping. Disabling the clamping will
// then allow the values to go all the way to 32bits. Overflows will
// be highlighted in the debugger.
var Disable24BitsClamping = false

// Write the result value for a register for each sample to a CSV file
// Default filename will be 'reg-<NUM>.csv'. Ignored if value is < 0.
var WriteRegisterToCSV = -1

//
// Debug stuff -- might disapear
//
var CHO_RDAL_is_NA = false
var CHO_RDAL_is_NA_COMPC = false
var CHO_RDAL_is_RPTR2 = false
var CHO_RDAL_is_RPTR2_COMPC = false
var CHO_RDAL_is_COMPA = false
var CHO_RDAL_is_COMPC = false
var CHO_RDAL_is_COS = false

// Output filename for the CPU profiler
var ProfilerFilename = ""
