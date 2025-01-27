// Package pgn uses data from canboat.json to convert NMEA 2000 messages to strongly-typed golang data.
package pgn

import (
	"fmt"
)

// MaxPGNLength is the maximum length of a PGN in bytes
const MaxPGNLength = 223 // 31*7 + 6

// PgnInfoLookup is a map of PGNs to PgnInfo pointers.
var PgnInfoLookup map[uint32][]*PgnInfo

// UnseenLookup is a map of PGNs not yet seen in log files to PgnInfo pointers.
var UnseenLookup map[uint32][]*PgnInfo

// init initializes PgnInfoLookup from pgnList (defined in pgninfo_generated.go)
func init() {
	PgnInfoLookup = make(map[uint32][]*PgnInfo)
	UnseenLookup = make(map[uint32][]*PgnInfo)

	for i, pi := range pgnList {
		if PgnInfoLookup[pi.PGN] == nil {
			PgnInfoLookup[pi.PGN] = make([]*PgnInfo, 0)
		}
		PgnInfoLookup[pi.PGN] = append(PgnInfoLookup[pi.PGN], &pgnList[i])
	}

	for i, pi := range unseenList {
		if UnseenLookup[pi.PGN] == nil {
			UnseenLookup[pi.PGN] = make([]*PgnInfo, 0)
		}
		UnseenLookup[pi.PGN] = append(UnseenLookup[pi.PGN], &unseenList[i])
	}
}

// IsProprietaryPGN returns true if its argument is in one of the proprietary ranges.
func IsProprietaryPGN(pgn uint32) bool {
	if pgn >= 0x0EF00 && pgn <= 0x0EFFF {
		// proprietary PDU1 (addressed) single-frame range 0EF00 to 0xEFFF (61184 - 61439) messages.
		// Addressed means that you send it to specific node on the bus. This you can easily use for responding,
		// since you know the sender. For sender it is bit more complicate since your device address may change
		// due to address claiming. There is N2kDeviceList module for handling devices on bus and find them by
		// "NAME" (= 64 bit value set by SetDeviceInformation ).
		return true
	} else if pgn >= 0x0FF00 && pgn <= 0x0FFFF {
		// proprietary PDU2 (non addressed) single-frame range 0xFF00 to 0xFFFF (65280 - 65535).
		// Non addressed means that destination wil be 255 (=broadcast) so any capable device can handle it.
		return true
	} else if pgn >= 0x1EF00 && pgn <= 0x1EFFF {
		// proprietary PDU1 (addressed) fast-packet PGN range 0x1EF00 to 0x1EFFF (126720 - 126975)
		return true
	} else if pgn >= 0x1FF00 && pgn <= 0x1FFFF {
		// proprietary PDU2 (non addressed) fast packet range 0x1FF00 to 0x1FFFF (130816 - 131071)
		return true
	}

	return false
}

// GetProprietaryInfo returns Manufacturer and Industry constants.
// Invalid data returned if called on non-proprietary PGNs.
func GetProprietaryInfo(data []uint8) (ManufacturerCodeConst, IndustryCodeConst, error) {
	stream := NewDataStream(data)
	var man ManufacturerCodeConst
	var ind IndustryCodeConst
	var err error
	var v uint64
	if v, err = stream.readLookupField(11); err == nil {
		man = ManufacturerCodeConst(v)
	}
	_ = stream.skipBits(2)
	if v, err = stream.readLookupField(3); err == nil {
		ind = IndustryCodeConst(v)
	}
	return man, ind, err
}

// GetFieldDescriptor returns the FieldDescriptor for the specified PGN variant.
func GetFieldDescriptor(pgn uint32, manID ManufacturerCodeConst, fieldIndex uint8) (*FieldDescriptor, error) {
	var retval *FieldDescriptor
	var err error

	if pi, piKnown := PgnInfoLookup[pgn]; piKnown {
		if !IsProprietaryPGN(pgn) { // should validate other match fields, but for now return the first
			retval = pi[0].Fields[int(fieldIndex)]
		} else {
			if manID != 0 { // we have a manufacturer to match against
				for _, p := range pi {
					if p.ManId == manID {
						retval = p.Fields[int(fieldIndex)]
					}
				}
			} else { // proprietary, but no manid, so no way to distinguish
				if len(pi) == 1 { // we only know of one variant, so we can give it a shot
					retval = pi[0].Fields[int(fieldIndex)]
				} else {
					err = fmt.Errorf("error: cannot distinguish between variants for pgn: %d", pgn)
					return nil, err
				}
			}

		}
		if retval == nil {
			err = fmt.Errorf("error: Field Index: %d, not found for pgn: %d with manufacturer code: %d", fieldIndex, pgn, manID)
		}
		return retval, err
	}
	return nil, fmt.Errorf("PGN not found")
}

// SearchUnseenList returns true if the PGN has no Canboat samples.
func SearchUnseenList(pgn uint32) bool {
	return UnseenLookup[pgn] != nil
}

// IsFast returns true if the specified PGN is Fast
func IsFast(pgn uint32) bool {
	if pi, piKnown := PgnInfoLookup[pgn]; piKnown {
		return pi[0].Fast
	}
	return false // should never be called with an invalid PGN, but avoid a panic
}

// GetPgnInfo returns the first PgnInfo for a given PGN, or nil if not found
func GetPgnInfo(pgn uint32) *PgnInfo {
	if pis, exists := PgnInfoLookup[pgn]; exists && len(pis) > 0 {
		return pis[0]
	}
	return nil
}
