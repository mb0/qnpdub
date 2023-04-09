package ffm

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/mb0/qnpdub/av"
)

// Probe runs probe with default options and returns the parsed output or an error.
func Probe(path string) (*Info, error) { return Def().Probe(path) }

// Probe runs probe all with default options and returns the results or the first error.
func ProbeAll(paths ...string) ([]*Info, error) { return Def().ProbeAll(paths...) }

// Probe runs ffprobe and returns the parsed output or an error.
func (o *Opts) Probe(path string) (*Info, error) {
	out, err := o.Cmd("ffprobe", o.Global, DefProbe, Args(path)).Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe %q: %w", path, err)
	}
	res := Info{Path: path}
	err = json.Unmarshal(out, &res)
	if err != nil {
		return nil, fmt.Errorf("ffprobe %q: %w", path, err)
	}
	return &res, nil
}

// Probe returns probe results of all paths or the first error.
func (o *Opts) ProbeAll(paths ...string) ([]*Info, error) {
	res := make([]*Info, 0, len(paths))
	for _, path := range paths {
		nfo, err := o.Probe(path)
		if err != nil {
			return res, err
		}
		res = append(res, nfo)
	}
	return res, nil
}

// Info contains the probe result.
// See the various obj getter methods for known field names.
type Info struct {
	Path    string
	Format  Obj  `json:"format"`
	Streams Objs `json:"streams"`
}

// File returns the filename without the directory or an empty string.
func (nfo *Info) File() string { _, n := filepath.Split(nfo.Path); return n }
func (nfo *Info) Dir() string  { d, _ := filepath.Split(nfo.Path); return d }

// Video returns the first video stream obj or nil.
func (nfo *Info) Video() Obj { return nfo.Streams.First("codec_type", "video") }

// Audio returns the first audio stream obj or nil.
func (nfo *Info) Audio() Obj { return nfo.Streams.First("codec_type", "audio") }

func Paths(nfos []*Info) []string {
	res := make([]string, 0, len(nfos))
	for _, nfo := range nfos {
		res = append(res, nfo.Path)
	}
	return res
}

// Obj wraps an any map with convenience methods to access data.
// It is used for all but the root json object of probe results.
type Obj map[string]any

// Objs wraps an obj slice with convenience methods to access data.
// It is used for streams and sidedata
type Objs []Obj

// Str returns a string value of field with key or an empty string.
// Known string fields are:
//
//	format: filename, format[_long]_name
//	stream: codec[_long]_name, profile, codec_type, codec_time_base, codec_tag[_string], id
//	audio: sample_fmt, channel_layout
//	video: pix_fmt, chroma_location, field_order, is_avc
func (o Obj) Str(key string) string {
	s, _ := o[key].(string)
	return s
}

// Int returns an integer value of field with key or 0.
// Known integer fields are:
//
//	format: nb_streams, nb_programs, size, bit_rate, probe_score
//	stream: index, start_pts, duration_ts, bit_rate, nb_frames, extradata_size
//	audio: sample_rate, channels, bits_per_sample
//	video: [coded_]width, [coded_]height, closed_captions, film_grain,
//	       has_b_frames, level, nal_length_size, refs, bits_per_raw_sample
func (o Obj) Int(key string) (n int64) {
	switch v := o[key].(type) {
	case float64:
		n = int64(v)
	case string:
		n, _ = strconv.ParseInt(v, 10, 64)
	}
	return n
}

// Dur returns a duration value of field with key or 0.
// Known duration fields are start_time, duration in both format and stream.
func (o Obj) Dur(key string) (d av.Dur) {
	if s := o.Str(key); s != "" {
		d, _ = av.ParseDur(s)
	}
	return d
}

// Rate returns a rate value of field with key or a zero rate.
// Known rate fields are r_frame_rate, avg_frame_rate and time_base in streams.
func (o Obj) Rate(key string) (r av.Rate) {
	if s := o.Str(key); s != "" {
		r, _ = av.ParseRate(s)
	}
	return r
}

// Ratio returns a ratio value of field with key or a zero ratio.
// Known ratio fields are sample_aspect_ratio and display_aspect_ratio in video streams.
func (o Obj) Ratio(key string) (r av.Ratio) {
	if s := o.Str(key); s != "" {
		r, _ = av.ParseRatio(s)
	}
	return r
}

// Obj returns an Obj value of field with key or a nil obj.
// Known obj fields besides format are tags in format and streams, and disposition in streams.
func (o Obj) Obj(key string) Obj {
	m, _ := o[key].(map[string]any)
	return Obj(m)
}

// Objs returns an Objs list of field with key or a nil list.
// Known obj list fields besides streams is side_data_list in video streams.
func (o Obj) Objs(key string) Objs {
	vs, _ := o[key].([]any)
	os := make(Objs, 0, len(vs))
	for _, v := range vs {
		m, _ := v.(map[string]any)
		os = append(os, Obj(m))
	}
	return os
}

// First returns the first obj that has val at key or nil.
func (os Objs) First(key string, val any) Obj {
	for _, o := range os {
		if o[key] == val {
			return o
		}
	}
	return nil
}

// All returns a list of all objs that have val at key or a nil list.
func (os Objs) All(key string, val any) (res Objs) {
	for _, o := range os {
		if o[key] == val {
			res = append(res, o)
		}
	}
	return res
}

// Ints collects and returns a list of all integers and the min and max value at key.
func (os Objs) Ints(key string) (res []int64, min, max int64) {
	for i, o := range os {
		n := o.Int(key)
		res = append(res, n)
		if i == 0 {
			min, max = n, n
		} else if n < min {
			min = n
		} else if n > max {
			max = n
		}
	}
	return res, min, max
}

// SumInt returns the total of all integers and at key.
func (os Objs) SumInt(key string) (sum int64) {
	for _, o := range os {
		sum += o.Int(key)
	}
	return sum
}

// SumDur returns the total of all durations and at key.
func (os Objs) SumDur(key string) (sum av.Dur) {
	for _, o := range os {
		sum += o.Dur(key)
	}
	return sum
}
