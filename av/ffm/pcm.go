package ffm

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/mb0/qnpdub/av/pcm"
)

// GenPCMCmd returns a command to generate a waveform for a audio or video file.
func GenPCMCmd(path, dest string, f pcm.Format) *exec.Cmd {
	return Def().Cmd("ffmpeg", Args(
		"-i", path, // path to audio or video file
		"-ac", "1", // set the number of audio channels
		"-af", fmt.Sprintf("aresample=%d", f.Rate.Num),
		"-map", "0:a", // select only the audio channel
		"-c:a", f.PCM.String(), // convert audio to pcm format
		"-f", "data", dest, // output as data to dest
	))
}

// GenPCMInto generates waveform for the media file at path into the given writer.
func GenPCMInto(path string, into io.Writer, f pcm.Format) error {
	cmd := GenPCMCmd(path, "-", f) // render data to stdout
	cmd.Stdout = into
	return cmd.Run()
}
