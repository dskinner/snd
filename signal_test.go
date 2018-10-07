package snd

import (
	"testing"
)

func TestIndex(t *testing.T) {
	const n = 256
	var f float64 = 256. / 44100.
	var freq float64 = 440.

	h := make(Discrete, n)
	h.Sample(SineFunc, 1./n, 0)

	out := make(Discrete, n)
	out.Sample(h.At, 1./(f*freq), 0)

	sampleFreq := n / (f * freq)

	t.Log(sampleFreq)

	var idx float64
	for i := 0; i < 256; i++ {

		sample := float64(i) / float64(sampleFreq)

		t.Logf("%v: a[%.4f] b[%.4f]\n", i, idx/n, sample)

		// a := int(idx) & (n - 1)

		// u := idx / n
		// u -= float64(int(u))
		// u *= float64(n)
		// b := int(u)

		// a := h.Index(idx / n)
		// b := out[i]

		// if a != b {
		// t.Fail()
		// t.Logf("%v: idx[%.4f] a[%.4f] b[%.4f]\n", i, idx, a, b)
		// }

		idx += freq * f
	}
}

func TestSample(t *testing.T) {
	const n = 256
	a := make(Discrete, n)
	a.Sample(SineFunc, 1/n, 0)

	b := make(Discrete, n)
	b.Sample(a.At, 1/n, 0)

	for i := range a {
		if a[i] != b[i] {
			t.Fail()
		}
	}

	if t.Failed() {
		for i := range a {
			t.Logf("%v: a[%.4f] b[%.4f]\n", i, a[i], b[i])
		}
	}
}

func _TestSampleDown(t *testing.T) {
	const sr = 256
	a := make(Discrete, sr)
	a.Sample(SineFunc, 1./sr, 0)

	b := make(Discrete, sr)
	b.Sample(a.At, 2./sr, 0)

	for i := 0; i < sr; i++ {
		if a[i] != b[i] {
			t.Fail()
		}
		t.Logf("%v: a[%.4f] b[%.4f]\n", i, a[i], b[i])
	}
}

var resf float64

func BenchmarkDiscreteInterp(b *testing.B) {
	sig := Sine()
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		resf = sig.Interp(float64(n) / float64(len(sig)))
	}
}

func BenchmarkDiscreteAt(b *testing.B) {
	sig := Sine()
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		resf = sig.At(float64(n) / float64(len(sig)))
	}
}

func BenchmarkDiscreteIndex(b *testing.B) {
	sig := Sine()
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		resf = sig.Index(n)
	}
}
