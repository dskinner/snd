package snd

import (
	"math"
)

const twopi = 2 * math.Pi

type Sampler interface {
	// Sample reads values from src at phase by interval and returns next phase to sample.
	//
	// The side effect resulting from sampling is implementation specific.
	Sample(src Continuous, interval, phase float64) float64
}

// Continuous represents a continuous-time signal.
//
// Effectively, many things may be represented as continuous. A function returning
// math.Sin(2*math.Pi*t) samples a sine wave while Discrete.Interpolate returns an
// interpolated value from its table, with both satisfying this type.
type Continuous func(t float64) float64

// Sample satisfies Sampler interface.
//
// The side effect of fn sampling src can not be known, so this returns phase+interval.
func (fn Continuous) Sample(src Continuous, interval, phase float64) float64 {
	_ = fn(src(phase))
	return phase + interval
}

// Discrete represents a discrete-time signal.
//
// If building a lookup table, let sampling interval equal the recipricol of len(Discrete).
//
//  const n = 1024
//  sig := make(Discrete, n)
//  sig.Sample(SineFunc, 1/n, 0)
//
// If samples are intended to be played back in sequence, provide normalized frequency
// against output sample rate; e.g. to sample four seconds of 440Hz sine wave at 44.1kHz
//
//  r := 44100.0
//  t := 4.0
//  out := make(Discrete, int(r*t)) // allocate four seconds of space
//  out.Sample(SineFunc, 440/r, 0)  // sample 440Hz sine wave
//
// To play these samples over four second period, use an oscillator as clock.
//
//  osc := NewOscil(out, 1/t, nil)
//
// TODO document sampling at different rates.
type Discrete []float64

// Sample satisfies Sampler interface.
func (sig Discrete) Sample(src Continuous, interval, phase float64) float64 {
	for i := 0; i < len(sig); i++ {
		sig[i] = src(phase)
		phase += interval
	}
	return phase
}

// Interp uses the fractional component of t to return an interpolated sample.
//
// TODO currently does linear interpolation...
func (sig Discrete) Interp(t float64) float64 {
	if t <= -0 {
		t = -t
	}
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

// At uses the fractional component of t to return the sample at the truncated index.
func (sig Discrete) At(t float64) float64 {
	if t <= -0 {
		t = -t
	}
	t -= float64(int(t))
	t *= float64(len(sig))
	return sig[int(t)]
}

func (sig Discrete) Index(i int) float64 {
	return sig[i&(len(sig)-1)]
}

// Normalize alters sig so values belong to [-1..1].
func (sig Discrete) Normalize() {
	var max float64
	for _, x := range sig {
		a := math.Abs(x)
		if max < a {
			max = a
		}
	}
	for i, x := range sig {
		sig[i] = x / max
	}
}

// NormalizeRange alters sig so values belong to [s..e].
//
// Calling this method for values that already occupy [s..e] will degrade values
// further due to round-off error.
func (sig Discrete) NormalizeRange(s, e float64) {
	if s > e {
		s, e = e, s
	}
	n := e - s

	var min, max float64
	for _, x := range sig {
		if min > x {
			min = x
		}
		if max < x {
			max = x
		}
	}
	r := max - min
	for i, x := range sig {
		sig[i] = s + n*(x-min)/r
	}
}

// Reverse sig in place so the first element becomes the last and the last element becomes the first.
func (sig Discrete) Reverse() {
	for l, r := 0, len(sig)-1; l < r; l, r = l+1, r-1 {
		sig[l], sig[r] = sig[r], sig[l]
	}
}

// UnitInverse sets each sample to 1 - x.
func (sig Discrete) UnitInverse() {
	for i, x := range sig {
		sig[i] = 1 - x
	}
}

// AdditiveInverse sets each sample to -x.
func (sig Discrete) AdditiveInverse() {
	for i, x := range sig {
		sig[i] = -x
	}
}

func (sig Discrete) MultiplyScalar(x float64) {
	for i := range sig {
		sig[i] *= x
	}
}

func (sig Discrete) Multiply(xs Discrete) {
	for i := range sig {
		sig[i] *= xs.At(float64(i) / float64(len(sig)))
	}
}

// AdditiveSynthesis adds the fundamental, fd, for the partial harmonic, pth, to sig.
func (sig Discrete) AdditiveSynthesis(fd Discrete, pth int) {
	for i := range sig {
		j := i * pth % (len(fd) - 1)
		sig[i] += fd[j] * (1 / float64(pth))
	}
}

// SineFunc is the continuous signal of a sine wave.
func SineFunc(t float64) float64 {
	return math.Sin(twopi * t)
}

// Sine returns a discrete sample of SineFunc.
func Sine() Discrete {
	sig := make(Discrete, 1024)
	sig.Sample(SineFunc, 1./1024, 0)
	return sig
}

// TriangleFunc is the continuous signal of a triangle wave.
func TriangleFunc(t float64) float64 {
	// return 2*math.Abs(SawtoothFunc(t)) - 1
	return 4*math.Abs(t+0.25-math.Floor(0.75+t)) - 1
}

// Triangle returns a discrete sample of TriangleFunc.
func Triangle() Discrete {
	sig := make(Discrete, 1024)
	sig.Sample(TriangleFunc, 1./1024, 0)
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
	sig := make(Discrete, 1024)
	sig.Sample(SquareFunc, 1./1024, 0)
	return sig
}

// SawtoothFunc is the continuous signal of a sawtooth wave.
func SawtoothFunc(t float64) float64 {
	// fmt.Printf("%f %f\n", t, math.Floor(0.5+t))
	return 2 * (t - math.Floor(0.5+t))
}

// Sawtooth returns a discrete sample of SawtoothFunc.
func Sawtooth() Discrete {
	sig := make(Discrete, 1024)
	sig.Sample(SawtoothFunc, 1./1024, 0)
	return sig
}

// fundamental default used for sinusoidal synthesis.
var fundamental = Sine()

// SquareSynthesis adds odd partial harmonics belonging to [3..n], creating a sinusoidal wave.
func SquareSynthesis(n int) Discrete {
	sig := Sine()
	for i := 3; i <= n; i += 2 {
		sig.AdditiveSynthesis(fundamental, i)
	}
	sig.Normalize()
	return sig
}

// SawtoothSynthesis adds all partial harmonics belonging to [2..n], creating a sinusoidal wave
// that is the inverse of a sawtooth.
func SawtoothSynthesis(n int) Discrete {
	sig := Sine()
	for i := 2; i <= n; i++ {
		sig.AdditiveSynthesis(fundamental, i)
	}
	sig.Normalize()
	return sig
}
