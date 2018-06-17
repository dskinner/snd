package snd

import (
	"fmt"
	"math"
	"testing"
)

func TestComplex(t *testing.T) {
	phase := 0.1
	amp := 0.5
	a := complex(phase, amp)
	t.Log(a)
}

func TestRadian(t *testing.T) {

	var rad Radian
	const n = 10
	onethird := Radian(math.MaxUint64 / 3.)
	for i := 1; i < n; i++ {
		rad += onethird + onethird + onethird

		fmt.Println(rad.Degrees())
		fmt.Println(math.MaxUint64 - rad)
		fmt.Println()

		// if rad.Degrees() != 360 {
		// t.Fatalf("failed on iteration %v", i)
		// }
	}

	// var rad Radian2

	// fmt.Println(RPiTwo % 3)
	// x := float64(RPiTwo) / 3.0

	// onethird := Radian2(x)
	// fmt.Println(onethird)
	// for i := 0; i < 10; i++ {
	// rad += onethird
	// fmt.Println(rad.Degrees())
	// }

	// var z Radian2

	// const n = 44100
	// for i := 1; i < n; i++ {
	// z += RPiTwo
	// if uint64(z)%uint64(RPiTwo) != 0 {
	// t.Fatalf("%v: mod failed for %v", i, z)
	// }
	// if uint64(i) != z.Hertz() {
	// t.Fatalf("%v: hertz failed for %v", i, z)
	// }
	// t.Log(z.Degrees(), z.Hertz(), uint64(z)%uint64(RPiTwo))
	// }

	// z += 1
	// t.Log(z.Degrees(), z.Hertz(), uint64(z)%uint64(RPiTwo))

	// x := uint64(math.MaxUint64)
	// for ; x%uint64(RPiTwo) != 0; x-- {
	// }
	// t.Log(x % uint64(RPiTwo))
	// t.Logf("0x%016X", x)

	// THz GHz MHz kHz  Hz
	// 281,474,976,710,655
	// t.Log(Radian2(MaxHertz).Hertz())

	// 281,474,976,645,120
	// t.Log(Radian2(MaxHertz2).Hertz())
	// 4,294,967,295
	// t.Log(MaxHertz2 / 0x100000000)
	// t.Log(MaxHertz3 / 0x1000000000000)

	//           mHz μHz nHz pHz fHz
	// 000000000.000,000,000,232,830,64365386963e-10
	// t.Log(1. / (1. + math.MaxUint32))

	// t.Log(RPiTwo)

	// t.Log(2. / 65536)

	// 0.000,015,258,789,0625

	// t.Log(RPi.Degrees())
	// t.Log(math.MaxUint64 / RPi)

	// x := 2 * RPi
	// t.Logf("0x%016X", x)

	// 562,949,953,421,311

	// a := 750 * Millihertz
	// r := a.Angular()
	// t.Log(r.Degrees() / 360.0 / float64(HHertz))

	// r2 := Radian(a)
	// t.Log(r2)
	// t.Log(r2.Degrees())
}

func TestIndex(t *testing.T) {
	const n = 256
	var f float64 = 256. / 44100.
	var freq float64 = 440.

	h := make(Discrete, n)
	Sample(h, SineFunc, 1./n, 0)

	out := make(Discrete, n)
	Sample(out, h.Index, 1./(f*freq), 0)

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
	a := make(Discrete, 256)
	Sample(a, SineFunc, 1./256, 0)

	b := make(Discrete, 256)
	Sample(b, a.Index, 1./256, 0)

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
	Sample(a, SineFunc, 1./sr, 0)

	b := make(Discrete, sr)
	Sample(b, a.Index, 2./sr, 0)

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
