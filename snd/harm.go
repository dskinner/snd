package snd

import "math"

// TODO for getting max abs, maybe just outside of loop (benches faster)
// since i'm pretty sure we can't have NaN or -+Inf
// for _, e := range h {
//   if max < e {
//     max = e
//   } else if max < -e {
//     max = -e
//   }
// }

const twopi = 2 * math.Pi

var DefaultHarmLen = 1024

// Harm is a collection of harmonic function evaluations.
type Harm []float64

// Make evaluates fn over the length of h. If h is nil, h will be allocated
// with length DefaultHarmLen.
func (h *Harm) Make(num int, phase float64, fn HarmFunc) {
	if *h == nil {
		*h = make(Harm, DefaultHarmLen)
	}
	fn(*h, num, phase)
}

// HarmFunc defines a harmonic function to be evaluated over the length of h.
type HarmFunc func(h Harm, num int, phase float64)

// SineFunc evaluates a sine wave over the length of h. Argument num is ignored
// given that a sine wave does not have harmonics.
func SineFunc(h Harm, num int, ph float64) {
	var (
		i float64
		l = float64(len(h))
	)
	for ; i < l; i++ {
		h[int(i)] = math.Sin(twopi*(i/l) + ph)
	}
}

// Sine is a helper function returning an evaluated Harm.
func Sine() Harm {
	var h Harm
	h.Make(1, 0, SineFunc)
	return h
}

func SquareFunc(h Harm, num int, ph float64) {
	var (
		i, n float64
		l            = float64(len(h))
		max  float64 = 1
	)

	// TODO should zero out data given += below
	for ; i < l; i++ {
		for n = 1; n <= float64(num); n += 2 {
			h[int(i)] += (1 / n) * math.Sin(n*(twopi*(i/l))+ph)
		}
		abs := math.Abs(h[int(i)])
		if max < abs {
			max = abs
		}
	}

	// normalize
	for i := range h {
		h[i] = h[i] / max
	}
	h[len(h)-1] = h[0]
}

func Square() Harm {
	var h Harm
	h.Make(1, 0, SquareFunc)
	return h
}

// TODO triangle
// TODO seems like triangle is pulse ??? but its not ???

func SawtoothFunc(h Harm, num int, ph float64) {
	var (
		i, n float64
		l            = float64(len(h))
		max  float64 = 1
	)

	// TODO should zero out data given += below
	for ; i < l; i++ {
		for n = 1; n <= float64(num); n++ {
			h[int(i)] += (1 / n) * math.Sin(n*(twopi*(i/l))+ph)
		}
		abs := math.Abs(h[int(i)])
		if max < abs {
			max = abs
		}
	}

	// normalize
	for i := range h {
		h[i] = h[i] / max
	}
	h[len(h)-1] = h[0]
}

func Sawtooth() Harm {
	var h Harm
	h.Make(1, 0, SawtoothFunc)
	return h
}

func PulseFunc(h Harm, num int, ph float64) {
	var (
		i, n float64
		l    = float64(len(h))

		max float64 = 1
	)

	// TODO should zero out data given += below
	for ; i < l; i++ {
		for n = 1; n <= float64(num); n++ {
			h[int(i)] += math.Sin(n*(twopi*(i/l)) + ph)
		}

		abs := math.Abs(h[int(i)])
		if max < abs {
			max = abs
		}
	}

	// normalize
	for i := range h {
		h[i] = h[i] / max
	}
	h[len(h)-1] = h[0]
}

func Pulse() Harm {
	var h Harm
	h.Make(1, 0, PulseFunc)
	return h
}
