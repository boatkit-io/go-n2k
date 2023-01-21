//go:build testing

package testutils

import (
	"strconv"
	"strings"

	"github.com/boatkit-io/n2k/pkg/adapter/canadapter"
)

func CanFrameFromRaw(in string) canadapter.Frame {
	elems := strings.Split(in, ",")
	priority, _ := strconv.ParseUint(elems[1], 10, 8)
	pgn, _ := strconv.ParseUint(elems[2], 10, 32)
	source, _ := strconv.ParseUint(elems[3], 10, 8)
	destination, _ := strconv.ParseUint(elems[4], 10, 8)
	length, _ := strconv.ParseUint(elems[5], 10, 8)

	id := CanIdFromData(uint32(pgn), uint8(source), uint8(priority), uint8(destination))
	retval := canadapter.Frame{
		ID:     id,
		Length: 8,
	}
	for i := 0; i < int(length); i++ {
		b, _ := strconv.ParseUint(elems[i+6], 16, 8)
		retval.Data[i] = uint8(b)
	}

	return retval
}

func CanIdFromData(pgn uint32, sourceId uint8, priority uint8, destination uint8) uint32 {
	return uint32(sourceId) | (pgn << 8) | (uint32(priority) << 26) | uint32(destination)
}
