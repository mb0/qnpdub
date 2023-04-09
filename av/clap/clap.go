package clap

import (
	"fmt"
	"os"

	"github.com/mb0/qnpdub/av"
	"github.com/mb0/qnpdub/av/ffm"
	"github.com/mb0/qnpdub/av/pcm"
	"github.com/mb0/qnpdub/peak"
)

var (
	defFormat = pcm.Format{PCM: pcm.S8, Rate: av.Hz(8000)}
	defChunk  = 8 << 10
)

// Detector is a helper for clap detection in audio or video files.
type Detector struct {
	Format pcm.Format
	Chunk  int
	*peak.Detector[int16]
	bbuf []byte  // byte chunk buf
	sbuf []int16 // sample chunk buf
}

// New returns a new clap detector with the given waveform format and chunk size in bytes.
func New(f pcm.Format, chunk int) *Detector {
	// we detect with lag of a quarter chunk, that is 256ms or 2k samples at 8khz.
	sc := chunk / f.Bytes
	return &Detector{Format: f, Chunk: chunk,
		Detector: peak.New[int16](0, 3, sc/4, sc/2),
		bbuf:     make([]byte, chunk),
		sbuf:     make([]int16, sc),
	}
}

// Default returns a new detector with 8khz-8bit-format at 8k chunk size (16kb, 8kb buffer, 1.024s).
func Default() *Detector {
	return New(defFormat, defChunk)
}

// Load returns a waveform for the given media file path or an error.
// It generates the waveform file alongside the media file, if it does not exist.
func (d *Detector) Load(path string) (*pcm.File, error) {
	dest := fmt.Sprintf("%s.%s", path, d.Format.String())
	err := d.checkFile(dest, "wavf")
	if err != nil {
		err = d.checkFile(path, "media")
		if err != nil {
			return nil, err
		}
		err = ffm.GenPCMCmd(path, dest, d.Format).Run()
		if err != nil {
			return nil, fmt.Errorf("wavf gen failed: %w", err)
		}
	}
	return pcm.Open(dest, d.Format)
}

// LoadAll returns a list of waveforms for the given media file path or the first error.
func (d *Detector) LoadAll(paths ...string) ([]*pcm.File, error) {
	ws := make([]*pcm.File, 0, len(paths))
	for _, path := range paths {
		w, err := d.Load(path)
		if err != nil {
			return nil, err
		}
		ws = append(ws, w)
	}
	return ws, nil
}

// Detect returns a list of offsets of significant peaks at the end of w or an error.
func (d *Detector) Detect(w *pcm.File, n int) ([]int, error) {
	if w == nil || w.Count == 0 {
		return nil, fmt.Errorf("empty file")
	}
	d.Reset()
	r := av.NewChunkReader(w, d.bbuf)
	pro := av.Probe(*r, w.Count*w.Bytes, true)
	// one chunks give us 0.768s silence data (1.024s - 0.256s warmup lag)
	c, err := d.readChunks(pro, 1)
	if err != nil {
		return nil, err
	}

	// we collect n chunks with peaks and select the loudest of each
	loud := make([]peak.Peaks[int16], 0, n*4)
	if pk := c.Peaks[0]; len(pk.Sigs) > 0 {
		loud = append(loud, pk)
	}
	step := av.Chunks(d.Chunk, w.Bytes*int(w.Beats(5*av.S)))
	var max int16
Probe:
	for i := 0; i*step < pro.Max; i++ {
		c, err = d.readChunks(pro, step)
		if err != nil {
			return nil, err
		}
		for _, pk := range c.Peaks {
			if len(pk.Sigs) == 0 {
				continue
			}
			if pk.Max > max {
				max = pk.Max
			}
			loud = append(loud, pk)
			if len(loud) > n*3 {
				break Probe
			}
		}
	}
	offs := make([]int, 0, n)
	for _, pk := range loud {
		if pk.Max >= max/3 {
			offs = append(offs, pk.Mao)
			if len(offs) >= n {
				break
			}
		}
	}
	return offs, nil
}
func (d *Detector) readChunks(pro *av.Prober, n int) (chunks, error) {
	c := chunks{Peaks: make([]peak.Peaks[int16], 0, n)}
	err := pro.Next(n, func(off int, buf []byte) error {
		add := pcm.PCM.Add8
		if d.Format.Bytes == 2 {
			add = pcm.PCM.Add16
		}
		if sn := len(buf) / d.Format.Bytes; cap(d.sbuf) < sn {
			d.sbuf = make([]int16, sn)
		}
		d.sbuf = add(d.Format.PCM, buf, d.sbuf[:0])
		d.ResetIdx()
		soff := off / d.Format.Bytes
		if c.Off == 0 {
			c.Off = off / d.Format.Bytes
		}
		pk := d.Feed(soff, d.sbuf...)
		c.Peaks = append(c.Peaks, pk)
		return nil
	})
	if pro.Rev {
		reverse(c.Peaks)
	}
	return c, err
}

func (d *Detector) checkFile(path, typ string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s %q not found: %w", typ, path, err)
	} else if fi.Size() == 0 {
		return fmt.Errorf("%s %q empty", typ, path)
	}
	return nil
}

// chunks holds a starting offset and a list of peaks for each chunk read.
type chunks struct {
	Off   int // in samples
	Peaks []peak.Peaks[int16]
}

func reverse[T any](s []T) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
