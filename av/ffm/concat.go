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
//	     movie=<video1.mp4>:seek_point=<voff>, fps=<fps>, scale=<scale> [v1]; \
//	     movie=<videoN.mp4>, fps=<fps>, scale=<scale> [vN]; \
//	     [v1] [v2] [vN] concat=n=N:v=1:a=0 [outv];' \
//	     -filter_complex ' \
//	     amovie=<audio.flac>:seek_point=<aoff> [outa];' \
//	     -map [outv] -map [outa] \
//	     -c:v h264 -g 18 -bf 2 -c:a aac \
//	     -t <dur> <output.mp4>
func (o *Opts) Concat(output string, videos, audios []string) error {
	var args []string
	args = append(args, o.videoArgs(videos...)...)
	args = append(args, o.audioArgs(audios...)...)
	if o.Yes {
		args = append(args, "-y")
	}
	if o.Dur != 0 {
		args = append(args, "-t", o.Dur.String())
	}
	cmd := o.Cmd("ffmpeg", DefLog, args, Args(output))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("concat err: %v\n%v\n%s", err, cmd.Args, out)
	}
	return nil
}

func (o *Opts) videoArgs(paths ...string) []string {
	if len(paths) == 0 {
		return nil
	}
	res := Args("-filter_complex", "", "-map", "[outv]")
	res = append(res, o.VCodec...)
	var fs strings.Builder
	for i, path := range paths {
		fmt.Fprintf(&fs, "movie=%s", path)
		if i == 0 && o.Vod > 0 {
			fmt.Fprintf(&fs, ", trim=start=%s", o.Vod.Secs())
		}
		fmt.Fprintf(&fs, ", setpts=(PTS-STARTPTS)")
		if o.Fps.Den != 0 {
			fmt.Fprintf(&fs, ", fps=%s", o.Fps)
		}
		if o.Rot != 0 {
			fmt.Fprintf(&fs, ", rotate=PI/%d", 180/o.Rot)
			if (o.Rot/90)%2 == 1 {
				fmt.Fprintf(&fs, ":ow=ih:oh=iw")
			}
		}
		if !o.Dim.Zero() {
			fmt.Fprintf(&fs, ", scale=%s", o.Dim)
		}
		if len(paths) > 1 {
			fmt.Fprintf(&fs, " [v%d];\n", i+1)
		} else {
			fmt.Fprintf(&fs, " [outv]\n")
		}
	}
	if len(paths) > 1 {
		for i := range paths {
			fmt.Fprintf(&fs, "[v%d] ", i+1)
		}
		fmt.Fprintf(&fs, "concat=n=%d:v=1:a=0 [outv]\n", len(paths))
	}
	res[1] = fs.String()
	return res
}

func (o *Opts) audioArgs(paths ...string) []string {
	if len(paths) == 0 {
		return nil
	}
	res := Args("-filter_complex", "", "-map", "[outa]")
	res = append(res, o.ACodec...)
	var fs strings.Builder
	for i, path := range paths {
		fmt.Fprintf(&fs, "amovie=%s", path)
		if i == 0 && o.Aod > 0 {
			fmt.Fprintf(&fs, ", atrim=start=%s", o.Aod.Secs())
		}
		fmt.Fprintf(&fs, ", asetpts=(PTS-STARTPTS)")
		if len(paths) > 1 {
			fmt.Fprintf(&fs, " [a%d];\n", i+1)
		} else {
			fmt.Fprintf(&fs, " [outa]\n")
		}
	}
	if len(paths) > 1 {
		for i := range paths {
			fmt.Fprintf(&fs, "[a%d] ", i+1)
		}
		fmt.Fprintf(&fs, "concat=n=%d:v=0:a=1 [outa]\n", len(paths))
	}
	res[1] = fs.String()
	return res
}
