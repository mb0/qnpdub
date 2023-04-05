package wavf

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

var (
	PCM_S8    = PCM{true, 1, nil}
	PCM_U8    = PCM{false, 1, nil}
	PCM_S16LE = PCM{true, 2, binary.LittleEndian}
	PCM_S16BE = PCM{true, 2, binary.BigEndian}
	PCM_U16LE = PCM{false, 2, binary.LittleEndian}
	PCM_U16BE = PCM{false, 2, binary.BigEndian}
)

type Format struct {
	PCM
	Rate int
}

func (f Format) Duration(count int) time.Duration {
	// number_of_samples divided by sample_rate per second, normalized to nano seconds
	return time.Duration(count) * time.Second / time.Duration(f.Rate)
}

func (f Format) Samples(dur time.Duration) int {
	return int((dur + 1) * time.Duration(f.Rate) / time.Second)
}
func (f Format) String() string {
	return fmt.Sprintf("%s_%d", f.PCM, f.Rate)
}

type PCM struct {
	Sign  bool
	Bytes int
	binary.ByteOrder
}

func (pcm PCM) add8(b []byte, res []int16) []int16 {
	for o := 0; o < len(b); o++ {
		s := b[o]
		var n int16
		if pcm.Sign {
			n = int16(int8(s))
		} else {
			n = int16(s) - 0x80
		}
		res = append(res, n*0x100) // scale to 16 bits
	}
	return res
}

func (pcm PCM) add16(b []byte, res []int16) []int16 {
	by := int(pcm.Bytes >> 3)
	for o := 0; o < len(b); o += by {
		s := pcm.Uint16(b[o:])
		var n int16
		if pcm.Sign {
			n = int16(s)
		} else {
			n = int16(int32(s) - 0x8000)
		}
		res = append(res, n)
	}
	return res
}

func (pcm PCM) String() string {
	var b strings.Builder
	b.WriteString("pcm_")
	if pcm.Sign {
		b.WriteByte('s')
	} else {
		b.WriteByte('u')
	}
	fmt.Fprint(&b, pcm.Bytes*8)
	switch pcm.ByteOrder {
	case binary.LittleEndian:
		b.WriteString("le")
	case binary.BigEndian:
		b.WriteString("be")
	}
	return b.String()
}
