package canadapter

import (
	"fmt"
	"testing"

	"github.com/boatkit-io/n2k/pkg/converter"
	"github.com/boatkit-io/n2k/pkg/pgn"
	"github.com/boatkit-io/n2k/pkg/pkt"

	"github.com/stretchr/testify/assert"
)

func TestPgn127501(t *testing.T) {
	raw := "2023-01-21T00:04:17Z,3,127501,224,0,8,00,03,c0,ff,ff,ff,ff,ff"
	f := converter.CanFrameFromRaw(raw)
	pInfo := pgn.NewMessageInfo(f)
	p := pkt.NewPacket(pInfo, f.Data[:])
	assert.NotEmpty(t, p.Candidates)
	p.AddDecoders()
	assert.Equal(t, len(p.Decoders), 1)
	decoder := p.Decoders[0]
	stream := pgn.NewDataStream(p.Data)
	ret, err := decoder(p.Info, stream)
	assert.Nil(t, err)
	assert.IsType(t, pgn.BinarySwitchBankStatus{}, ret)
}

func TestPgn127501Write(t *testing.T) {
	raw := "2023-01-21T00:04:17Z,3,127501,224,0,8,00,03,c0,ff,ff,ff,ff,ff"
	f := converter.CanFrameFromRaw(raw)
	pInfo := pgn.NewMessageInfo(f)
	p := pkt.NewPacket(pInfo, f.Data[:])
	assert.NotEmpty(t, p.Candidates)
	p.AddDecoders()
	assert.Equal(t, len(p.Decoders), 1)
	decoder := p.Decoders[0]
	stream := pgn.NewDataStream(p.Data)
	ret, err := decoder(p.Info, stream)
	assert.Nil(t, err)
	assert.IsType(t, pgn.BinarySwitchBankStatus{}, ret)
}

func TestRawToDataStream(t *testing.T) {

	var nonMatches = []string{
		"2024-08-27T14:36:06Z,2,130306,15,0,8,8a,ff,ff,ff,ff,00,ff,ff",
		"2024-08-27T14:36:06Z,2,130306,15,0,8,8a,ff,ff,ff,ff,02,ff,ff",
		"2024-08-27T14:36:06Z,2,130306,15,0,8,8a,ff,ff,ff,ff,03,ff,ff",
	}
	var goodStrings = []string{
		"2024-08-27T14:36:06Z,2,129026,43,0,8,62,ff,ff,ff,00,00,ff,ff",
		"2024-08-27T14:36:06Z,2,129025,8,0,8,8d,a5,27,19,c5,25,d9,d5",
		"2024-08-27T14:36:06Z,2,129026,15,0,8,95,fc,43,13,00,00,ff,ff",
		"2024-08-27T14:36:06Z,2,129025,43,0,8,e0,a2,27,19,f8,26,d9,d5",
		"2024-08-27T14:36:06Z,2,129025,15,0,8,8b,a5,27,19,ab,25,d9,d5",
	}
	var rawStrings = []string{
		"2024-08-27T14:36:06Z,2,129026,43,0,8,62,ff,ff,ff,00,00,ff,ff",
		"2024-08-27T14:36:06Z,2,129025,8,0,8,8d,a5,27,19,c5,25,d9,d5",
		"2024-08-27T14:36:06Z,2,129026,15,0,8,95,fc,43,13,00,00,ff,ff",
		"2024-08-27T14:36:06Z,2,129025,43,0,8,e0,a2,27,19,f8,26,d9,d5",
		"2024-08-27T14:36:06Z,2,129025,15,0,8,8b,a5,27,19,ab,25,d9,d5",
		"2024-08-27T14:36:06Z,2,129025,15,0,8,8b,a5,27,19,ab,25,d9,d5",
		"2024-08-27T14:36:06Z,2,129026,15,0,8,95,fc,43,13,00,00,ff,ff",
		"2024-08-27T14:36:06Z,2,129025,43,0,8,e0,a2,27,19,f8,26,d9,d5",
	}
	summary := false

	for _, s := range rawStrings {
		f := converter.CanFrameFromRaw(s)
		if f == nil {
			panic("bad input to TestRawToDataStream")
		}
		m := pgn.NewMessageInfo(f)
		assert.False(t, pgn.IsFast(m.PGN))
		pInfo := pgn.PgnInfoLookup[m.PGN]
		assert.True(t, len(pInfo) == 1)
		decoder := pInfo[0].Decoder
		d := pgn.NewDataStream(f.Data[:])
		p, err := decoder(m, d)
		assert.Nil(t, err)
		data := make([]uint8, 223)
		stream := pgn.NewDataStream(data)
		if v, ok := p.(pgn.PgnStruct); ok {
			_, err := v.Encode(stream)
			assert.Nil(t, err)
			//			assert.EqualValues(t, m, info)
			assert.ElementsMatch(t, stream.GetData(), f.Data)
		} else {
			panic("datastream test: encode didn't return a PgnStruct")
		}
		if summary {
			fmt.Printf("nonmatches:%d, good:%d\n", len(nonMatches), len(goodStrings))

		}
		fmt.Printf("nonmatches:%d\n", len(nonMatches))
	}
}
