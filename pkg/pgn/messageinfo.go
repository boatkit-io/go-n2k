package pgn

import (
	"time"

	"github.com/boatkit-io/n2k/pkg/converter"
	"github.com/brutella/can"
)

// MessageInfo contains context needed to process an NMEA 2000 message.
type MessageInfo struct {
	// when did we get the message
	Timestamp time.Time

	// 3-bit
	Priority uint8

	// 19-bit number
	PGN uint32

	// actually 8-bit
	SourceId uint8

	// target address, when relevant (PGNs with PF < 240)
	TargetId uint8
}

// NewMessageInfo provides context for a canbus Frame.
func NewMessageInfo(message *can.Frame) MessageInfo {
	h := converter.DecodeCanId(message.ID)
	p := MessageInfo{
		Timestamp: h.TimeStamp,
		SourceId:  h.SourceId,
		PGN:       h.PGN,
		Priority:  h.Priority,
	}
	return p
}
