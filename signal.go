package snd

import "math"

const twopi = 2 * math.Pi

type Sampler interface {
	// Sample samples len(s) values into s at the sampling frequency f.
	//
	// If sampling a Discrete, set the sampling frequency to desiredSampleRate*(len(source)/sourceSampleRate).
	//
	// If building a lookup table, let f = len(s).
	//
	// TODO should this accept phase arg as well?
	Sample(s Discrete, f Hertz)
}

// Continuous represents a continuous-time signal.
type Continuous func(t float64) float64

// Sample satisfies Sampler interface.
func (fn Continuous) Sample(s Discrete, f Hertz) {
	for i := 0; i < len(s); i++ {
		t := float64(i) / float64(f)
		s[i] = fn(t)
	}
}

// Discrete represents a discrete-time signal.
type Discrete []float64

// Sample satisfies Sampler interface, interpolating values where necessary.
// To sample without interpolation, use Continuous(sig.Index).Sample.
func (sig Discrete) Sample(s Discrete, f Hertz) {
	for i := 0; i < len(s); i++ {
		t := float64(i) / float64(f)
		s[i] = sig.Interpolate(t)
	}
}

// Interpolate uses the fractional component of t to return an interpolated value,
// where t may be considered the sampling period of len(sig).
//
// TODO currently does linear interpolation...
func (sig Discrete) Interpolate(t float64) float64 {
	t -= float64(int(t))   // fractional
	t *= float64(len(sig)) // integer is index, fractional is amount to interpolate to next index

	frac := t - float64(int(t))
	i := int(t - frac)

	if frac == 0 {
		return sig[i]
	}

	j := i + 1
	if j == len(sig) {
		j = 0
	}

	return (1-frac)*sig[i] + frac*sig[j]
}

// Index uses the fractional component of t to find the nearest
// truncated index and returns the value, where t may be considered
// the sampling period of len(sig).
func (sig Discrete) Index(t float64) float64 {
	t -= float64(int(t))
	t *= float64(len(sig))
	return sig[int(t)]
}

// Normalize alters sig so values belong to [-1..1].
func (sig *Discrete) Normalize() {
	var max float64
	for _, x := range *sig {
		a := math.Abs(x)
		if max < a {
			max = a
		}
	}
	for i, x := range *sig {
		(*sig)[i] = x / max
	}
}

// NormalizeRange alters sig so values belong to [s..e].
//
// Calling this method for values that already occupy [s..e] will degrade values
// further due to round-off error.
func (sig *Discrete) NormalizeRange(s, e float64) {
	if s > e {
		s, e = e, s
	}
	n := e - s

	var min, max float64
	for _, x := range *sig {
		if min > x {
			min = x
		}
		if max < x {
			max = x
		}
	}
	r := max - min
	for i, x := range *sig {
		(*sig)[i] = s + n*(x-min)/r
	}
}

// Reverse sig in place so the first element becomes the last and the last element becomes the first.
func (sig *Discrete) Reverse() {
	for l, r := 0, len(*sig)-1; l < r; l, r = l+1, r-1 {
		(*sig)[l], (*sig)[r] = (*sig)[r], (*sig)[l]
	}
}

// UnitInverse sets each sample to 1 - x.
func (sig *Discrete) UnitInverse() {
	for i, x := range *sig {
		(*sig)[i] = 1 - x
	}
}

// AdditiveInverse sets each sample to -x.
func (sig *Discrete) AdditiveInverse() {
	for i, x := range *sig {
		(*sig)[i] = -x
	}
}

// AdditiveSynthesis adds the fundamental, fd, for the partial harmonic, pth, to s.
func AdditiveSynthesis(s Discrete, fd Discrete, pth int) {
	for i := range s {
		j := i * pth % (len(fd) - 1)
		s[i] += fd[j] * (1 / float64(pth))
	}
}

// TODO example potential bike-shedding
//
// func Sine(t float64) float64 { return math.Sin(twopi * t) }
// func SineSampler() Sampler   { return Continuous(Sine) }
// func SineTable(n int) Discrete {
//   sig := make(Discrete, n)
//   SineSampler().Sample(sig, Hertz(n))
//   return sig
// }
//
// var Sine = Continuous(sine)
// func sine(t float64) float64 { return math.Sin(twopi * t) }
// func SineTable(n int) Discrete {
//   sig := make(Discrete, n)
//   Sine.Sample(sig, Hertz(n))
//   return sig
// }

// SineFunc is the continuous signal of a sine wave.
// Every integer value of t is one cycle.
func SineFunc(t float64) float64 {
	return math.Sin(twopi * t)
}

// Sine returns a discrete sample of SineFunc.
func Sine() Discrete {
	sig := make(Discrete, DefaultBufferLen)
	Continuous(SineFunc).Sample(sig, Hertz(DefaultBufferLen))
	return sig
}

// TriangleFunc is the continuous signal of a triangle wave.
func TriangleFunc(t float64) float64 {
	return 2*math.Abs(SawtoothFunc(t)) - 1
}

// Triangle returns a discrete sample of TriangleFunc.
func Triangle() Discrete {
	sig := make(Discrete, DefaultBufferLen)
	Continuous(TriangleFunc).Sample(sig, Hertz(DefaultBufferLen))
	return sig
}

// SquareFunc is the continuous signal of a square wave.
func SquareFunc(t float64) float64 {
	if math.Signbit(SineFunc(t)) {
		return -1
	}
	return 1
}

// Square returns a discrete sample of SquareFunc.
func Square() Discrete {
	sig := make(Discrete, DefaultBufferLen)
	Continuous(SquareFunc).Sample(sig, Hertz(DefaultBufferLen))
	return sig
}

// SawtoothFunc is the continuous signal of a sawtooth wave.
func SawtoothFunc(t float64) float64 {
	return 2 * (t - math.Floor(0.5+t))
}

// Sawtooth returns a discrete sample of SawtoothFunc.
func Sawtooth() Discrete {
	sig := make(Discrete, DefaultBufferLen)
	Continuous(SawtoothFunc).Sample(sig, Hertz(DefaultBufferLen))
	return sig
}

// fundamental default used for sinusoidal synthesis.
var fundamental = Sine()

// SquareSynthesis adds odd partial harmonics belonging to [3..n], creating a sinusoidal wave.
func SquareSynthesis(n int) Discrete {
	sig := Sine()
	for i := 3; i <= n; i += 2 {
		AdditiveSynthesis(sig, fundamental, i)
	}
	sig.Normalize()
	return sig
}

// SawtoothSynthesis adds all partial harmonics belonging to [2..n], creating a sinusoidal wave
// that is the inverse of a sawtooth.
func SawtoothSynthesis(n int) Discrete {
	sig := Sine()
	for i := 2; i <= n; i++ {
		AdditiveSynthesis(sig, fundamental, i)
	}
	sig.Normalize()
	return sig
}
