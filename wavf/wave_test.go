package wavf

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestGenCmd(t *testing.T) {
	f := Format{PCM: PCM_S8, Rate: 8000}
	tests := []struct {
		path string
	}{
		{"../testdata/example.flac"},
		{"../testdata/example.mp4"},
	}
	for _, test := range tests {
		dest := fmt.Sprintf("%s.%s", test.path, f.String())
		_, err := os.Stat(dest)
		if testing.Short() {
			if err != nil {
				continue
			}
		} else {
			if err == nil {
				_, name := filepath.Split(dest)
				dest = filepath.Join(os.TempDir(), name)
			}
			os.Remove(dest)
			err = GenCmd(test.path, dest, f).Run()
			if err != nil {
				t.Errorf("error gen %s %v", dest, err)
				continue
			}
		}
		h, err := hash(dest)
		if err != nil {
			t.Errorf("error hash %s %v", dest, err)
		}
		gold := test.path + ".gold"
		g, err := ioutil.ReadFile(gold)
		if err != nil {
			t.Errorf("error read %s %v", gold, err)
			err = ioutil.WriteFile(gold, h, 0755)
			if err != nil {
				t.Errorf("error write %s %v", gold, err)
			}
			continue
		}
		want := fmt.Sprintf("%x", h)
		got := fmt.Sprintf("%x", g)
		if want != got {
			t.Errorf("hash %s got %s want %s", test.path, got, want)
		}
	}
}

func hash(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	h := sha256.New()
	_, err = io.Copy(h, f)
	return h.Sum(nil), err

}
