package clap

import (
	"fmt"
	"os"
	"time"

	"github.com/mb0/qnpdub/peak"
	"github.com/mb0/qnpdub/wavf"
)

var (
	defFormat = wavf.Format{PCM: wavf.PCM_S8, Rate: 8000}
	defChunk  = 8 << 10
)

// Detector is a helper for clap detection in audio or video files.
type Detector struct {
	wavf.Format
	wavf.Chunked
	*peak.Detector[int16]
}

// New returns a new clap detector with the given waveform format and chunk size.
func New(f wavf.Format, chunk int) *Detector {
	// we detect with lag of a eighth chunk, that is 128ms or 1k samples at 8khz.
	return &Detector{f, wavf.Chunked{N: chunk}, peak.New[int16](0, 3, chunk/4, chunk/2)}
}

// Default returns a new detector with 8khz-8bit-format at 8k chunk size (16kb, 8kb buffer, 1.024s).
func Default() *Detector {
	return New(defFormat, defChunk)
}

// Load returns a waveform for the given media file path or an error.
// It generates the waveform file alongside the media file, if it does not exist.
func (d *Detector) Load(path string) (*wavf.W, error) {
	dest := fmt.Sprintf("%s.%s", path, d.Format.String())
	err := d.checkFile(dest, "wavf")
	if err != nil {
		err = d.checkFile(path, "media")
		if err != nil {
			return nil, err
		}
		err = wavf.GenCmd(path, dest, d.Format).Run()
		if err != nil {
			return nil, fmt.Errorf("wavf gen failed: %w", err)
		}
	}
	return wavf.Open(dest, d.Format)
}

// Detect returns a list of offsets of significant peaks at the end of w or an error.
func (d *Detector) Detect(w *wavf.W, n int) ([]int, error) {
	if w == nil || w.Count == 0 {
		return nil, fmt.Errorf("empty file")
	}
	d.Reset()
	probe := Probe(d, w, true)
	// one chunks give us 0.768s silence data (1.024s - 0.256s warmup lag)
	c, err := probe.Next(1)
	if err != nil {
		return nil, err
	}
	// we collect three chunks with peaks and select the loudest of each
	loud := make([]peak.Peaks[int16], 0, n*4)
	if pk := c.Peaks[0]; len(pk.Sigs) > 0 {
		loud = append(loud, pk)
	}
	step := d.Chunks(w.Samples(5 * time.Second))
	var max int16
Probe:
	for i := 0; i*step < probe.len; i++ {
		c, err = probe.Next(step)
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

func (d *Detector) checkFile(path, typ string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s %q not found: %w", typ, path, err)
	} else if fi.Size() == 0 {
		return fmt.Errorf("%s %q empty", typ, path)
	}
	return nil
}
