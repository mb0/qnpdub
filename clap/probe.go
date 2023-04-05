package clap

import (
	"github.com/mb0/qnpdub/peak"
	"github.com/mb0/qnpdub/wavf"
)

// Chunks holds a starting offset and a list of peaks for each chunk read.
type Chunks struct {
	Off   int // in samples
	Peaks []peak.Peaks[int16]
}

// Prober is a helper to probe waveform data in chunks.
// Probing can be backwards, from end to the start and reverse peaks.
type Prober struct {
	*Detector
	*wavf.W
	rev      bool
	cur, len int // in chunks
}

// Probe returns a new prober with the given settings and waveform.
func Probe(d *Detector, w *wavf.W, rev bool) *Prober {
	p := &Prober{d, w, rev, 0, d.Chunks(w.Count)}
	if rev {
		p.cur = p.len
	}
	return p
}

// Next returns the n chunks in probe direction or an error.
func (p *Prober) Next(n int) (Chunks, error) {
	if p.rev {
		p.cur -= n
		if p.cur < 0 {
			n -= p.cur
			p.cur = 0
		}
	}
	c := Chunks{Off: p.cur * p.N, Peaks: make([]peak.Peaks[int16], 0, n)}
	if p.rev {
		c.Off = p.Count - (p.len-p.cur)*p.N
		if c.Off < 0 {
			c.Off = 0
		}
	}
	err := p.Extract(p.W, c.Off, n*p.N, func(ch []int16) {
		p.ResetIdx()
		pk := p.Feed(c.Off+len(c.Peaks)*p.N, ch...)
		c.Peaks = append(c.Peaks, pk)
	})
	if !p.rev {
		p.cur += n
	} else {
		reverse(c.Peaks)
	}
	return c, err
}

func reverse[T any](s []T) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
