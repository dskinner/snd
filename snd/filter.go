package snd

import "math"

type Filter struct {
	*snd

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
		snd:  newSnd(in),
		freq: freq,
		bw:   bw,
		a:    (1 - rsq) * math.Sin(math.Acos(cos)),
		b0:   rr * cos,
		b1:   rsq,
	}
}

func (f *Filter) Prepare() {
	f.snd.Prepare()
	for i := range f.out {
		f.out[i] = f.a*f.in.Output()[i] + f.b0*f.delay[0] - f.b1*f.delay[1]
		f.delay[1] = f.delay[0]
		f.delay[0] = f.out[i]
	}
}
