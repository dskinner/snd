package snd

import "math"

// TODO reading below
//  http://wilsonminesco.com/16bitMathTables/
//  https://en.wikipedia.org/wiki/Binary_scaling#Binary_angles

/*

Sampling

F(nₖT) is a sampling function.

n = {0, ..., k-1}, for integer values of k

sample period
TODO consider appending *2π here and elsewhere as relevant instead of shoehorning in last minute
T = 1/fₛ

sample frequency
fₛ = 1/T

max sample frequency without error ...
α = max(F(n₁T), ..., F(nₖ-₁T))
this is not correct, samples do not describe max frequency without error
because F(n) is not guaranteed to ever return any particular value.
better might be to say that α is the max fₛ allowed by a system for
normalized values. So ...

all possible sampling frequencies
S = {f₁, ..., fₛ}, set of integers

selection from S
α = S[s]

minimum sample period without error
t = 1/α

let f equal any other selection from S

let e represent error frequency

r = max(fₛ, α) mod min(fₛ, α)
for r != 0, e > 0
e = (max(fₛ, α)-r) / max(fₛ, α)

Example
fₛ = 2
α = 3
r = 3 mod 2 = 1
e = (3-1)/3 = 2/3

So the error frequency is 2/3. This can be seen in the following:

n = {1, 2, 3, 4, 5, 6}
T = 1/fₛ = 1/2
nT*2π = {π, 2π, 3π, 4π, 5π, 6π}

t = 1/α = 1/3
nt*2π = {2π/3, 4π/3, 2π, 8π/3, 10π/3, 4π} // !!! is this appropriate? to sample the same n and then compare?

So α is stepping by 2/3 while fₛ is stepping by 1. These are not aligned.

t() = {2π/3, 4π/3, 2π, 8π/3, 10π/3, 4π}
F() = {   π,   2π, 3π,   4π,    5π, 6π}

t() = {2/3, 4/3, 6/3, 8/3, 10/3, 12/3}
F() = {1/1, 2/1, 3/1, 4/1,  5/1,  6/1}

So where the numerators at t[i] and F[i] are not evenly divisible, this means
the F[i] result is out of phase. The result in this case is the angular frequency.
The frequency is not incorrect, it's just in the wrong place, so its an error with
the angular velocity. Perhaps unintuitively, the fix for this is to change angular
frequency which in effect is actually making the sample period no longer a constant,
b/c we are not changing it arbitrarily, we are setting it from an angular velocity
calculated in a higher bandwidth than available in fₛ.

Theory, if fₛ is scaled upward so that it aligns with α, the additional bandwidth
allows us to calculate the appropriate angular frequency for that phase. In addition,
we can locate the out-of-phase angular frequency to determine exactly how far out of phase
it was.

        1  2  3  4  5  6
t() = -•-•-•-•-•-•------
F() = --•--•--•--•--•--•

well, these results actually scale linearly so right now, not much to see here.
maybe in an effort to simplify using values 2 and 3 instead of 65535,etc, i picked a bad example?
*/

/*
the use of complex would allow for non-constant periods as the actual phase is part of the number

TODO review use of complex numbers
http://scipp.ucsc.edu/~johnson/phys160/ComplexNumbers.pdf // saved to ~/Documents/
Complex numbers are commonly used in electrical engineering, as well as in
physics. In general they are used when some quantity has a phase as well as a
magnitude. Such a situation occurs when one deals with sinusoidal oscillating voltage
and current (other examples in physics include optics, where wave interference is
important, and quantum mechanical wave
functions). I want to emphasize that
complex numbers are used to make
calculations easier!
*/

/*
TODO read about error function! looks fun
https://en.wikipedia.org/wiki/Error_function
*/

type Frequency uint64

// TODO somehow need to map 1Hz to 2*Pi
func (f Frequency) Angular() Radian2 {
	return Radian2(f) * RPiTwo
}

// TODO maybe drop Giga? and add one below nano? Or ditch nanohertz for tetra
// using uint16 rotations uint64 for Radians gives the max value of
// 562,949,953,421,311
// that's five groupings that can be supported
// 44,100Hz is average
// 192,000Hz is hidef for sure, 4 times 48kHz
// 352,800Hz is extreme, 8 times the amount of 44,100
// 5,644,800Hz is highest listed frequency
//
// I think 192kHz should be target so we want space for two times that, 384kHz.
// But basically that means we can go to 999kHz, its just the first group goes to kHz.
//
// This should basically have a one-to-one mapping to Radian, or something like that
// 1Hz = 2Pi rads
//
// These "round" numbers aren't really useful b/c they don't map to Radian at all,
// at least Radian2
const (
	Nanohertz  Frequency = 1
	Microhertz           = 1000 * Nanohertz
	Millihertz           = 1000 * Microhertz
	HHertz               = 1000 * Millihertz
	Kilohertz            = 1000 * HHertz
	Megahertz            = 1000 * Kilohertz
	Gigahertz            = 1000 * Megahertz // caps out at 18.44GHz
	// Tetrahertz           = 1000 * Gigahertz
)

//
// s = sample function
// n = integer value
//
// http://wilsonminesco.com/16bitMathTables/
// https://en.wikipedia.org/wiki/Binary_scaling#Binary_angles
//
// list all alias errors (also see More examples subheading):
// https://en.wikipedia.org/wiki/Aliasing#Sampling_sinusoidal_functions
//
// For precision, i may want to work in something besides hertz, like millihertz, etc
//

type Radian uint64

// func (rad Radian) Hertz() float64 {
// return
// }

func (rad Radian) Degrees() float64 {
	return float64(rad) / math.MaxUint64 * 360
}

func (rad Radian) Float64() float64 {
	return float64(rad) / math.MaxUint64 * twopi
	// 2*Pi equals zero so it can't be done like this
	// return float64(rad*2*Pi) / math.MaxUint64
}

// func (rad Radian) String() string {
// return fmt.Sprintf("%vπ/%v", rad, rad)
// }

// TODO I may want to consider mapping this to first 16 bits to allow for knowing number of rotations.
// That would allow this to exactly represent hertz but I need to figure out what the max hertz would be
const (
	PiOverTwo      Radian = 0x4000000000000000 // 90
	Pi             Radian = 0x8000000000000000 // 180
	PiThreeOverTwo Radian = 0xC000000000000000 // 270
	PiTwo          Radian = 0
	OneRadian      Radian = 0x28BE60DB93910600
)

/*
Rationale for use of Radian Type

I started thinking of tracking progress by the phase. Phase can mean initial offset from zero-point of cycle, but can also mean fraction of a cycle that has elapsed.

For a discrete-time signal, this is the integer number of periods that has elapsed, but phase represents infinite precision and can be used when talking about continuous-time signals. So periods are a countable subset of phase which is the set of all real numbers in [0..1].

Due to this, phase also acts as a normalized value, and much of this package works with normalized values.

For example, if rendering a 440Hz sine wave @ 44.1kHz for 5 seconds, an ideal algorithm should work with a normalized frequency to fill a concrete amount of memory. In this case, 44.1kHz is a cycle, and 440Hz is a phase value of that cycle so the normalized frequency is 440/44100. To play this back as intended, using an oscillator as a clock we can define a cycle as 5 seconds and the sample rate of 44.1kHz as the phase, which equates to 1 second. Setting the oscillators frequency to 1/5 will play the samples back in the intended timing.

These are all normalized values but this is also starting to stretch the definition of phase towards something seemingly arbitrary (though it's quite fitting). Enter, the radian.

The radian IS arbitrary. The radian is a dimensionless unit. The radian is perfectly suitable for use with a continuous-time signal thanks to pi.

Phase can be expressed as radians. But why would we?

When dealing with precision, float64 can lead to errors. Currently, normalized values are expressed as float64 in range [0..1]. In the case of expressing a sample period as a phase value, and then needing to accumulate that value, this can lead to error. This means that depending on the intended use, thought needs to be given to when and where it might be appropriate to use float64 and where it might not be. This can affect design decisions and is another encumberance that demands considering.

Instead, I'm proposing use of the radian type expressed as a uint64. A yet-to-be determined number of least significant bits (minimum int16) can be used as a rotary, with constants for π/2, π, 3π/2, and 2π respectively as 0x40, 0x80, 0xC0, and 0x001 (padded as necessary for intended bits used).

A 16bit rotary allows expressing a max ~281.474THz that advances by ~15.258mHz with each increment.

Question: How does this avoid issues like expressing 1/3? This would be 2π/3 rads which does not have a direct correlary.

If outputing int16 to the hardware, then the max precision of the hardware is ~15.258μHz anyway. Considering the value 1/3, when this doesn't match and gets rounded to the nearest phase,

---

I was reflecting on what makes radians work so well because, in effect, to produce valid output I'd need to multiply the radian value by the reciprocal sample rate also given in radians. This cancels out the value of Pi.

So whats the point? The Pi doesn't even do anything and if i were to stick with ints, I'm effectively just working with rationals.

It's a mindset. It's a reminder. Rendering the floating point value of a radian and working with that doesn't actually make sense because Pi extends infinitely.

Working in radians doesn't provide any tangible benefit over working with rationals computationally. But considering the value of Pi does force you to consider that you're working with an infinite set. Working in radians is a constant reminder of that.
*/

// Having radian as int16 to store rotations works but is not particularly precise. Makes 2Pi multiplication nice though.
type Radian2 uint64

const MaxHertz = 0xFFFFFFFFFFFF0000
const MaxHertz2 = 0xFFFFFFFF00000000
const MaxHertz3 = 0xFFFF000000000000

func (rad Radian2) Hertz() uint64 {
	return uint64(rad / RPiTwo)
	// return float64(rad) / float64(RPiTwo)
}

func (rad Radian2) Degrees() float64 {
	return float64(rad) / float64(RPiTwo) * 360
}

func (rad Radian2) Float64() float64 {
	return float64(rad) / math.MaxUint16 * twopi
	// 2*Pi equals zero so it can't be done like this
	// return float64(rad*2*Pi) / math.MaxUint64
}

const (
	RPiOverTwo      Radian2 = 0x4000 // 90
	RPi             Radian2 = 0x8000 // 180
	RPiThreeOverTwo Radian2 = 0xC000 // 270
	RPiTwo          Radian2 = 0x10000
	ROneRadian      Radian2 = 0x28BE // this would accumulate error pretty sure so ditch it
)

type System struct {
	stub
	// TODO ins should be some generalization as it could be
	// a Discrete or a file. Maybe use Sampler? Or some kind of Reader?
	// Out is always Discrete.
	ins []Discrete
	out Discrete
	sr  float64
}

func (sys *System) Channels() int       { return len(sys.ins) }
func (sys *System) SampleRate() float64 { return sys.sr }
func (sys *System) Samples() Discrete   { return sys.out }

func (sys *System) Prepare(h uint64) {}

// TODO define type Radian
// TODO Radian type?

// TODO rename Signal? due to the potential need for multiple methods
type Sampler interface {
	// Sample reads values from src at the sampling frequency f into dst.
	//
	// If sampling a Discrete, set the sampling frequency to (desiredSampleRate*sourceSampleRate)/len(src).
	//
	// If building a lookup table, let f = len(dst).
	//
	// If samples are intended to be played back in sequence, provide normalized frequency
	// against output sample rate; e.g. sample four seconds of 440Hz sine wave at 44.1kHz
	//
	//  r := 44100.0
	//  t := 4.0
	//  out := make(Discrete, r*t)     // allocate four seconds of space
	//  out.Sample(SineFunc, 440/r, 0) // sample 440Hz sine wave
	//
	// To play samples in realtime, use an oscillator as clock.
	//
	//  osc := NewOscil(out, 1/t, nil)
	//
	// TODO instead of frequency argument, should it just be period? Should probably reference
	// it as normalized frequency. Definitely don't call it "rate".
	// If the frequency is called norm, what's the phase called? Essentially only fractional part
	// of phase is ever used so it's also normalized.
	// Maybe just call it period and mention that is synonymous with norm-freq.
	// Period is a more obvious coralary to phase.
	Sample(src Continuous, interval, phase float64) float64
}

// TODO maybe not ...
// TODO biggest reason for having a Reader.Read type is so that multiple oscillators
// being modulated by this are done so with the same signal and not in a time-invariant way.
// TODO but perhaps instead just make it a method on Discrete that takes tc uint64 and hardware sample-rate
// type Reader interface {
// Read returns successive samples from the reader. Implementations of Read should not panic if the
// underlying data is exhausted.
// Read() float64
// }

/* ************************** */

// TODO reading below
// http://wilsonminesco.com/16bitMathTables/
// https://en.wikipedia.org/wiki/Binary_scaling#Binary_angles

type Buf struct {
	Discrete
	// phase float64
	i int // phase numerator
}

// Reader interface Read doesn't work if used by multiple ... oh well, official nail in the coffin
func (buf *Buf) Read() float64 {
	// x := buf.Index(buf.phase)
	// buf.phase += 1 / float64(len(buf.Discrete))
	x := buf.Index(float64(buf.i) / float64(len(buf.Discrete)))
	buf.i++
	return x
}

// TODO remove when no longer needed
// func phaseErr(c int, phase float64) (newphase, e float64) {
// 	var (
// 		r = 44100.0
// 		f = 441.0
// 		p = f / r
// 		n = 1
// 	)

// 	cphase := float64(c*n) / (r / f)
// 	e = (cphase - phase) / p

// 	for i := 0; i < n; i++ {
// 		phase += p
// 	}
// 	return phase, e
// }
