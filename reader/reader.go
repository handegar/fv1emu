package reader

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
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

	var ints []uint32
	for {
		line, _, err := rdr.ReadLine()
		if err == io.EOF {
			break
		}

		if len(line) == 11 && string(line) == ":00000001FF" {
			break
		}

		rawstr := line[1+2+4+2 : len(line)-2]
		var val []byte = make([]byte, 4)
		_, err = fmt.Sscanf(string(rawstr), "%X", &val)
		if err != nil {
			return nil, err
		}

		ints = append(ints, binary.BigEndian.Uint32(val))
	}

	return ints, err
}
