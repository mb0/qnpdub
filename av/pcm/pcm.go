package pcm

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/mb0/qnpdub/av"
)

var (
	S8    = PCM{true, 1, nil}
	U8    = PCM{false, 1, nil}
	S16LE = PCM{true, 2, binary.LittleEndian}
	S16BE = PCM{true, 2, binary.BigEndian}
	U16LE = PCM{false, 2, binary.LittleEndian}
	U16BE = PCM{false, 2, binary.BigEndian}
)

type Format struct {
	PCM
	av.Rate
}

func (f Format) String() string {
	return fmt.Sprintf("%s_%d", f.PCM, f.Rate.Num)
}

type PCM struct {
	Sign  bool
	Bytes int
	binary.ByteOrder
}

func (pcm PCM) Add8(b []byte, res []int16) []int16 {
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

func (pcm PCM) Add16(b []byte, res []int16) []int16 {
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
