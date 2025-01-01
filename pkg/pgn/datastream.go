package pgn

import "math"

// DataStream instances provide methods to read data types from a stream.
// byteOffset and bitOffset combine to act as the read "cursor".
// The low level read functions update the cursor.
type DataStream struct {
	data []uint8

	byteOffset uint16
	bitOffset  uint8
}

// GetData returns the DataStream's current contents
func (d *DataStream) GetData() []uint8 {
	return d.data[:d.byteOffset]
}

// NewDataStream returns a new DataStream. Call it with the data from a complete Packet.
func NewDataStream(data []uint8) *DataStream {
	return &DataStream{
		data:       data,
		byteOffset: 0,
		bitOffset:  0,
	}
}

// getBitOffset method returns the cursor in bits.
func (s *DataStream) getBitOffset() uint32 {
	return uint32(s.byteOffset)*8 + uint32(s.bitOffset)
}

// resetToStart method resets the stream. Commented out since its currently unused.
func (s *DataStream) resetToStart() {
	s.byteOffset = 0
	s.bitOffset = 0
}

// calcMaxPositiveValue calculates the maximum value that can be represented
// with a given length of signed or unsigned contents.
func calcMaxPositiveValue(bitLength uint16, signed bool) uint64 {
	// calculate maximum valid value
	maxVal := uint64(0xFFFFFFFFFFFFFFFF)

	maxVal >>= 64 - bitLength // the largest value representable in length of field
	if signed {               // high bit set means it's negative, so maximum positive value is 1 bit shorter
		maxVal >>= 1 // we know it's a positive value, so safe for us to check.
	}
	switch bitLength {
	case 1: // leave alone
	case 2, 3: // for fields < 4 bits long, largest possible positive value indicates the field is missing
		maxVal -= 1
	default: // for larger fields, largest positive value means missing, that value minus 1 means invalid
		maxVal -= 2
	}
	return maxVal
}

// missingValue calculates the value representing a missing (nil) wire value
func missingValue(bitLength uint16, signed bool) uint64 {
	missing := uint64(0xFFFFFFFFFFFFFFFF)
	missing >>= 64 - bitLength // the largest value representable in length of field if unsigned
	if signed {                // high bit set means it's negative, so maximum positive value is 1 bit shorter
		missing >>= 1 // missing flag is max positive value; negative value has high bit set
	}
	return missing
}

// calcPrecision calculates the resulting precision of applying a given resolution to a given value
func calcPrecision(resolution float64) uint8 {
	precision := resolution
	digits := uint8(0)
	for {
		if precision < 0 || precision >= 1.0 {
			break
		}
		precision *= 10
		digits++
	}
	return digits
}

// roundFloat rounds a float64 to the specified precision
func roundFloat(val float64, precision uint8) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

// roundFloat32 rounds a float32 to the specified precision
func roundFloat32(val float32, precision uint8) float32 {
	ratio := math.Pow(10, float64(precision))
	return float32(math.Round(float64(val)*ratio) / ratio)
}
