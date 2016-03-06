// Package snd provides methods and types for sound processing and synthesis.
//
// Audio hardware is accessed via package snd/al which in turn manages the
// dispatching of sound synthesis via golang.org/x/mobile/audio/al. Start
// the dispatcher as follows:
//
//  const buffers = 1
//  if err := al.OpenDevice(buffers); err != nil {
//      log.Fatal(err)
//  }
//  al.Start()
//
// Once running, add a source for sound synthesis. For example:
//
//  osc := snd.NewOscil(snd.Sine(), 440, nil)
//  al.AddSource(osc)
//
// This results in a 440Hz tone being played back through the audio hardware.
//
// Synthesis types in package snd implement the Sound interface and many type
// methods accept a Sound argument that can affect sampling. For example, one
// may modulate an oscillator by passing in a third argument to NewOscil.
//
//  sine := snd.Sine()
//  mod := snd.NewOscil(sine, 2, nil)
//  osc := snd.NewOscil(sine, 440, mod)
//
// The above results in a lower frequency sound that may require decent speakers
// to hear properly.
//
// Signals
//
// Note the sine argument in the previous example. There are two conceptual types
// of sounds, ContinuousFunc and Discrete. ContinuousFunc represents an indefinite
// series over time. Discrete is the sampling of a ContinuousFunc over an interval.
// Functions such as Sine, Triangle, and Square (non-exhaustive) return Discretes
// created by sampling a ContinuousFunc such as SineFunc, TriangleFunc, and SquareFunc.
//
// Discrete signals serve as a lookup table to efficiently synthesize sound.
// A Discrete is a []float64 and can sample any ContinuousFunc, within the
// package or user defined which is a func(t float64) float64.
//
// Discrete signals may be further modified with intent or arbitrarily. For example,
// Discrete.Add(Discrete, int) performs additive synthesis and is used by functions
// such as SquareSynthesis(int) to return an approximation of a square signal
// based on a sinusoidal.
//
// Time Approximation
//
// Functions that take a time.Duration argument approximate the value to the
// closest number of frames. For example, if sample rate is 44.1kHz and duration
// is 75ms, this results in the argument representing 3307 frames which is
// approximately 74.99ms.
package snd // import "dasa.cc/snd"

import (
	"fmt"
	"math"
	"time"
)

// TODO double check benchmarks, results may be incorrect due to new dispatcher scheme
// TODO pick a consistent api style
// TODO many sounds don't respect off, double check everything
// TODO many sounds only support mono
// TODO support upsampling and downsampling
// TODO migrate most things to Discrete (e.g. mono.out)
// TODO many Prepare funcs need to check if their inputs have altered state (turned on/off, etc)
// during sampling, not just before or after, otherwise this introduces a delay. For example,
// the current defaults of 256 frame length buffer at 44.1kHz would result in a 5.8ms delay.
// Solution needs to account for the updated method for dispatching prepares.
// TODO look into a type Sampler interface { Sample(int) float64 }
// TODO more documentation
// TODO implement sheperd tone for fun:
// https://en.wikipedia.org/wiki/Shepard_tone
// http://music.columbia.edu/cmc/MusicAndComputers/chapter4/04_02.php

const (
	DefaultSampleRate     float64 = 44100
	DefaultSampleBitDepth int     = 16 // TODO not currently used for anything
	DefaultBufferLen      int     = 256
	DefaultAmpFac         float64 = 0.31622776601683794 // -10dB
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

// Hertz is defined as cycles per second and is synonymous with frequency.
type Hertz float64

// Angular returns the angular frequency as 2 * pi * hz and is synonymous with radians.
func (hz Hertz) Angular() float64 {
	return twopi * float64(hz)
}

// Normalized returns the angular frequency of hz divided by the sample rate sr.
func (hz Hertz) Normalized(sr float64) float64 {
	return hz.Angular() / sr
}

func (hz Hertz) String() string {
	return fmt.Sprintf("%vHz", float64(hz))
}

// BPM respresents beats per minute and is a measure of tempo.
type BPM float64

// Dur returns the time duration of bpm.
func (bpm BPM) Dur() time.Duration {
	return time.Duration(float64(time.Minute) / float64(bpm))
}

// Hertz returns the frequency of bpm as bpm / 2.
func (bpm BPM) Hertz() float64 {
	return float64(bpm) / 2
}

// Sound represents a type capable of producing sound data.
type Sound interface {
	// SampleRate returns the number of digital samples of sound pressure per second.
	SampleRate() float64

	// Channels returns the frame size in samples of the internal buffer.
	Channels() int

	// BufferLen returns size of internal buffer in samples.
	BufferLen() int

	// SetBufferLen resizes the internal buffer to n samples.
	SetBufferLen(n int)

	// Prepare is when a sound should prepare sample frames.
	//
	// TODO consider prepare as proxy of some new method, such as Do(), to prevent
	// unecessary prepares over the same uint64. But, this could also be managed
	// by the new dispatcher.
	Prepare(uint64)

	// Samples returns prepared samples slice.
	//
	// TODO maybe ditch this, point of architecture is you can't mess
	// with an input's output but a slice exposes that. Or, discourage
	// use by making a copy of data.
	Samples() []float64

	// Sample returns the sample at pos mod BufferLen().
	Sample(pos int) float64

	// On allows calls to Prepare to fill the internal buffer with valid data.
	On()

	// Off makes calls to Prepare write out zeros to the internal buffer.
	Off()

	// IsOff identifies if this sound is off.
	//
	// TODO meh ... work out something better.
	IsOff() bool

	// Inputs should return all inputs a Sound wants discoverable (for dispatcher).
	Inputs() []Sound
}

// Mono returns a new mono channel that receives input from in.
//
// TODO this is ill-conceived
func Mono(in Sound) Sound { return newmono(in) }

// Stereo returns a new stereo channel that receives input from in.
//
// TODO this is ill-conceived
func Stereo(in Sound) Sound { return newstereo(in) }

// StereoSound represents a regular Sound with left and right channels.
//
// TODO this is just an example of something I may or may not want if I enable
// Input() and SetInput() on Sound and generify all implementations.
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

func (sd *mono) SampleRate() float64 { return sd.sr }
func (sd *mono) Samples() []float64 {
	// out := make([]float64, len(sd.out))
	// copy(out, sd.out)
	// return out
	return sd.out
}
func (sd *mono) Sample(i int) float64 { return sd.out[i&(len(sd.out)-1)] }
func (sd *mono) Channels() int        { return 1 }
func (sd *mono) BufferLen() int       { return len(sd.out) }
func (sd *mono) SetBufferLen(n int) {
	if n == 0 || n&(n-1) != 0 {
		panic(fmt.Errorf("snd: SetBufferLen(%v) not a power of 2", n))
	}
	sd.out = make([]float64, n)
}
func (sd *mono) IsOff() bool { return sd.off }
func (sd *mono) Off() {
	sd.off = true
	// if sd.in != nil {
	// sd.in.Off()
	// }
}
func (sd *mono) On() {
	sd.off = false
	// if sd.in != nil {
	// sd.in.On()
	// }
}

func (sd *mono) Inputs() []Sound { return []Sound{sd.in} }

// TODO consider not having mono or stereo actually implement sound by remove this
func (sd *mono) Prepare(uint64) {}

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

func (sd *stereo) SampleRate() float64 { return sd.l.sr }
func (sd *stereo) Samples() []float64 {
	// out := make([]float64, len(sd.out))
	// copy(out, sd.out)
	// return out
	return sd.out
}
func (sd *stereo) Sample(i int) float64 { return sd.out[i&(len(sd.out)-1)] }
func (sd *stereo) Channels() int        { return 2 }
func (sd *stereo) BufferLen() int       { return len(sd.out) }
func (sd *stereo) SetBufferLen(n int) {
	if n == 0 || n&(n-1) != 0 {
		panic(fmt.Errorf("snd: SetBufferLen(%v) not a power of 2", n))
	}
	sd.out = make([]float64, n*2)
}
func (sd *stereo) IsOff() bool { return sd.l.off || sd.r.off }
func (sd *stereo) Off()        { sd.l.off, sd.r.off = false, false }
func (sd *stereo) On()         { sd.l.off, sd.r.off = true, true }

func (sd *stereo) Inputs() []Sound { return []Sound{sd.in} }

func (sd *stereo) Prepare(tc uint64) {}
