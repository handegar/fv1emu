package reader

import (
	"bufio"
	"encoding/binary"
	"os"
)

func ReadHEX(filename string) ([]uint32, error) {
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
