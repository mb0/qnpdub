// Package av provides units and conversions for audio and video related code.
package av

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var S = Dur(time.Second)

// Dur is a time duration wrapper and therefor in nanoseconds.
type Dur time.Duration

// ParseDur parses durations in the two formats used by ffmpeg.
// -HH:MM:SS.mm or -?([0-9]+:)?[0-9]{1:2}:[0-9]{1:2}([.][0-9]+)?
// -S+.mm[s|ms|us] or -?[0-9]+([.][0-9]+)?(s|ms|us)?
func ParseDur(str string) (res Dur, err error) {
	neg := len(str) > 0 && str[0] == '-'
	if neg {
		str = str[1:]
	}
	org := str
	unit, un := durUnit(str)
	str = str[:len(str)-un]
	var cn int
	if un == 0 && strings.Contains(str, ":") {
		prts := strings.Split(str, ":")
		if cn = len(prts) - 1; cn > 2 {
			return 0, fmt.Errorf("invalid format %s", org)
		}
		for i, d := len(prts)-2, Dur(unit*60); i >= 0; i, d = i-1, d*60 {
			prt := prts[i]
			n, err := strconv.ParseUint(prt, 10, 64)
			if err != nil || len(prt) > 2 && len(prts)-i < 3 {
				return 0, fmt.Errorf("invalid format %s", org)
			}
			res += Dur(n) * d
		}
		str = prts[len(prts)-1]
	}
	fst, snd, ok := strings.Cut(str, ".")
	n, err := strconv.ParseUint(fst, 10, 64)
	if err != nil || cn > 0 && len(fst) > 2 {
		return 0, fmt.Errorf("invalid format %s", org)
	}
	res += Dur(n) * Dur(unit)
	if ok {
		f, err := strconv.ParseUint(snd, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid format %s", org)
		}
		for n := len(snd); n > 0; n-- {
			unit = unit / 10
		}
		res += Dur(f) * Dur(unit)
	}
	if neg {
		res = -res
	}
	return res, nil
}


// Val returns a time duration for d.
func (d Dur) Val() time.Duration { return time.Duration(d) }

// String returns d in the default format.
func (d Dur) String() string { return formatDur(d, true) }

// MarshalText returns d in seconds format as bytes.
func (d Dur) MarshalText() ([]byte, error) { return []byte(formatDur(d, false)), nil }

// UnmarshalText parses the duration from b into d or returns an error.
func (d *Dur) UnmarshalText(b []byte) (err error) {
	*d, err = ParseDur(string(b))
	return err
}

func (d Dur) Sync(r Rate) Dur {
	return r.Dur(r.Beats(d+S/Dur(3*r.Num))) 
}

// FormatDur formats a time duration in an accepted ffmpeg format.
// It uses the HH:MM:SS format for durations above one minute and S+ format below.
func FormatDur(d time.Duration) string { return formatDur(Dur(d), true) }

// FormatSec formats a time duration in the ffmpeg S+ format.
func FormatSec(d time.Duration) string { return formatDur(Dur(d), false) }

func formatDur(d Dur, fmtMin bool) string {
	var b strings.Builder
	var u, s, r Dur
	if d < 0 {
		d = -d
		b.WriteByte('-')
	}
	for u = Dur(time.Second); u > 0; u /= 1000 {
		r = d % u
		s, d = (d-r)/u, r
		if u < Dur(time.Second) {
			fmt.Fprintf(&b, "%03d", s)
		} else {
			if s < 60 || !fmtMin {
				fmt.Fprintf(&b, "%d", s)
			} else {
				m := s / 60
				s = s - (m * 60)
				h := m / 60
				m = m - (h * 60)
				if h > 0 {
					fmt.Fprintf(&b, "%d:%02d:%02d", h, m, s)
				} else {
					fmt.Fprintf(&b, "%d:%02d", m, s)
				}
			}
			if d > 0 {
				b.WriteByte('.')
			}
		}
		if d == 0 {
			break
		}
	}
	return b.String()
}

func durUnit(str string) (time.Duration, int) {
	if ln := len(str) - 1; ln > 0 && str[ln] == 's' {
		if ln-1 > 0 {
			if c := str[ln-1]; c == 'm' {
				return time.Millisecond, 2
			} else if c == 'u' {
				return time.Microsecond, 2
			}
		}
		return time.Second, 1
	}
	return time.Second, 0
}
