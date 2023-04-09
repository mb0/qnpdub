package av

import (
	"fmt"
	"io"
)

// Returns the number of chunks required for the number of values.
func Chunks(chunk, n int) int {
	if n > 0 && chunk > 0 {
		return (n + chunk - 1) / chunk
	}
	return 0
}

// ChunkReader is a read seeker wrapper that allows to read byte chunks.
type ChunkReader struct {
	io.ReadSeeker
	Chunk int // bytes per chunk
	buf   []byte
}

func NewChunkReader(r io.ReadSeeker, chk []byte) *ChunkReader {
	return &ChunkReader{r, len(chk), chk}
}

// ReadChunks seek to byte offset and reads n bytes in chunks.
func (c *ChunkReader) ReadChunks(off, n int, h func(int, []byte) error) error {
	_, err := c.Seek(int64(off), io.SeekStart)
	if err != nil {
		return fmt.Errorf("seek %d failed: %w", off, err)
	}
	if c.Chunk == 0 {
		return fmt.Errorf("invalid chunk size %d", c.Chunk)
	} else if cap(c.buf) < c.Chunk {
		c.buf = make([]byte, c.Chunk)
	}
	for rest := n; rest > 0; rest -= c.Chunk {
		l := c.Chunk
		if rest < l {
			l = rest
		}
		_, err = io.ReadFull(c.ReadSeeker, c.buf[:l])
		if err != nil {
			return err
		}
		err = h(off, c.buf[:l])
		if err != nil {
			return err
		}
		off += c.Chunk
	}
	return nil
}

func (c *ChunkReader) Close() error {
	if cc, ok := c.ReadSeeker.(io.Closer); ok {
		return cc.Close()
	}
	return nil
}

type Prober struct {
	ChunkReader
	Rev      bool
	Cur, Max int // in chunks
	Size     int // in bytes
}

// Probe returns a new prober with the given settings and waveform.
func Probe(cr ChunkReader, size int, rev bool) *Prober {
	p := &Prober{ChunkReader: cr, Rev: rev, Max: Chunks(cr.Chunk, size), Size: size}
	if rev {
		p.Cur = p.Max
	}
	return p
}

// Next returns the n chunks in probe direction or an error.
func (p *Prober) Next(n int, hand func(int, []byte) error) error {
	if p.Rev {
		p.Cur -= n
		if p.Cur < 0 {
			n -= p.Cur
			p.Cur = 0
		}
	}
	off := p.Cur * p.Chunk
	if p.Rev {
		off = p.Size - (p.Max-p.Cur)*p.Chunk
		if off < 0 {
			off = 0
		}
	}
	err := p.ReadChunks(off, n*p.Chunk, hand)
	if !p.Rev {
		p.Cur += n
	}
	return err
}
