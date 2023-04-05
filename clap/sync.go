package clap

import (
	"fmt"
	"time"

	"github.com/mb0/qnpdub/wavf"
)

// SyncPaths syncronizes the media files at paths by clap detection to a frame rate.
func (d *Detector) SyncPaths(rate int, paths ...string) ([]time.Duration, error) {
	ws := make([]*wavf.W, 0, len(paths))
	for _, path := range paths {
		w, err := d.Load(path)
		if err != nil {
			return nil, err
		}
		ws = append(ws, w)
	}
	return d.SyncWavfs(rate, ws...)
}

// SyncWavfs syncronizes the waveforms by clap detection to a frame rate.
func (d *Detector) SyncWavfs(rate int, ws ...*wavf.W) ([]time.Duration, error) {
	if len(ws) < 2 {
		return nil, fmt.Errorf("needs at least two waveforms")
	}
	const n = 8
	webs := make([]Web, 0, len(ws))
	var ldex [n][]int // length index
	for i, w := range ws {
		// detect n signals from each waveform
		off, err := d.Detect(w, n)
		if err != nil {
			return nil, err
		}
		if len(off) < 1 {
			return nil, fmt.Errorf("sync empty %q", w.Path)
		}
		l := len(off) - 1
		// collect by length and compute dist web
		ldex[l] = append(ldex[l], i)
		webs = append(webs, DistWeb(off))
	}
	// collect indices from ldex, sorted by length  descending.
	var hilo []int
	for i := range ldex {
		hilo = append(hilo, ldex[len(ldex)-1-i]...)
	}
	// match webs and collect claps and max clap offset
	var max int
	claps := make([]int, 0, len(ws))
	lst := webs[hilo[0]]
	for i, idx := range hilo[1:] {
		cur := webs[idx]
		m := match(lst, cur)
		if o := m.bc + m.bo; o > max {
			max = o
		}
		ac, bc := m.ac, m.bc
		if m.rev {
			ac, bc = m.bc, m.ac
		}
		if i == 0 {
			claps = append(claps, ac)
		}
		claps = append(claps, bc)
		lst = cur
	}
	// calculate offsets relative to max clap
	res := make([]time.Duration, 0, len(ws))
	for _, clap := range claps {
		// sync offset to frame rate
		ff := d.Rate / rate
		off := d.Duration(max - clap + ff)
		res = append(res, off)
	}
	return res, nil
}

func match(a, b Web) (m matcher) {
	m.a, m.b = a, b
	// select the web with max distance
	am, bm := a.MonoMax(), b.MonoMax()
	m.rev = bm > am
	if m.rev {
		m.a, m.b = b, a
	}
	if len(m.a.Vals) == 0 || len(m.b.Vals) == 0 {
		m.ac, m.bc = -1, -1
		m.ao, m.bo = -1, -1
		return
	}
	// default to head
	m.ac, m.bc = m.a.Vals[0], m.b.Vals[0]
	// calc default offsets
	m.calcOffs()
	m.find()
	return
}

type matcher struct {
	a, b   Web
	ac, bc int
	ao, bo int
	rev    bool
}

func (m *matcher) find() {
	// find better match if we have at least three and two offsets
	if !(len(m.a.Vals) >= 3 && len(m.b.Vals) >= 2) {
		return
	}
	max := 0 // we have one default match
	for i, bc := range m.b.Vals {
		j, score := matchRows(m.b.Row(i), m.a)
		if score > max {
			max, m.bc, m.ac = score, bc, m.a.Vals[j]
		}
	}
	m.calcOffs()
}

func matchRows(br []int, a Web) (idx, max int) {
	for _, bd := range br {
		for j := range a.Vals {
			score := 0
			ar := a.Row(j)
			for _, ad := range ar {
				if ad == bd {
					score++
				}
			}
			if score > max {
				max = score
				idx = j
			}
		}
	}
	return idx, max
}

func (m *matcher) calcOffs() {
	d := m.ac - m.bc
	if d < 0 {
		m.ao, m.bo = -d, 0
	} else {
		m.ao, m.bo = 0, d
	}
}

// Web caches all distances between values.
type Web struct {
	Vals []int
	Dist []int
}

// DistWeb calculates and returns the distance web for vals.
func DistWeb(vals []int) Web {
	dist := make([]int, 0, len(vals)*(len(vals)-1)/2)
	for i, src := range vals {
		for _, dst := range vals[i+1:] {
			d := src - dst
			if d < 0 {
				d *= -1
			}
			dist = append(dist, d)
		}
	}
	return Web{Vals: vals, Dist: dist}
}

// MonoMax returns the maximum distance for monotone sequences.
func (w Web) MonoMax() int {
	if idx := len(w.Vals) - 2; idx >= 0 && idx < len(w.Dist) {
		return w.Dist[idx]
	}
	return 0
}

// Row returns the distance row for one value by index.
func (w Web) Row(idx int) []int {
	if idx < 0 || idx > len(w.Vals) || len(w.Dist) < 1 {
		return nil
	}
	o, n := 0, len(w.Vals)-1
	row := make([]int, 0, n)
	// select the vertical component
	for i := 0; i < idx; i++ {
		row = append(row, w.Dist[o+idx-1-i])
		o += n - i
	}
	// select the horizontal component
	return append(row, w.Dist[o:o+n-idx]...)
}
