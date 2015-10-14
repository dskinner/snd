package snd

import (
	"sort"
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
	mixwf, err := NewWaveform(nil, 4, lp)
	if err != nil {
		b.Fatal(err)
	}
	pan := NewPan(0, mixwf)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		pan.Prepare(uint64(n + 1))
	}
}

func TestWalk(t *testing.T) {
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
	mixwf, err := NewWaveform(nil, 4, lp)
	if err != nil {
		t.Fatal(err)
	}
	pan := NewPan(0, mixwf)

	//
	var ins []*inp
	getins(pan, 0, &ins)
	t.Log("inputs length", len(ins))
	sort.Sort(bywt(ins))
	for i, v := range ins {
		t.Log(i, v)
	}
}

func BenchmarkWalk(b *testing.B) {
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
	mixwf, err := NewWaveform(nil, 4, lp)
	if err != nil {
		b.Fatal(err)
	}
	pan := NewPan(0, mixwf)

	//
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var ins []*inp
		getins(pan, 0, &ins)
		sort.Sort(bywt(ins))
	}
}

type inp struct {
	sd Sound
	wt int
}

// TODO janky methods
type bywt []*inp

func (a bywt) Len() int           { return len(a) }
func (a bywt) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a bywt) Less(i, j int) bool { return a[i].wt > a[j].wt }

// func (a bywt) multi() [][]Sound {
// if len(a) == 0 {
// return nil
// }

// n := a[0].wt
// out := make([][]Sound, n)
// for i, p := range a {

// }
// }

func getins(sd Sound, wt int, out *[]*inp) {
	for _, in := range sd.Inputs() {
		if in == nil {
			continue
		}

		at := -1
		for i, p := range *out {
			if p.sd == in {
				if p.wt >= wt {
					return // object has or will be traversed on different branch
				}
				at = i
				break
			}
		}
		if at != -1 {
			(*out)[at].sd = in
			(*out)[at].wt = wt
		} else {
			*out = append(*out, &inp{in, wt})
		}
		getins(in, wt+1, out)
	}
}
