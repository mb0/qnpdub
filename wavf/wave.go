package wavf

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

// GenCmd returns a command to generate a waveform for a audio or video file.
func GenCmd(path, dest string, f Format) *exec.Cmd {
	cmd := exec.Command("ffmpeg",
		"-i", path, // path to audio or video file
		"-ac", "1", // set the number of audio channels
		"-filter:a", fmt.Sprintf("aresample=%d", f.Rate),
		"-map", "0:a", // select only the audio channel
		"-codec:a", f.PCM.String(), // convert audio to pcm format
		"-f", "data", dest, // output as data to dest
	)
	return cmd
}

// GenInto generates waveform for the media file at path into the given writer.
func GenInto(path string, into io.Writer, f Format) error {
	cmd := GenCmd(path, "-", f) // render data to stdout
	cmd.Stdout = into
	return cmd.Run()
}

type Info struct {
	Format
	Path  string
	Count int
}

type W struct {
	Info
	Reader io.ReadSeekCloser
}

func Open(path string, f Format) (*W, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}
	count := int(fi.Size() / int64(f.Bytes))
	return &W{Info{f, path, count}, file}, nil
}

func (w *W) Close() error {
	return w.Reader.Close()
}

type Chunked struct {
	N   int
	chk []int16
	buf []byte
}

func (c *Chunked) Chunks(n int) int {
	if n > 0 {
		return (n + c.N - 1) / c.N
	}
	return 0
}

func (c *Chunked) Read(w *W, n int, h func([]int16)) error {
	switch w.Bytes {
	case 1, 2:
	default:
		return fmt.Errorf("bitsize %d not implemented", w.Bytes*8)
	}
	nn := n * w.Bytes
	// chunk 16<<10 uses 32kb space and uses at 8khz (2.048s) and a
	//   64kb buffer with 16bit samples
	//   32kb buffer for 8bit samples
	if c.N == 0 {
		c.N = 16 << 10
	}
	if len(c.chk) < c.N {
		c.chk = make([]int16, c.N)
	}
	bufn := c.N * w.Bytes
	if len(c.buf) < bufn {
		c.buf = make([]byte, bufn)
	}
	for rest := nn; rest > 0; rest -= bufn {
		l := bufn
		if rest < l {
			l = rest
		}
		_, err := io.ReadFull(w.Reader, c.buf[:l])
		if err != nil {
			return err
		}
		add := PCM.add8
		if w.Bytes == 2 {
			add = PCM.add16
		}
		h(add(w.PCM, c.buf, c.chk[:0]))
	}
	return nil
}
func (c *Chunked) Extract(w *W, off, n int, h func([]int16)) error {
	boff := off * w.Bytes
	_, err := w.Reader.Seek(int64(boff), io.SeekStart)
	if err != nil {
		return fmt.Errorf("seek %d failed: %w", boff, err)
	}
	return c.Read(w, n, h)
}
