package av

import (
	"testing"
	"time"
)

func TestRateDur(t *testing.T) {
	f1 := Hz(8000)
	f2 := Hz(1960000)
	tests := []struct {
		Rate
		count int
		want  time.Duration
	}{
		{f1, 0, 0},
		{f1, 8000, time.Second},
		{f1, 4000, 500 * time.Millisecond},
		{f1, 40, 5 * time.Millisecond},
		{f1, 4, 500 * time.Microsecond},
		{f1, 2, 250 * time.Microsecond},
		{f1, 1, 125 * time.Microsecond},
		{f2, 1, 510 * time.Nanosecond},
	}
	for _, test := range tests {
		got := test.Dur(test.count)
		if got != Dur(test.want) {
			t.Errorf("%s dur %d got %s want %s", test.Rate, test.count, got, Dur(test.want))
		}
		count := test.Beats(got)
		if count != test.count {
			t.Errorf("%s count %s got %d want %d", test.Rate, got, count, test.count)
		}
	}
}
