package ffm

import (
	"fmt"
	"strings"
)

// Concat concatenates video and audio streams and creates the combined result at output.
// If you want to combine both for a list of video files, pass the same list audio files as well.
//
// The following is approx the result of an example with multiple video and one audio file:
//
//	ffmpeg -v fail -filter_complex ' \
//	movie=<video1.mp4>:seek_point=<voff>, fps=<fps>, scale=<scale> [v1]; \
//	movie=<videoN.mp4>, fps=<fps>, scale=<scale> [vN]; \
//	[v1] [v2] [vN] concat=n=N:v=1:a=0 [outv]; \
//	amovie=<audio.flac>:seek_point=<aoff> [outa];' \
//	-map [outv] -map [outa]
//	-c:v h264 -g 18 -bf 2 -c:a aac
//	-t <dur> <output.mp4>
func (o *Opts) Concat(output string, videos, audios []string) error {
	tail := Args(output)
	if o.Dur != 0 {
		tail = Args("-t", o.Dur.String(), output)
	}
	var b strings.Builder
	var maps []string
	if o.filterOutV(&b, videos...) {
		maps = append(maps, "-map", "[outv]")
		maps = append(maps, o.VCodec...)
	}
	if o.filterOutA(&b, audios...) {
		maps = append(maps, "-map", "[outa]")
		maps = append(maps, o.ACodec...)
	}
	_, err := o.Cmd("ffprobe", DefLog,
		Args("-filter_complex", b.String()),
		maps, tail,
	).Output()
	return err
}

func (o *Opts) filterOutA(w *strings.Builder, paths ...string) bool {
	var seek string
	if o.Aod != 0 {
		seek = fmt.Sprintf(":seek_point=%s", o.Aod)
	}
	switch len(paths) {
	case 0:
		return false
	case 1:
		fmt.Fprintf(w, "amovie=%s%s [outa]", paths[0], seek)
	default:
		for i, path := range paths {
			fmt.Fprintf(w, "amovie=%s%s [a%d];\n", path, seek, i+1)
			seek = ""
		}
		for i := range paths {
			fmt.Fprintf(w, "[a%d] ", i+1)
		}
		fmt.Fprintf(w, "concat=n=%d:v=0:a=1 [outa];\n", len(paths))
	}
	return true
}

func (o *Opts) filterOutV(w *strings.Builder, paths ...string) bool {
	var seek, filters string
	if o.Vod != 0 {
		seek = fmt.Sprintf(":seek_point=%s", o.Vod)
	}
	if o.Fps.Den != 0 {
		filters += fmt.Sprintf(", fps=%s", o.Fps)
	}
	if o.Dim.H != 0 {
		filters += fmt.Sprintf(", scale=%s", o.Dim)
	}
	switch len(paths) {
	case 0:
		return false
	case 1:
		fmt.Fprintf(w, "movie=%s%s%s [outv];\n", paths[0], seek, filters)
	default:
		for i, path := range paths {
			fmt.Fprintf(w, "movie=%s%s%s [v%d];\n", path, seek, filters, i+1)
			seek = ""
		}
		for i := range paths {
			fmt.Fprintf(w, "[v%d] ", i+1)
		}
		fmt.Fprintf(w, "concat=n=%d:v=1:a=0 [outv];\n", len(paths))
	}
	return true
}
