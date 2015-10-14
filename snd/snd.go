package snd // import "dasa.cc/piano/snd"
import (
	"fmt"
	"math"
)

// TODO
// https://en.wikipedia.org/wiki/Piano_key_frequencies
// https://en.wikipedia.org/wiki/Fundamental_frequency
// https://en.wikipedia.org/wiki/Harmonic_series_(mathematics)
// https://en.wikipedia.org/wiki/Sine_wave
// https://en.wikipedia.org/wiki/Triangle_wave
// https://en.wikipedia.org/wiki/Additive_synthesis
// https://en.wikipedia.org/wiki/Wave
// https://en.wikipedia.org/wiki/Waveform
// https://en.wikipedia.org/wiki/Wavelength
// https://en.wikipedia.org/wiki/Sampling_(signal_processing)
// http://public.wsu.edu/~jkrug/MUS364/audio/Waveforms.htm
// https://en.wikipedia.org/wiki/Window_function#Hamming_window
// http://uenics.evansville.edu/~amr63/equipment/scope/oscilloscope.html
//
// http://www.csounds.com/manual/html/

/* http://dollopofdesi.blogspot.com/2011/09/interleaving-audio-files-to-different.html

Each second of sound has so many (on a CD, 44,100) digital samples of sound pressure per second.
The number of samples per second is called sample rate or sample frequency.
In PCM (pulse code modulation) coding, each sample is usually a linear representation of amplitude
as a signed integer (sometimes unsigned for 8 bit).

There is one such sample for each channel, one channel for mono, two channels for stereo,
four channels for quad, more for surround sound. One sample frame consists of one sample
for each of the channels in turn, by convention running from left to right.

Each sample can be one byte (8 bits), two bytes (16 bits), three bytes (24 bits), or maybe
even 20 bits or a floating-point number. Sometimes, for more than 16 bits per sample,
the sample is padded to 32 bits (4 bytes) The order of the bytes in a sample is different
on different platforms. In a Windows WAV soundfile, the less significant bytes come first
from left to right ("little endian" byte order). In an AIFF soundfile, it is the other way
round, as is standard in Java ("big endian" byte order).
*/

// TODO most everything in this package is not correctly handling/respecting Enabled()

const (
	DefaultSampleRate     float64 = 44100
	DefaultSampleBitDepth int     = 16 // TODO not currently used for anything
	DefaultBufferLen      int     = 256
	DefaultAmpMult        float64 = 0.31622776601683794 // -10dB
)

const epsilon = 0.0001

func equals(a, b float64) bool {
	return equaleps(a, b, epsilon)
}

func equaleps(a, b float64, eps float64) bool {
	return (a-b) < eps && (b-a) < eps
}

// Decibel is relative to full scale; anything over 0dB will clip.
type Decibel float64

// Amp converts dB to amplitude multiplier.
func (db Decibel) Amp() float64 {
	return math.Pow(10, float64(db)/20)
}

func (db Decibel) String() string {
	return fmt.Sprintf("%vdB", float64(db))
}

type Hertz float64

func (hz Hertz) Angular() float64 {
	return twopi * float64(hz)
}

func (hz Hertz) Normalized(sr float64) float64 {
	return hz.Angular() / sr
}

type Sound interface {
	SampleRate() float64
	Channels() int

	// BufferLen returns size of internal buffer in samples.
	BufferLen() int
	SetBufferLen(int)

	// Prepare is when a sound should prepare sample frames. returns true if ok to continue.
	// TODO actually probably possible to completely avoid all the ok = ...; !ok by just making
	// this a different method that no one is overriding ...
	// that also avoids all the times code ignores the ok
	// basically, just have said new method be proxy that calls Prepare(uint64) or maybe just go back
	// to method Do(uint64) and have Prepare be proxy.
	// would be nice to have that be where dependent prepares are called, like Osc has so many
	// after prepares are done, i suppose all output is actually parallelizable at that point ... hmmm, is it?
	Prepare(uint64)

	// TODO maybe ditch this, point of architecture is you can't mess
	// with an input's output but a slice exposes that. Or, discourage
	// use by making a copy of data.
	// Samples returns prepared samples slice.
	Samples() []float64

	// Sample returns the sample at pos mod BufferLen().
	Sample(pos int) float64

	Off()
	On()

	// Inputs should return all inputs a Sound wants discoverable (for auto-prepare).
	Inputs() []Sound
}

func Mono(in Sound) Sound { return newmono(in) }

func Stereo(in Sound) Sound { return newstereo(in) }

// TODO this is just an example of something I may or may not want
// if i enable Input() and SetInput() on Sound and generify all implementations.
type StereoSound interface {
	Sound
	Left() Sound
	SetLeft(Sound)
	Right() Sound
	SetRight(Sound)
}

type mono struct {
	sr  float64
	in  Sound
	out []float64
	tc  uint64
	off bool
}

func newmono(in Sound) *mono {
	return &mono{
		sr:  DefaultSampleRate,
		in:  in,
		out: make([]float64, DefaultBufferLen),
	}
}

func (s *mono) SampleRate() float64  { return s.sr }
func (s *mono) Samples() []float64   { return s.out }
func (s *mono) Sample(i int) float64 { return s.out[i&(len(s.out)-1)] }
func (s *mono) Channels() int        { return 1 }
func (s *mono) BufferLen() int       { return len(s.out) }
func (s *mono) SetBufferLen(n int) {
	if n == 0 || n&(n-1) != 0 {
		panic(fmt.Errorf("snd: SetBufferLen(%v) not a power of 2", n))
	}
	s.out = make([]float64, n)
}
func (s *mono) IsOff() bool { return s.off }
func (s *mono) Off()        { s.off = true }
func (s *mono) On()         { s.off = false }

func (s *mono) Inputs() []Sound { return []Sound{s.in} }

// TODO consider not having mono or stereo actually implement sound by remove this
func (s *mono) Prepare(uint64) {}

type stereo struct {
	l, r *mono
	in   Sound
	out  []float64
	tc   uint64
}

func newstereo(in Sound) *stereo {
	return &stereo{
		l:   newmono(nil),
		r:   newmono(nil),
		in:  in,
		out: make([]float64, DefaultBufferLen*2),
	}
}

func (s *stereo) SampleRate() float64  { return s.l.sr }
func (s *stereo) Samples() []float64   { return s.out }
func (s *stereo) Sample(i int) float64 { return s.out[i%len(s.out)] }
func (s *stereo) Channels() int        { return 2 }
func (s *stereo) BufferLen() int       { return len(s.out) }
func (s *stereo) SetBufferLen(n int) {
	if n == 0 || n&(n-1) != 0 {
		panic(fmt.Errorf("snd: SetBufferLen(%v) not a power of 2", n))
	}
	s.out = make([]float64, n*2)
}
func (s *stereo) IsOff() bool { return s.l.off || s.r.off }
func (s *stereo) Off()        { s.l.off, s.r.off = false, false }
func (s *stereo) On()         { s.l.off, s.r.off = true, true }

func (s *stereo) Inputs() []Sound { return []Sound{s.in} }

func (s *stereo) Prepare(tc uint64) {
}
