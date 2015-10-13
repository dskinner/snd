package snd

import "testing"

func BenchmarkOsc(b *testing.B) {
	osc := Osc(Sine(), 440, nil)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		osc.Prepare(uint64(n))
	}
}

func BenchmarkOscMod(b *testing.B) {
	osc := Osc(Sine(), 440, Osc(Sine(), 2, nil))
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		osc.Prepare(uint64(n))
	}
}

func BenchmarkOscAmp(b *testing.B) {
	osc := Osc(Sine(), 440, nil)
	osc.SetAmp(1, Osc(Sine(), 2, nil))
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		osc.Prepare(uint64(n))
	}
}

func BenchmarkOscPhase(b *testing.B) {
	osc := Osc(Sine(), 440, nil)
	osc.SetPhase(1, Osc(Sine(), 2, nil))
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		osc.Prepare(uint64(n))
	}
}

func BenchmarkOscAll(b *testing.B) {
	osc := Osc(Sine(), 440, Osc(Sine(), 2, nil))
	osc.SetAmp(1, Osc(Sine(), 2, nil))
	osc.SetPhase(1, Osc(Sine(), 2, nil))
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		osc.Prepare(uint64(n))
	}
}

func BenchmarkOscReuse(b *testing.B) {
	mod := Osc(Sine(), 2, nil)
	osc := Osc(Sine(), 440, mod)
	osc.SetAmp(1, mod)
	osc.SetPhase(1, mod)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		osc.Prepare(uint64(n))
	}
}
