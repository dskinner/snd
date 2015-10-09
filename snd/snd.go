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
	DefaultFrameBufferLen int     = 256 // TODO maybe just call SamplePeriodLen
)

const epsilon = 0.0001

func equals(a, b float64) bool {
	return equaleps(a, b, epsilon)
}

func equaleps(a, b float64, eps float64) bool {
	return (a-b) < eps && (b-a) < eps
}

type Decibel float64

// Amp converts dB to amplitude multiplier.
func (db Decibel) Amp() float64 {
	return math.Pow(10, float64(db)/20)
}

func (db Decibel) String() string {
	return fmt.Sprintf("%vdB", float64(db))
}

type Sound interface {
	SampleRate() float64
	// TODO FrameLen feels ambigious considerings channels is the length of samples in a frame,
	// so frame length almost sounds like number of channels. Consider just calling Frames? NFrames?
	// Or simply replace with Samples() or SamplesLen() which is less ambigious. Then Frames() can be
	// Samples() / Channels(). Or, don't provide either or have Samples() replace Output(), len(Samples())
	// where Samples() is only returning a slice pointer anyway so its cheap and concise.
	FrameLen() int
	Channels() int

	Amp(sample int) float64
	SetAmp(mult float64, mod Sound)

	// Prepare is when a sound should prepare sample frames.
	//
	// TODO consider passing additional params regarding global state.
	Prepare()

	// Output returns prepared sample frames.
	Output() []float64
	OutputAt(pos int) float64

	Enabled() bool
	SetEnabled(bool)

	// TODO maybe i want this, maybe I dont
	// Input() Sound
	// SetInput(Sound)
}

func Mono(in Sound) Sound {
	return newSnd(in)
}

func Stereo(in Sound) Sound {
	return newStereosnd(in)
}

// TODO this is just an example of something I may or may not want
// if i enable Input() and SetInput() on Sound and generify all implementations.
type StereoSound interface {
	Sound
	Left() Sound
	// TODO maybe i need to consider these inputs as Samplers instead of Sounds
	SetLeft(Sound)
	Right() Sound
	SetRight(Sound)
}

// TODO rename to Mono
type snd struct {
	sr float64

	amp    float64
	ampmod Sound

	// TODO how about:
	// phase float64
	// group float64

	in  Sound
	out []float64

	enabled bool
}

func newSnd(in Sound) *snd {
	return &snd{
		sr:      DefaultSampleRate,
		amp:     1,
		in:      in,
		out:     make([]float64, DefaultFrameBufferLen),
		enabled: true,
	}
}

func (s *snd) SampleRate() float64 { return s.sr }
func (s *snd) FrameLen() int       { return len(s.out) / s.Channels() }
func (s *snd) Channels() int       { return 1 }

func (s *snd) Amp(i int) float64 {
	if s.ampmod != nil {
		return s.amp * s.ampmod.Output()[i]
	}
	return s.amp
}

func (s *snd) SetAmp(amp float64, ampmod Sound) {
	s.amp = amp
	s.ampmod = ampmod
}

func (s *snd) SetInput(in Sound) {
	s.in = in
}

func (s *snd) Output() []float64 {
	return s.out
}

// TODO start using this where appropriate
func (s *snd) OutputAt(i int) float64 {
	return s.out[i%len(s.out)]
}

func (s *snd) Prepare() {
	// if s.enabled {
	if s.in != nil {
		s.in.Prepare()
	}
	if s.ampmod != nil {
		s.ampmod.Prepare()
	}
	// }
}

func (s *snd) Enabled() bool {
	return s.enabled
}

func (s *snd) SetEnabled(b bool) {
	s.enabled = b
}

type stereosnd struct {
	l, r *snd
	in   Sound
	out  []float64
}

func newStereosnd(in Sound) *stereosnd {
	return &stereosnd{
		l:  newSnd(nil),
		r:  newSnd(nil),
		in: in,
		// framebufferlen * channels
		out: make([]float64, DefaultFrameBufferLen*2),
	}
}

func (s *stereosnd) SampleRate() float64 { return s.l.sr }
func (s *stereosnd) FrameLen() int       { return len(s.out) / s.Channels() }
func (s *stereosnd) Channels() int       { return 2 }

func (s *stereosnd) Amp(i int) float64 {
	// TODO just sum?
	return (s.l.Amp(i) + s.r.Amp(i)) / 2
}

func (s *stereosnd) SetAmp(amp float64, ampmod Sound) {
	s.l.SetAmp(amp, ampmod)
	s.r.SetAmp(amp, ampmod)
}

func (s *stereosnd) Enabled() bool {
	return s.l.enabled || s.r.enabled
}

func (s *stereosnd) SetEnabled(b bool) {
	s.l.enabled = b
	s.r.enabled = b
}

func (s *stereosnd) Prepare() {
	// TODO should i do this? Only relevant is SetLeft and SetRight where exposed here
	// s.l.Prepare()
	// s.r.Prepare()
	if s.in != nil {
		s.in.Prepare()
	}
}

func (s *stereosnd) Output() []float64 {
	return s.out
}

func (s *stereosnd) OutputAt(i int) float64 {
	return s.out[i%len(s.out)]
}
