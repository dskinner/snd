package snd

import (
	"testing"
	"time"
)

type unit struct{ *mono }

func newunit() *unit {
	u := &unit{newmono(nil)}
	for i := range u.out {
		u.out[i] = DefaultAmpMult
	}
	return u
}

func (u *unit) Prepare(tc uint64) bool { return true }

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

func BenchmarkPiano(b *testing.B) {
	mix := NewMixer()
	for i := 0; i < 12; i++ {
		oscil := Osc(Sawtooth(4), 440, Osc(Sine(), 2, nil))
		oscil.SetPhase(1, Osc(Square(4), 200, nil))
		comb := NewComb(0.8, 10*time.Millisecond, oscil)
		adsr := NewADSR(50*time.Millisecond, 500*time.Millisecond, 100*time.Millisecond, 350*time.Millisecond, 0.4, 1, comb)
		instr := NewInstrument(adsr)
		mix.Append(instr)
	}
	loop := Loop(5*time.Second, mix)
	mixloop := NewMixer(mix, loop)
	lp := NewLowPass(1500, mixloop)
	// mixwf, err := NewWaveform(nil, 4, lp)
	// if err != nil {
	// b.Fatal(err)
	// }
	// pan := NewPan(0, mixwf)
	pan := NewPan(0, lp)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		pan.Prepare(uint64(n + 1))
	}
}
