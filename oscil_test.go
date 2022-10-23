package snd

import (
	"testing"

	"dasa.cc/signal"
)

func BenchmarkOscil(b *testing.B) {
	osc := NewOscil(signal.Sine(), 440, nil)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		osc.Prepare(uint64(n))
	}
}

func BenchmarkOscilMod(b *testing.B) {
	osc := NewOscil(signal.Sine(), 440, NewOscil(signal.Sine(), 2, nil))
	inps := GetInputs(osc)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, inp := range inps {
			inp.sd.Prepare(uint64(n))
		}
	}
}

func BenchmarkOscilAmp(b *testing.B) {
	osc := NewOscil(signal.Sine(), 440, nil)
	osc.SetAmp(1, NewOscil(signal.Sine(), 2, nil))
	inps := GetInputs(osc)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, inp := range inps {
			inp.sd.Prepare(uint64(n))
		}
	}
}

func BenchmarkOscilPhase(b *testing.B) {
	osc := NewOscil(signal.Sine(), 440, nil)
	osc.SetPhase(NewOscil(signal.Sine(), 2, nil))
	inps := GetInputs(osc)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, inp := range inps {
			inp.sd.Prepare(uint64(n))
		}
	}
}

func BenchmarkOscilAll(b *testing.B) {
	osc := NewOscil(signal.Sine(), 440, NewOscil(signal.Sine(), 2, nil))
	osc.SetAmp(1, NewOscil(signal.Sine(), 2, nil))
	osc.SetPhase(NewOscil(signal.Sine(), 2, nil))
	inps := GetInputs(osc)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, inp := range inps {
			inp.sd.Prepare(uint64(n))
		}
	}
}

func BenchmarkOscilReuse(b *testing.B) {
	mod := NewOscil(signal.Sine(), 2, nil)
	osc := NewOscil(signal.Sine(), 440, mod)
	osc.SetAmp(1, mod)
	osc.SetPhase(mod)
	inps := GetInputs(osc)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, inp := range inps {
			inp.sd.Prepare(uint64(n))
		}
	}
}
