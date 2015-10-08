package snd

import "testing"

// TODO test harms against each other
// https://en.wikipedia.org/wiki/Pulse_wave
// "A pulse wave can be created by subtracting a sawtooth wave from a phase-shifted version of itself"
// this may be way to at least guarantee some kind of consistency in the library and the HarmFuncs
// though this wouldn't be a "complete" test.
func TestHarm(t *testing.T) {
	for _, fn := range []HarmFunc{SineFunc, SquareFunc, SawtoothFunc, PulseFunc} {
		var h Harm
		h.Eval(1, 0, fn)
		for i, e := range h {
			if e < -1 || e > 1 {
				t.Errorf("%#v not normalized at index %v", fn, i)
			}
		}
	}
}

func BenchmarkSine(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = Sine()
	}
}

func BenchmarkSquare(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = Square(1)
	}
}

func BenchmarkSawtooth(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = Sawtooth(1)
	}
}

func BenchmarkPulse(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = Pulse(1)
	}
}
