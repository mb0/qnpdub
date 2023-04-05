package peak

import (
	"testing"
)

func TestDetector(t *testing.T) {
	data1 := []float64{1, 1, 1.1, 1, 0.9, 1, 1, 1.1, 1, 0.9, 1, 1.1, 1, 1, 0.9, 1, 1, 1.1, 1, 1, 1, 1, 1.1, 0.9, 1, 1.1, 1, 1, 0.9, 1, 1.1, 1, 1, 1.1, 1, 0.8, 0.9, 1, 1.2, 0.9, 1, 1, 1.1, 1.2, 1, 1.5, 1, 3, 2, 5, 3, 2, 1, 1, 1, 0.9, 1, 1, 3, 2.6, 4, 3, 3.2, 2, 1, 1, 0.8, 4, 4, 2, 2.5, 1, 1, 1}
	tests := []struct {
		dec  *Detector[float64]
		data []float64
		want []int
	}{
		{dec: New[float64](0, 5, 30, 0), data: data1,
			want: []int{45, 47, 48, 49, 50, 51, 58, 59, 60, 61, 62, 63, 67, 68, 69, 70},
		},
		{dec: New[float64](.1, 5, 30, 0), data: data1,
			want: []int{45, 47, 48, 49, 50, 58, 59, 60, 61, 62, 67, 68},
		},
		{dec: New[float64](0, 10, 30, 0), data: data1,
			want: []int{47, 49},
		},
		{dec: New[float64](.25, 5, 30, 0), data: data1,
			want: []int{45, 47, 48, 49, 50},
		},
		{dec: New[float64](.5, 5, 30, 0), data: data1,
			want: []int{45, 47, 49},
		},
	}
	for _, test := range tests {
		d := test.dec
		got := d.Feed(0, test.data...)
		if len(got.Sigs) != len(test.want) {
			t.Errorf("signal len got %d %v want %d", len(got.Sigs), got, len(test.want))
			continue
		}
		for i, g := range got.Sigs {
			w := test.want[i]
			if g.Idx != w {
				t.Errorf("signal idx %d got %d want %d", i, g.Idx, w)
			}
			if v := test.data[w]; g.Val != v {
				t.Errorf("signal val %d got %g want %g", i, g.Val, v)
			}
		}
	}
}
