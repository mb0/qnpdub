package av

import (
	"fmt"
	"strconv"
	"strings"
)

type Rate struct {
	Num, Den int
}

func Hz(n int) Rate { return Rate{Num: n, Den: 1} }

func ParseRate(str string) (r Rate, err error) {
	r.Num, r.Den, err = parsePair(str, "/")
	if r.Den == 0 && err == nil && r.Num != 0 {
		r.Den = 1
	}
	return r, err
}

func (r Rate) Dur(n int) Dur {
	if r.Num == 0 || r.Den == 0 {
		return 0
	}
	return Dur(n) * Dur(r.Den) * S / Dur(r.Num)
}
func (r Rate) Beats(d Dur) int {
	if r.Num == 0 || r.Den == 0 {
		return 0
	}
	return int(d+1) * r.Num / (r.Den * int(S))
}
func (r Rate) FrameStr(d Dur) string {
	n := r.Beats(d) + 1
	s := (r.Dur(n) / S) * S
	rest := n - r.Beats(s)
	return fmt.Sprintf("%s+%d", s, rest)
}

func (r Rate) Zero() bool                   { return r.Num == 0 && r.Den == 0 }
func (r Rate) String() string               { return fmt.Sprintf("%d/%d", r.Num, r.Den) }
func (r Rate) MarshalText() ([]byte, error) { return []byte(r.String()), nil }
func (r *Rate) UnmarshalText(b []byte) (err error) {
	*r, err = ParseRate(string(b))
	return err
}

type Ratio struct {
	W, H int
}

func ParseRatio(str string) (r Ratio, err error) {
	r.W, r.H, err = parsePair(str, ":")
	return r, err
}

func (r Ratio) Zero() bool                   { return r.W == 0 && r.H == 0 }
func (r Ratio) String() string               { return fmt.Sprintf("%d:%d", r.W, r.H) }
func (r Ratio) MarshalText() ([]byte, error) { return []byte(r.String()), nil }
func (r *Ratio) UnmarshalText(b []byte) (err error) {
	*r, err = ParseRatio(string(b))
	return err
}

func parsePair(str, sep string) (a, b int, err error) {
	fst, snd, ok := strings.Cut(str, sep)
	var aa, bb int64
	aa, err = strconv.ParseInt(fst, 10, 64)
	if err == nil && ok {
		bb, err = strconv.ParseInt(snd, 10, 64)
	}
	return int(aa), int(bb), err
}
