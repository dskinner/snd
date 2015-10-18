package snd

import (
	"fmt"
	"math"
)

const twopi = 2 * math.Pi

const defaultDiscreteLen = 1024

type ContinuousFunc func(t float64) float64

type Discrete []float64

func (sig *Discrete) Sample(fn ContinuousFunc) {
	if *sig == nil {
		*sig = make(Discrete, defaultDiscreteLen)
	}
	if n := len(*sig); n == 0 || n&(n-1) != 0 {
		panic(fmt.Errorf("Discrete len(%v) not a power of 2", n))
	}
	n := float64(len(*sig))
	for i := 0.0; i < n; i++ {
		(*sig)[int(i)] = fn(i / n)
	}
}

func (sig *Discrete) Normalize() {
	max := -1.0
	for _, x := range *sig {
		if max < x {
			max = x
		}
	}
	for i, x := range *sig {
		(*sig)[i] = x / max
	}
	(*sig)[len(*sig)-1] = (*sig)[0]
}

// Add performs additive synthesis.
func (sig *Discrete) Add(a Discrete, ph float64) {
	if len(*sig) != len(a) {
		panic("lengths do not match")
	}
	// https://en.wikipedia.org/wiki/Additive_synthesis
	for i, x := range *sig {
		if i == 0 {
			continue
		}
		(*sig)[i] = x * math.Cos(a[i]+ph)
	}
}

// SynthesisFunc produces a discrete signal constructed using additive synthesis.
// TODO how could this be removed and perform additive/subtractive synthesis via
// methods on Discrete.
type SynthesisFunc func(h int, ph float64) Discrete

// SineFunc represents the continuous signal of a sine wave.
func SineFunc(t float64) float64 {
	return math.Sin(twopi * t)
}

func Sine() (sig Discrete) {
	sig.Sample(SineFunc)
	return
}

func TriangleFunc(t float64) float64 {
	return 2*math.Abs(SawtoothFunc(t)) - 1
}

func Triangle() (sig Discrete) {
	sig.Sample(TriangleFunc)
	return
}

// SquareFunc represents the continuous signal of a square wave.
func SquareFunc(t float64) float64 {
	if math.Signbit(math.Sin(twopi * t)) {
		return -1
	}
	return 1
}

func Square() (sig Discrete) {
	sig.Sample(SquareFunc)
	return
}

// SawtoothFunc represents the continuous signal of a sawtooth wave.
func SawtoothFunc(t float64) float64 {
	return 2 * (t - math.Floor(0.5+t))
}

func Sawtooth() (sig Discrete) {
	sig.Sample(SawtoothFunc)
	return
}

// SquareSynthesis returns a discrete signal of a square wave constructed
// using additive synthesis with h harmonics and ph phase.
func SquareSynthesis(h int, ph float64) Discrete {
	sig := make(Discrete, defaultDiscreteLen)
	for i := 0.0; i < defaultDiscreteLen; i++ {
		t := i / defaultDiscreteLen
		for n := 1.0; n <= float64(h); n += 2 {
			sig[int(i)] += (1 / n) * math.Sin(n*(twopi*t)+ph)
		}
	}
	sig.Normalize()
	return sig
}

func SawtoothSynthesis(h int, ph float64) Discrete {
	sig := make(Discrete, defaultDiscreteLen)
	for i := 0.0; i < defaultDiscreteLen; i++ {
		t := i / defaultDiscreteLen
		for n := 1.0; n <= float64(h); n++ {
			sig[int(i)] += (1 / n) * math.Sin(n*(twopi*t)+ph)
		}
	}
	sig.Normalize()
	return sig
}

func PulseSynthesis(h int, ph float64) Discrete {
	sig := make(Discrete, defaultDiscreteLen)
	for i := 0.0; i < defaultDiscreteLen; i++ {
		t := i / defaultDiscreteLen
		for n := 1.0; n <= float64(h); n++ {
			sig[int(i)] += math.Sin(n*(twopi*t) + ph)
		}
	}
	sig.Normalize()
	return sig
}
