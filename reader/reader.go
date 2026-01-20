package reader

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/faiface/beep"
	"github.com/faiface/beep/wav"
)

func ReadBin(filename string) ([]uint32, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}
	defer file.Close()

	stats, statsErr := file.Stat()
	if statsErr != nil {
		return nil, statsErr
	}

	var size int64 = stats.Size()
	bytes := make([]byte, size)

	buf := bufio.NewReader(file)
	_, err = buf.Read(bytes)

	if err != nil {
		return nil, err
	}

	var ints []uint32
	for i := 0; i < len(bytes); i += 4 {
		ints = append(ints, binary.BigEndian.Uint32(bytes[i:i+4]))
	}

	return ints, err
}

// HEX exports from SpinCAD (Intel HEX format)
func ReadHex(filename string) ([]uint32, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}
	defer file.Close()

	rdr := bufio.NewReader(file)

	var bytes []byte
	for {
		line, _, err := rdr.ReadLine()
		if err == io.EOF {
			break
		}

		recordTypeStr := line[1+2+4 : 1+2+4+2]
		recordType := 0
		_, err = fmt.Sscanf(string(recordTypeStr), "%X", &recordType)
		if recordType == 1 {
			break
		} else if recordType != 0 {
			continue
		}

		numBytesStr := line[1 : 1+2]
		numBytes := 0
		_, err = fmt.Sscanf(string(numBytesStr), "%X", &numBytes)

		var vals []byte = make([]byte, numBytes)
		rawstr := line[len(line)-(2*numBytes)-2 : len(line)-2]
		_, err = fmt.Sscanf(string(rawstr), "%X", &vals)
		if err != nil {
			return nil, err
		}

		bytes = append(bytes, vals...)
	}

	// Collect all 32bits chunks
	var ints []uint32
	for i := 0; i < (len(bytes) - 4); i += 4 {
		chunk := []byte{bytes[i], bytes[i+1], bytes[i+2], bytes[i+3]}
		ints = append(ints, binary.BigEndian.Uint32(chunk))
	}

	return ints, err
}

func ReadWAV(filename string) (*os.File, beep.Streamer, beep.Format, error) {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	stream, wavFormat, err := wav.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	return f, stream, wavFormat, err
}
