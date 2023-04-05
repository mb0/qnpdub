package wavf

import (
	"testing"
	"time"
)

func TestPCMString(t *testing.T) {
	tests := []struct {
		PCM
		want string
	}{
		{PCM_S16BE, "pcm_s16be"},
	}
	for _, test := range tests {
		got := test.PCM.String()
		if got != test.want {
			t.Errorf("got %s want %s", got, test.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	f1 := Format{PCM_S16BE, 8000}
	f2 := Format{PCM_S16BE, 1960000}
	tests := []struct {
		Format
		bytes int
		want  time.Duration
	}{
		{f1, 0, 0},
		{f1, 16000, time.Second},
		{f1, 8000, 500 * time.Millisecond},
		{f1, 80, 5 * time.Millisecond},
		{f1, 8, 500 * time.Microsecond},
		{f1, 4, 250 * time.Microsecond},
		{f1, 2, 125 * time.Microsecond},
		{f2, 2, 510 * time.Nanosecond},
	}
	for _, test := range tests {
		count := test.bytes / test.PCM.Bytes
		got := test.Duration(count)
		if got != test.want {
			t.Errorf("%s:%d for %d got %s want %s", test.PCM, test.Rate, test.bytes, got, test.want)
		}
		samples := test.Samples(got)
		if samples != count {
			t.Errorf("%s:%d for %d samples got %d want %d", test.PCM, test.Rate, test.bytes, samples, count)
		}
	}
}
