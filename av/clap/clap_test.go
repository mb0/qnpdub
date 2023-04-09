package clap

import (
	"reflect"
	"strings"
	"testing"

	"github.com/mb0/qnpdub/av"
)

var clapTests = []struct {
	path  string
	dur   string
	clap  string
	off   string
	next2 string
}{
	{"../testdata/example.flac",
		"1:20.382250", "1:18+13", "10:46+12",
		"1:11+7 1:10+18",
	},
	{"../testdata/example.mp4",
		"26.346750", "23+8", "11:41+17",
		"16+11 16+2",
	},
	{"../testdata/full",
		"12:13.383875", "12:04+24", "0+1",
		"11:58+1 11:57+27",
	},
}

func TestDetectorDetect(t *testing.T) {
	d := Default()
	r := av.Hz(30)
	for _, test := range clapTests {
		w, err := d.Load(test.path)
		if err != nil {
			t.Errorf("load %s: %v", test.path, err)
			continue
		}
		res, err := d.Detect(w, 3)
		if err != nil {
			t.Errorf("detect %s: %v", test.path, err)
			continue
		}
		if got := w.Dur(w.Count).String(); got != test.dur {
			t.Errorf("detect %s dur got %s want %s", test.path, got, test.dur)
		}
		if got := r.FrameStr(d.Format.Dur(res[0])); got != test.clap {
			t.Errorf("detect %s clap %s want %s", test.path, got, test.clap)
		}
		if got := frameStrs(d.Format.Rate, r, res[1:]); got != test.next2 {
			t.Errorf("detect %s next2 got %s want %s", test.path, got, test.next2)
		}
	}
}
func TestDetectorMatch(t *testing.T) {
	fr := av.Hz(30)
	var paths, want []string
	for _, test := range clapTests {
		paths = append(paths, test.path)
		want = append(want, test.off)
	}
	d := Default()
	ws, err := d.LoadAll(paths...)
	if err != nil {
		t.Errorf("load: %v", err)
	}
	claps, err := d.Match(fr, ws...)
	if err != nil {
		t.Errorf("sync: %v", err)
	}
	var got []string
	for _, clap := range claps {
		got = append(got, fr.FrameStr(clap.Off.Sync(fr)))
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("sync got %v want %v", got, want)
	}
}
func TestMatch(t *testing.T) {
	tests := []struct {
		a, b   []int
		ao, bo int
	}{
		{[]int{}, []int{}, -1, -1},
		{[]int{4}, []int{}, -1, -1},
		{[]int{4}, []int{4}, 0, 0},
		{[]int{4}, []int{5}, 1, 0},
		{[]int{4, 7}, []int{4}, 0, 0},
		{[]int{4, 7, 15, 17}, []int{4}, 0, 0},
		{[]int{4, 7}, []int{7}, 3, 0},
		{[]int{4}, []int{7, 4}, 3, 0},
		{[]int{4, 7, 15, 17}, []int{8}, 4, 0},
		{[]int{4, 7}, []int{4, 7}, 0, 0},
		{[]int{4, 7}, []int{0, 3}, 0, 4},
		{[]int{4, 7}, []int{3, 7}, 0, 1},
		{[]int{3, 7}, []int{4, 6}, 1, 0},
		{[]int{4, 7, 15}, []int{3, 8, 11}, 0, 4},
		{[]int{4, 7, 15, 20}, []int{4, 8, 16, 21}, 1, 0},
	}
	for _, test := range tests {
		m := match(DistWeb(test.a), DistWeb(test.b))
		ao, bo := m.ao, m.bo
		if m.rev {
			ao, bo = m.bo, m.ao
		}
		if ao != test.ao {
			t.Errorf("%v %v got a off %d want %d", test.a, test.b, ao, test.ao)
		}
		if bo != test.bo {
			t.Errorf("%v %v got b off %d want %d", test.a, test.b, bo, test.bo)
		}
	}
}

func TestDistWeb(t *testing.T) {
	tests := []struct {
		offs []int
		max  int
		want []int
		rows []int
	}{
		{[]int{}, 0, []int{}, []int{}},
		{[]int{4}, 0, []int{}, []int{}},
		{[]int{4, 7}, 3, []int{3}, []int{3, 3}},
		{[]int{4, 7, 15}, 11, []int{3, 11, 8}, []int{
			/*0*/ 3, 11,
			3 /*0*/, 8,
			11, 8, /*0*/
		}},
		{[]int{4, 7, 15, 17}, 13, []int{3, 11, 13, 8, 10, 2}, []int{
			/* 0 */ 3, 11, 13,
			3 /* 0 */, 8, 10,
			11, 8 /* 0 */, 2,
			13, 10, 2, /* 0 */
		}},
		{[]int{4, 7, 15, 17, 25}, 21, []int{3, 11, 13, 21, 8, 10, 18, 2, 10, 8}, []int{
			/* 0 */ 3, 11, 13, 21,
			3 /* 0 */, 8, 10, 18,
			11, 8 /* 0 */, 2, 10,
			13, 10, 2 /* 0 */, 8,
			21, 18, 10, 8, /* 0 */
		}},
	}
	for _, test := range tests {
		web := DistWeb(test.offs)
		if !reflect.DeepEqual(web.Dist, test.want) {
			t.Errorf("for %v got web %v want %v", test.offs, web.Dist, test.want)
		}
		if got := web.MonoMax(); got != test.max {
			t.Errorf("for %v got max %d want %d", test.offs, got, test.max)
		}
		rows := make([]int, 0, len(web.Vals)*(len(web.Vals)-1))
		for i := range web.Vals {
			rows = append(rows, web.Row(i)...)
		}
		if !reflect.DeepEqual(rows, test.rows) {
			t.Errorf("for %v got rows %v want %v", test.offs, rows, test.rows)
		}
	}
}

func frameStrs(srcr, dstr av.Rate, offs []int) string {
	var b strings.Builder
	for i, off := range offs {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(dstr.FrameStr(srcr.Dur(off)))

	}
	return b.String()
}
