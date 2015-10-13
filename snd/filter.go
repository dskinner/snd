package snd

import "math"

type Filter struct {
	*mono

	freq  float64
	bw    float64
	delay [2]float64

	a, b0, b1 float64
}

func NewFilter(freq, bw float64, in Sound) *Filter {
	r := 1 - math.Pi*bw/in.SampleRate()
	rr := 2 * r
	rsq := r * r
	cos := (rr / (1 + rsq)) * math.Cos(math.Pi*(freq/in.SampleRate()/2))

	return &Filter{
		mono: newmono(in),
		freq: freq,
		bw:   bw,
		a:    (1 - rsq) * math.Sin(math.Acos(cos)),
		b0:   rr * cos,
		b1:   rsq,
	}
}

func (f *Filter) Prepare(tc uint64) {
	if f.in != nil {
		f.in.Prepare(tc)
	}
	for i, x := range f.in.Samples() {
		f.out[i] = f.a*x + f.b0*f.delay[0] - f.b1*f.delay[1]
		f.delay[1] = f.delay[0]
		f.delay[0] = f.out[i]
	}
}

// type LowPass struct {
// *monosnd
// delay float64
// a, b  float64
// }

// @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
// READ IT NOW
// https://en.wikipedia.org/wiki/Discrete_cosine_transform
// Informative:
// https://en.wikipedia.org/wiki/JPEG#Discrete_cosine_transform
// Investigate:
// https://en.wikipedia.org/wiki/Window_function#Cosine_window
// Seems important:
// http://www.phys.nsu.ru/cherk/fft.pdf
// is interesting:
// https://github.com/FFTW/fftw3
// On Testing:
// http://dsp.stackexchange.com/questions/633/what-data-should-i-use-to-test-an-fft-implementation-and-what-accuracy-should-i
// @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@

// func NewLowPass(freq float64, in Sound) *LowPass {
// gaussian filter https://en.wikipedia.org/wiki/Gaussian_filter

// Formulas expressed in terms of f_s and/or T are readily converted
// to normalized frequency by setting those parameters to 1.
//
// constants T and f_s are then omitted from mathematical expressions.

// angular frequency
// w := twopi * freq
// normalized frequency
// norm := w / in.SampleRate()

// c := 2 - math.Cos(norm)
// b := math.Sqrt(c*c-1) - c
// a := 1 + b
// return &LowPass{newMonosnd(in), 0, a, b}
// }

// func (lp *LowPass) Prepare() {
// lp.snd.Prepare()
// for i := range lp.out {
// if lp.enabled {
// lp.delay = lp.a*lp.in.Output()[i] - lp.b*lp.delay
// lp.out[i] = lp.delay
// } else {
// lp.out[i] = 0
// }
// }
// }

// LowPass is a 3rd order IIR.
//
// Recursive implementation of the Gaussian filter.
//
// Very informative
// https://courses.cs.washington.edu/courses/cse466/13au/pdfs/lectures/Intro%20to%20DSP.pdf
type LowPass struct {
	*mono

	// normalization factor
	b float64
	// coefficients
	b0, b1, b2, b3 float64
	// delays
	d1, d2, d3 float64
}

func NewLowPass(freq float64, in Sound) *LowPass {
	q := 5.0
	s := in.SampleRate() / freq / q

	if s > 2.5 {
		q = 0.98711*s - 0.96330
	} else {
		q = 3.97156 - 4.14554*math.Sqrt(1-0.26891*s)
	}

	q2 := q * q
	q3 := q * q * q

	// redefined from paper to (1 / b0) to save an op div during prepare.
	b0 := 1 / (1.57825 + 2.44413*q + 1.4281*q2 + 0.422205*q3)
	b1 := 2.44413*q + 2.85619*q2 + 1.26661*q3
	b2 := -(1.4281*q2 + 1.26661*q3)
	b3 := 0.422205 * q3
	b := 1 - ((b1 + b2 + b3) * b0)

	b1 *= b0
	b2 *= b0
	b3 *= b0

	return &LowPass{mono: newmono(in), b: b, b0: b0, b1: b1, b2: b2, b3: b3}
}

func (lp *LowPass) Prepare(tc uint64) {
	if lp.in != nil {
		lp.in.Prepare(tc)
	}
	for i, x := range lp.in.Samples() {
		lp.out[i] = lp.b*x + lp.b1*lp.d1 + lp.b2*lp.d2 + lp.b3*lp.d3
		lp.d3, lp.d2, lp.d1 = lp.d2, lp.d1, lp.out[i]
	}
}
