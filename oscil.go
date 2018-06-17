package snd

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

type Oscil struct {
	*mono

	in    Discrete
	phase float64

	freq    float64
	freqmod Sound

	amp    float64
	ampmod Sound

	phasemod Sound

	fmod *Buf
	pmod *Buf
	amod *Buf
}

// TODO consider functional options

func NewOscil(in Discrete, freq float64, freqmod Sound) *Oscil {
	osc := &Oscil{
		mono: newmono(nil),
		in:   in,
		amp:  DefaultAmpFac,
	}
	osc.SetFreq(freq, freqmod)
	return osc
}

func (osc *Oscil) SetFreq(hz float64, mod Sound) {
	osc.freq = hz
	osc.freqmod = mod
	if mod != nil {
		osc.fmod = &Buf{Discrete: Discrete(mod.Samples())}
	}
}

func (osc *Oscil) SetAmp(fac float64, mod Sound) {
	osc.amp = fac
	osc.ampmod = mod
	if mod != nil {
		osc.amod = &Buf{Discrete: Discrete(mod.Samples())}
	}
}

func (osc *Oscil) AddPhase(amt float64) {
	osc.phase += amt
}

func (osc *Oscil) SetPhase(mod Sound) {
	osc.phasemod = mod
	if mod != nil {
		osc.pmod = &Buf{Discrete: Discrete(mod.Samples())}
	}
}

func (osc *Oscil) Inputs() []Sound {
	return []Sound{osc.freqmod, osc.ampmod, osc.phasemod}
}

// TODO remove when no longer needed
func phaseErr(c int, phase float64) (newphase, e float64) {
	var (
		r = 44100.0
		f = 441.0
		p = f / r
		n = 1
	)

	cphase := float64(c*n) / (r / f)
	e = (cphase - phase) / p

	for i := 0; i < n; i++ {
		phase += p
	}
	return phase, e
}

func (osc *Oscil) Prepare(tc uint64) {
	// tc is time relative, multiply by len(osc.out) for first frame of current buffer

	// period is given as 1/f where f is the sampling frequency.
	// f is given as s/r, the desired sample rate divided by the desired frequency.
	// so period is 1/(s/r).
	//
	// if a modulation frequency m is given, then f is instead given as s/(r*m)
	// and period is now given as 1/(s/(r*m)). This is equivalent to m * (1/(s/r)).

	// TODO sampling a signal at some frequency is a form of frequency modulation
	// so reconsider use of term "sampling frequency" in documentation.

	// TODO due to float-error over time, i should stop using float phase? and just use int period?
	// or can it be synced with the uint64 and verifiably proven that errors in phase during a single tick are of no consequence?
	// if syncing with tc is performed, a change of frequency during playback could result in osc's phase becoming out of alignment,
	// so a change of freq should trigger some kind of zero-shift of phase alignment.

	// Have confirmed phase oscillates around len(osc.out) with magnitude getting progressively larger over time (exponentially from a glance)

	// fmt.Println(float64(int(tc-1)*len(osc.out))*osc.freq/osc.sr - osc.phase)

	phase := osc.phase
	period := osc.freq / osc.sr

	for i := range osc.out {
		p := period
		if osc.fmod != nil {
			p *= osc.fmod.Read()
		}

		o := phase
		if osc.phasemod != nil {
			o += osc.pmod.Read()
		}

		a := osc.amp
		if osc.ampmod != nil {
			a *= osc.amod.Read()
		}

		osc.out[i] = a * osc.in.Index(o)
		phase += p
	}
	osc.phase = phase
}

type Osc struct {
	stub
	in   Discrete
	out  Discrete
	amp  float64
	freq float64
	sr   float64
}

func NewOsc(in Discrete, amp float64, fr float64) *Osc {
	return &Osc{
		sr:   DefaultSampleRate,
		in:   in,
		out:  make(Discrete, DefaultBufferLen),
		amp:  amp,
		freq: fr,
	}
}

func (osc *Osc) Prepare(tc uint64) {
	tc -= 1 // TODO sigh ...
	phase := float64(int(tc)*len(osc.out)) * osc.freq / osc.sr
	period := osc.freq / osc.sr
	for i := range osc.out {
		osc.out[i] = osc.amp * osc.in.Index(phase)
		phase += period
	}
}

func (osc *Osc) Channels() int        { return 1 }
func (osc *Osc) Samples() Discrete    { return osc.out }
func (osc *Osc) SampleRate() float64  { return osc.sr }
func (osc *Osc) Sample(i int) float64 { return osc.out[i] }

type OscPa Discrete

func (osc OscPa) Prepare(tc uint64)    {}
func (osc OscPa) SampleRate() float64  { return DefaultSampleRate } // f = sr/len(osc)
func (osc OscPa) Samples() Discrete    { return Discrete(osc) }
func (osc OscPa) Sample(i int) float64 { return Discrete(osc).Index(float64(i) / float64(len(osc))) }
func (osc OscPa) Channels() int        { return 1 }
func (osc OscPa) IsOff() bool          { return false }
func (osc OscPa) Off()                 {}
func (osc OscPa) On()                  {}
func (osc OscPa) Inputs() []Sound      { return nil }
