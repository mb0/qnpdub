package av

import (
	"testing"
	"time"
)

func TestParseDur(t *testing.T) {
	tests := []struct {
		raw  string
		want Dur
		str  string
	}{
		{"55", Dur(55 * time.Second), "55"},
		{"0.2", Dur(200 * time.Millisecond), "0.200"},
		{"0.12", Dur(120 * time.Millisecond), "0.120"},
		{"0.123us", Dur(123 * time.Nanosecond), "0.000000123"},
		{"200ms", Dur(200 * time.Millisecond), "0.200"},
		{"200000us", Dur(200 * time.Millisecond), "0.200"},
		{"12:03:45", Dur(12*time.Hour + 3*time.Minute + 45*time.Second), "12:03:45"},
		{"125:03:45", Dur(125*time.Hour + 3*time.Minute + 45*time.Second), "125:03:45"},
		{"23.189", Dur(23189 * time.Millisecond), "23.189"},
	}
	for _, test := range tests {
		got, err := ParseDur(test.raw)
		if err != nil {
			t.Errorf("parse dur %s err: %v", test.raw, err)
		}
		if got != test.want {
			t.Errorf("parse dur %s got %s want %s", test.raw, got, test.want)
			continue
		}
		if str := got.String(); str != test.str {
			t.Errorf("format dur %s got %s want %s", test.raw, str, test.str)
		}
		got, err = ParseDur("-" + test.raw)
		if err != nil || got > 0 {
			t.Errorf("parse minus dur -%s err: %v", test.raw, err)
		}
		if got != -test.want {
			t.Errorf("parse minus dur -%s got %s want %s", test.raw, got, -test.want)
			continue
		}
		if str := got.String(); str[0] != '-' || str[1:] != test.str {
			t.Errorf("format dur -%s got %s want -%s", test.raw, str, test.str)
		}
	}
	errs := []string{"55x", "0.-2", ".2", "123:45", "23.189:"}
	for _, str := range errs {
		_, err := ParseDur(str)
		if err == nil {
			t.Errorf("parse dur %s want err got none", str)
		}
	}
}
