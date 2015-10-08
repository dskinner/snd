package snd

import "testing"

func TestDecibel(t *testing.T) {
	tests := []struct {
		db  Decibel
		amp float64
	}{
		{0, 1},
		{1, 1.1220},
		{3, 1.4125},
		{6, 1.9952},
		{10, 3.1622},
	}

	for _, test := range tests {
		if !equals(test.db.Amp(), test.amp) {
			t.Errorf("%s have %v, want %v", test.db, test.db.Amp(), test.amp)
		}
	}
}
