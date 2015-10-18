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
	(*sig)[len(*sig)-1] = (*sig)[0]
}

// Add performs additive synthesis for the given partial harmonic, pt.
func (sig *Discrete) Add(a Discrete, pt int) {
	for i := range *sig {
		j := i * pt % (len(a) - 1)
		(*sig)[i] += a[j] * (1 / float64(pt))
	}
}

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

// SquareSynth adds odd harmonic partials up to n creating a sinusoidal wave.
func SquareSynthesis(n int) Discrete {
	fnd := Sine()
	sig := Sine()
	for i := 3; i <= n; i += 2 {
		sig.Add(fnd, i)
	}
	sig.Normalize()
	return sig
}

// SawtoothSynth adds all harmonic partials up to n creating a sinusoidal wave.
func SawtoothSynthesis(n int) Discrete {
	fnd := Sine()
	sig := Sine()
	for i := 2; i <= n; i++ {
		sig.Add(fnd, i)
	}
	sig.Normalize()
	return sig
}
