package snd // import "dasa.cc/piano/snd"

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

/* http://dollopofdesi.blogspot.com/2011/09/interleaving-audio-files-to-different.html

Each second of sound has so many (on a CD, 44,100) digital samples of sound pressure per second. The number of samples per second is called sample rate or sample frequency. In PCM (pulse code modulation) coding, each sample is usually a linear representation of amplitude as a signed integer (sometimes unsigned for 8 bit).
There is one such sample for each channel, one channel for mono, two channels for stereo, four channels for quad, more for surround sound. One sample frame consists of one sample for each of the channels in turn, by convention running from left to right.
Each sample can be one byte (8 bits), two bytes (16 bits), three bytes (24 bits), or maybe even 20 bits or a floating-point number. Sometimes, for more than 16 bits per sample, the sample is padded to 32 bits (4 bytes) The order of the bytes in a sample is different on different platforms. In a Windows WAV soundfile, the less significant bytes come first from left to right ("little endian" byte order). In an AIFF soundfile, it is the other way round, as is standard in Java ("big endian" byte order).
*/

var (
	DefaultSampleRate float64 = 44100
	// TODO force powers of 2
	DefaultSampleSize int = 256
)

type Sound interface {
	// Prepare is when a sound should prepare sample frames.
	//
	// TODO consider passing additional params regarding global state.
	Prepare()

	// Output returns prepared sample frames.
	Output() []float64

	// TODO maybe i want this, maybe I dont
	// Input() Sound
	// SetInput(Sound)
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

// TODO maybe ditch this for mixer which already has a unique signature
// compared to other sounds, make mixer prepare inputs.
// one thing that's different about this is that it doesn't try to output
// all of it's members. That means all snd's should call Prepare on their
// inputs if looking to go only-mixer route.
// type Slice []Sound

// func (sl Slice) Prepare() {
// for _, x := range sl {
// x.Prepare()
// }
// }

// TODO for now, last element of slice should be something like mixer
// or whatever is intended to be played.
// func (sl Slice) Output() []float64 {
// return sl[len(sl)-1].Output()
// }

type snd struct {
	// sample rate
	sr float64

	// amplitude
	amp   float64
	ampin *snd

	// modulation
	mod   float64
	modin *snd

	// TODO how about:
	// phase
	// group

	in  Sound
	out []float64
}

func newSnd(in Sound) *snd {
	return &snd{
		sr:  DefaultSampleRate,
		amp: 1,
		mod: 1,
		in:  in,
		out: make([]float64, DefaultSampleSize),
	}
}

func (s *snd) SetAmp(amp float64, ampin *snd) {
	s.amp = amp
	s.ampin = ampin
}

func (s *snd) SetMod(mod float64, modin *snd) {
	s.mod = mod
	s.modin = modin
}

func (s *snd) SetInput(in Sound) {
	s.in = in
}

func (s *snd) Output() []float64 {
	return s.out
}

func (s *snd) Prepare() {
	if s.in != nil {
		s.in.Prepare()
	}
}

type stereosnd struct {
	l, r *snd
	in   Sound
	out  []float64
}

func newStereosnd(in Sound) *stereosnd {
	return &stereosnd{
		l:   newSnd(nil),
		r:   newSnd(nil),
		in:  in,
		out: make([]float64, DefaultSampleSize*2),
	}
}

func (s *stereosnd) Prepare() {
	// TODO should i do this? Only relevant is SetLeft and SetRight where exposed here
	// s.l.Prepare()
	// s.r.Prepare()
	if s.in != nil {
		s.in.Prepare()
	}
}

// TODO elsewhere, given DefaultSampleSize (or perhaps use a new struct field)
// determine number of channels based on size of []float64 where appropriate (such as openal.go)
func (s *stereosnd) Output() []float64 {
	return s.out
}
