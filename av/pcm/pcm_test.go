package pcm

import "testing"

func TestPCMString(t *testing.T) {
	tests := []struct {
		PCM
		want string
	}{
		{S8, "pcm_s8"},
		{S16BE, "pcm_s16be"},
	}
	for _, test := range tests {
		got := test.PCM.String()
		if got != test.want {
			t.Errorf("got %s want %s", got, test.want)
		}
	}
}
