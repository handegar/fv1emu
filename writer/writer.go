package writer

import (
	"fmt"
	"os"

	"github.com/faiface/beep"
	"github.com/faiface/beep/wav"

	"github.com/handegar/fv1emu/settings"
	"github.com/handegar/fv1emu/utils"
)

type WriteStreamer struct {
	Data           [][2]float64
	SamplesWritten int
}

func (ws *WriteStreamer) Stream(samples [][2]float64) (n int, ok bool) {

	for i := 0; i < len(samples); i++ {

		if ws.SamplesWritten+i >= len(ws.Data) {
			//ws.SamplesWritten += i
			return i, false
		}

		utils.Assert(i < len(samples), "Index out of bounds")
		utils.Assert(len(samples[i]) == 2, "Index out of bounds")
		utils.Assert(len(ws.Data[ws.SamplesWritten+i]) == 2, "Index out of bounds")

		samples[i][0] = ws.Data[ws.SamplesWritten+i][0]
		samples[i][1] = ws.Data[ws.SamplesWritten+i][1]
	}

	ws.SamplesWritten += len(samples)
	return len(samples), ws.SamplesWritten < len(ws.Data)
}

func (ws *WriteStreamer) Err() error {
	return nil
}

func SaveAsWAV(filename string, wavFormat beep.Format, samples [][2]float64) error {
	fmt.Printf("* Writing to '%s' (%d samples, %d channels)\n",
		settings.OutputWav, len(samples), len(samples[0]))
	outWAVFile, err := os.Create(settings.OutputWav)
	if err != nil {
		fmt.Printf("Error creating output file: %s\n", err)
		return err
	}

	var outStream *WriteStreamer = new(WriteStreamer)
	outStream.Data = samples

	err = wav.Encode(outWAVFile, outStream, wavFormat)
	if err != nil {
		fmt.Printf("Error writing samples: %s\n", err)
		return err
	}

	return nil
}
