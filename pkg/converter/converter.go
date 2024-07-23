// Package converter provides routines that convert between various text
// data formats and Can frames.
package converter

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/brutella/can"
)

// CanFrameFromRaw parses an input string into a can.Frame.
func CanFrameFromRaw(in string) *can.Frame {
	elems := strings.Split(in, ",")
	priority, _ := strconv.ParseUint(elems[1], 10, 8)
	pgn, _ := strconv.ParseUint(elems[2], 10, 32)
	source, _ := strconv.ParseUint(elems[3], 10, 8)
	destination, _ := strconv.ParseUint(elems[4], 10, 8)
	length, _ := strconv.ParseUint(elems[5], 10, 8)

	id := CanIdFromData(uint32(pgn), uint8(source), uint8(priority), uint8(destination))
	retval := can.Frame{
		ID:     id,
		Length: 8,
	}
	for i := 0; i < int(length); i++ {
		b, _ := strconv.ParseUint(elems[i+6], 16, 8)
		retval.Data[i] = uint8(b)
	}

	return &retval
}

// CanIdFromData returns an encoded ID from its inputs.
func CanIdFromData(pgn uint32, sourceId uint8, priority uint8, destination uint8) uint32 {
	return uint32(sourceId) | (pgn << 8) | (uint32(priority) << 26) | uint32(destination)
}

// FrameHeader defines a structure to capture the RAW defined information comprising a CAN Frame ID
// and the recorded timestamp
type FrameHeader struct {
	TimeStamp time.Time
	SourceId  uint8
	PGN       uint32
	Priority  uint8
	TargetId  uint8
}

// DecodeCanId returns a frame header extracted from frame.Id
func DecodeCanId(id uint32) FrameHeader {
	r := FrameHeader{
		TimeStamp: time.Now(),
		SourceId:  uint8(id & 0xFF),
		PGN:       (id & 0x3FFFF00) >> 8,
		Priority:  uint8((id & 0x1C000000) >> 26),
	}

	pduFormat := uint8((r.PGN & 0xFF00) >> 8)
	if pduFormat < 240 {
		// This is a targeted packet, and the lower PS has the address
		r.TargetId = uint8(r.PGN & 0xFF)
		r.PGN &= 0xFFF00
	}
	return r

}

// RawFromCanFrame returns a string in RAW format encoding the frame
func RawFromCanFrame(f can.Frame) string {
	h := DecodeCanId(f.ID)
	return fmt.Sprintf("%s,%d,%d,%d,%d,%d,%02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x\n", h.TimeStamp.Format("2006-01-02T15:04:05Z"), h.Priority, h.PGN, h.SourceId, h.TargetId, f.Length, f.Data[0], f.Data[1], f.Data[2], f.Data[3], f.Data[4], f.Data[5], f.Data[6], f.Data[7])

}
