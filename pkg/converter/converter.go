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

// CanFrameFromRaw parses an input string into one or more can.Frames.
// For data longer than 8 bytes, it will create multiple frames according to the ISO-TP protocol.
func CanFrameFromRaw(in string) ([]*can.Frame, error) {
	elems := strings.Split(in, ",")
	if len(elems) < 6 {
		return nil, fmt.Errorf("invalid raw format: insufficient elements")
	}

	priority, err := strconv.ParseUint(elems[1], 10, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid priority: %w", err)
	}
	pgn, err := strconv.ParseUint(elems[2], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid pgn: %w", err)
	}
	source, err := strconv.ParseUint(elems[3], 10, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid source: %w", err)
	}
	destination, err := strconv.ParseUint(elems[4], 10, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid destination: %w", err)
	}
	length, err := strconv.ParseUint(elems[5], 10, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid length: %w", err)
	}

	if int(length) > len(elems)-6 {
		return nil, fmt.Errorf("invalid raw format: data length exceeds available bytes")
	}

	id := CanIdFromData(uint32(pgn), uint8(source), uint8(priority), uint8(destination))

	// For data <= 8 bytes, return a single frame
	if length <= 8 {
		frame := &can.Frame{
			ID:     id,
			Length: uint8(length),
		}
		for i := 0; i < int(length); i++ {
			b, err := strconv.ParseUint(elems[i+6], 16, 8)
			if err != nil {
				return nil, fmt.Errorf("invalid data byte at position %d: %w", i, err)
			}
			frame.Data[i] = uint8(b)
		}
		return []*can.Frame{frame}, nil
	}

	// For data > 8 bytes, create multiple frames
	var frames []*can.Frame
	remainingBytes := int(length)
	dataIndex := 6 // Start of data in elems

	// First frame
	firstFrame := &can.Frame{
		ID:     id,
		Length: 8,
	}
	// First byte: 0x1_ where _ is high nibble of length
	firstFrame.Data[0] = 0x10 | uint8((length>>8)&0x0F)
	// Second byte: low byte of length
	firstFrame.Data[1] = uint8(length & 0xFF)

	// Copy up to 6 bytes of data
	for i := 0; i < min(6, remainingBytes); i++ {
		b, err := strconv.ParseUint(elems[dataIndex+i], 16, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid data byte at position %d: %w", i, err)
		}
		firstFrame.Data[i+2] = uint8(b)
	}
	frames = append(frames, firstFrame)

	remainingBytes -= 6
	dataIndex += 6
	seqNum := uint8(1)

	// Consecutive frames
	for remainingBytes > 0 {
		frame := &can.Frame{
			ID:     id,
			Length: 8,
		}
		frame.Data[0] = 0x20 | (seqNum & 0x0F) // Consecutive frame PCI byte

		// Copy up to 7 bytes of data
		bytesToCopy := min(7, remainingBytes)
		for i := 0; i < bytesToCopy; i++ {
			b, err := strconv.ParseUint(elems[dataIndex+i], 16, 8)
			if err != nil {
				return nil, fmt.Errorf("invalid data byte at position %d: %w", dataIndex+i, err)
			}
			frame.Data[i+1] = uint8(b)
		}

		frames = append(frames, frame)
		remainingBytes -= bytesToCopy
		dataIndex += bytesToCopy
		seqNum++
	}

	return frames, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
