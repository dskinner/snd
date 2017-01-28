package snd

import "testing"

func TestSample(t *testing.T) {
	a := make(Discrete, 256)
	Continuous(SineFunc).Sample(a, 256)

	b := make(Discrete, 256)
	a.Sample(b, 256)

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
	Continuous(SineFunc).Sample(a, sr)

	b := make(Discrete, sr)
	a.Sample(b, sr/2)

	for i := 0; i < sr; i++ {
		if a[i] != b[i] {
			t.Fail()
		}
		t.Logf("%v: a[%.4f] b[%.4f]\n", i, a[i], b[i])
	}
}

var resf float64

func BenchmarkDiscrete(b *testing.B) {
	sig := Sine()
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i := 0; i < 256; i++ {
			resf = sig[i] // TODO benchmark doesn't make sense anymore
		}
	}
}

type foo [256]float64

func BenchmarkDiscreteCopy(b *testing.B) {
	sig := Sine()
	dup := make(Discrete, len(sig))
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		copy(dup, sig)
	}
}

func BenchmarkInterpolate(b *testing.B) {
	sig := Sine()
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		resf = sig.Interpolate(float64(n) / 1024)
	}
}
