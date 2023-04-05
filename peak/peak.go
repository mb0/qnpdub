// Package peak implements the famous peak detection algorithm from stackoverflow
//
// Brakel, J.P.G. van (2014). "Robust peak detection algorithm using z-scores". Stack Overflow.
// Available at: https://stackoverflow.com/questions/22583391/peak-signal-detection-in-realtime-timeseries-data/22640362#22640362 (version: 2020-11-08).
package peak

import (
	"math"

	"golang.org/x/exp/constraints"
)

type Num interface {
	constraints.Integer | constraints.Float
}

// Peaks contains a summery, offset and signals for a chunk of values.
type Peaks[N Num] struct {
	Sum[N]
	Off  int
	Sigs []Sig[N]
}

// Signal holds the index and value of a peak.
type Sig[N Num] struct {
	Idx int
	Val N
}

// Summery holds aggregate detauls for a chunk of values.
type Sum[N Num] struct {
	Mean, Vari float64
	Min, Max   N
	Mio, Mao   int
	Len        int
}

// Show adds the value at off to the summery.
func (sum *Sum[N]) Show(off int, val N) {
	if sum.Len == 0 {
		sum.Mio, sum.Mao = off, off
		sum.Min, sum.Max = val, val
	} else if val < sum.Min {
		sum.Mio = off
		sum.Min = val
	} else if val > sum.Max {
		sum.Mao = off
		sum.Max = val
	}
	sum.Len++
}

// Merge merges the length and minmax of o into sum.
func (sum *Sum[N]) Merge(o Sum[N]) {
	if o.Len > 0 {
		sum.Len += o.Len - 2
		sum.Show(o.Mio, o.Min)
		sum.Show(o.Mao, o.Max)
	}
}

// Stdd returns the standard deviation for this chunk.
func (sum Sum[N]) Stdd() float64 {
	return math.Sqrt(sum.Vari)
}

// Detector is a helper to detect peaks in value sequences.
type Detector[N Num] struct {
	conf
	Sum[N]
	win []float64
	val float64
	idx int
}

// New returns a new peak detector with influence, threshold, lag and warmup.
// Influence is a factor to increase the effect of signals on the threshold.
// Threshold is the number of standard deviations form the moving mean to qualify as peak.
// Lag is the number of signals in the window of the moving mean.
// Warmup is the number of signals before peaks are detected.
func New[N Num](infl, trsh float64, lag, warm int) *Detector[N] {
	if warm < lag {
		warm = lag
	}
	return &Detector[N]{conf: conf{infl: infl, trsh: trsh, lag: lag, warm: warm}}
}

// Feed finds and returns peaks for a chunk of values at an offset.
// The offset is returned in the result and used for min and max offsets.
func (d *Detector[N]) Feed(off int, vals ...N) Peaks[N] {
	if d.win == nil {
		d.win = make([]float64, d.lag)
	}
	r := Peaks[N]{Off: off}
	var v, w float64
	for i, val := range vals {
		wi := d.idx % d.lag
		v = float64(val)
		if d.Vari != 0 {
			if math.Abs(v-d.Mean) > d.trsh*d.Stdd() {
				r.Sigs = append(r.Sigs, Sig[N]{Idx: d.idx, Val: val})
				v = d.infl*v + (1-d.infl)*d.val
			}
			w = d.win[wi]
			p := (v - w) / float64(d.lag)
			d.Vari += (v + w - 2*d.Mean - p) * p
			d.Mean += p
		}
		r.Show(off+i, val)
		d.win[wi] = v
		d.val = v
		d.idx++
		if d.idx%d.warm == 0 && d.Vari == 0 {
			warmup(d)
		}
	}
	d.Merge(r.Sum)
	r.Mean, r.Vari = d.Mean, d.Vari
	return r
}

// Reset reverts the detector to its initial condition.
func (d *Detector[N]) Reset() { *d = Detector[N]{conf: d.conf} }

// ResetIdx resets the detector index only.
func (d *Detector[N]) ResetIdx() { d.idx = 0 }

type conf struct {
	infl, trsh float64
	lag, warm  int
}

func warmup[N Num](d *Detector[N]) {
	var pre, sqs float64
	for i, val := range d.win {
		if i == 0 {
			pre = val
			continue
		}
		mean := pre + (val-pre)/float64(i+1)
		sqs += (val - pre) * (val - mean)
		pre = mean
	}
	d.Mean = pre
	d.Vari = sqs / float64(len(d.win))
}
