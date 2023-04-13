// Package ffm provides helpers to run ffmpeg commands.
package ffm

import (
	"flag"
	"log"
	"os/exec"
	"sync"

	"github.com/mb0/qnpdub/av"
)

var (
	DefLog    = []string{"-v", "error"}
	DefVCodec = []string{"-c:v", "h264", "-g", "18", "-bf", "2"}
	DefACodec = []string{"-c:a", "aac"}
	DefProbe  = []string{"-print_format", "json", "-show_format", "-show_streams"}
)

// Def returns the common default options for operations.
func Def() *Opts {
	return &Opts{
		Global: DefLog,
		VCodec: DefVCodec,
		ACodec: DefACodec,
	}
}

// Opts contains common options used for ffmpeg operations.
type Opts struct {
	Global []string
	VCodec []string
	ACodec []string
	Dim    av.Ratio
	Fps    av.Rate
	Vod    av.Dur
	Aod    av.Dur
	Dur    av.Dur
	Rot    int
	Yes    bool
}

func (o *Opts) Flags() *flag.FlagSet {
	fs := flag.NewFlagSet("media", flag.ContinueOnError)
	fs.TextVar(&o.Vod, "vod", o.Vod, "first video offset duration")
	fs.TextVar(&o.Aod, "aod", o.Aod, "first audio offset duration")
	fs.TextVar(&o.Dur, "dur", o.Dur, "limit output duration")
	fs.TextVar(&o.Fps, "fps", o.Fps, "video frame rate")
	fs.TextVar(&o.Dim, "dim", o.Dim, "scale to output dimension")
	fs.IntVar(&o.Rot, "rot", o.Rot, "rotate by degrees")
	fs.BoolVar(&o.Yes, "yes", o.Yes, "override existing files")
	return fs
}

// Cmd returns an exec command with concatenated arguments and looked up path.
func (o *Opts) Cmd(name string, aas ...[]string) *exec.Cmd {
	args := []string{name}
	for _, aa := range aas {
		args = append(args, aa...)
	}
	return &exec.Cmd{Path: mustLook(name), Args: args}
}

// Args is a vanity function to convert a string to a string slice.
func Args(args ...string) []string { return args }

var (
	lock sync.Mutex
	look = map[string]string{}
)

func mustLook(name string) string {
	lock.Lock()
	defer lock.Unlock()
	if path := look[name]; path != "" {
		return path
	}
	path, err := exec.LookPath(name)
	if err != nil {
		log.Fatalf("cmd %s not found: %v", name, err)
	}
	look[name] = path
	return path
}
