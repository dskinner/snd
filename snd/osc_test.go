package snd

import "testing"

func BenchmarkOsc(b *testing.B) {
	osc := Osc(Sine(), 440, nil)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		osc.Prepare()
	}
}

func BenchmarkOscMod(b *testing.B) {
	osc := Osc(Sine(), 440, Osc(Sine(), 2, nil))
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		osc.Prepare()
	}
}
