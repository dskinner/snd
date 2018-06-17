package main

import (
	"math"
	"time"

	"dasa.cc/snd"
	"dasa.cc/snd/al"
)

var (
	sawtooth = snd.Sawtooth()
	sawsine  = snd.SawtoothSynthesis(8)
	square   = snd.Square()
	sqsine   = snd.SquareSynthesis(49)
	sine     = snd.Sine()
	triangle = snd.Triangle()
	notes    = snd.EqualTempermant(12, 440, 48)

	keys       [12]Key
	reverb     *snd.LowPass
	metronome  *snd.Mixer
	loop       *snd.Loop
	lowpass    *snd.LowPass
	keymix     *snd.Mixer
	keygain    *snd.Gain
	master     *snd.Mixer
	mastergain *snd.Gain

	bpm     = snd.BPM(80)
	loopdur = snd.Dtof(bpm.Dur(), snd.DefaultSampleRate) * 8 //nframes

	sndbank    = []KeyFunc{NewPianoKey, NewWobbleKey, NewBeatsKey, NewReeseKey}
	sndbankpos = 0

	ms = time.Millisecond
)

func makekeys() {
	keymix.Empty()
	for i := range keys {
		keys[i] = sndbank[sndbankpos](51 + i) // notes[51] is Major C
		keys[i].Freeze()
		keymix.Append(keys[i])
	}
	al.Notify()
}

type Key interface {
	snd.Sound
	Press()
	Release()
	Freeze()
}

type KeyFunc func(int) Key

type BeatsKey struct {
	*snd.Instrument
	adsr *snd.ADSR
}

func NewBeatsKey(idx int) Key {
	osc := snd.NewOscil(sawsine, notes[idx], snd.NewOscil(triangle, 4, nil))
	dmp := snd.NewDamp(bpm.Dur(), osc)
	d := snd.BPM(float64(bpm) * 1.25).Dur()
	dmp1 := snd.NewDamp(d, osc)
	drv := snd.NewDrive(d, osc)
	mix := snd.NewMixer(dmp, dmp1, drv)

	frz := snd.NewFreeze(bpm.Dur()*4, mix)

	adsr := snd.NewADSR(250*ms, 500*ms, 300*ms, 400*ms, 0.85, 1.0, frz)
	key := &BeatsKey{snd.NewInstrument(adsr), adsr}
	key.Off()
	return key
}

func (key *BeatsKey) Press() {
	key.adsr.Restart()
	key.adsr.Sustain()
	key.On()
}
func (key *BeatsKey) Release() {
	key.adsr.Release()
	key.OffIn(400 * ms)
}
func (key *BeatsKey) Freeze() {}

type WobbleKey struct {
	*snd.Instrument
	adsr *snd.ADSR
}

func NewWobbleKey(idx int) Key {
	osc := snd.NewOscil(sine, notes[idx], snd.NewOscil(triangle, 2, nil))
	adsr := snd.NewADSR(50*ms, 100*ms, 200*ms, 400*ms, 0.6, 0.9, osc)
	key := &WobbleKey{snd.NewInstrument(adsr), adsr}
	key.Off()
	return key
}

func (key *WobbleKey) Press() {
	key.adsr.Restart()
	key.adsr.Sustain()
	key.On()
}
func (key *WobbleKey) Release() {
	key.adsr.Release()
	key.OffIn(400 * ms)
}
func (key *WobbleKey) Freeze() {}

type ReeseKey struct {
	*snd.Instrument
	adsr *snd.ADSR
}

func NewReeseKey(idx int) Key {
	sine := snd.Sawtooth()
	freq := notes[idx-36]
	osc0 := snd.NewOscil(sine, freq*math.Pow(2, 30.0/1200), nil)
	osc0.SetAmp(snd.Decibel(-3).Amp(), nil)
	osc1 := snd.NewOscil(sine, freq*math.Pow(2, -30.0/1200), nil)
	osc1.SetAmp(snd.Decibel(-3).Amp(), nil)

	freq = notes[idx-24]
	osc2 := snd.NewOscil(sine, freq, nil)
	osc2.SetAmp(snd.Decibel(-6).Amp(), nil)
	osc3 := snd.NewOscil(sine, freq, nil)
	osc3.SetAmp(snd.Decibel(-6).Amp(), nil)

	mix := snd.NewMixer(osc0, osc1, osc2, osc3)
	adsr := snd.NewADSR(1*time.Millisecond, 750*time.Millisecond, 1*time.Millisecond, 2*time.Millisecond, 0.75, 0.8, mix)

	key := &ReeseKey{snd.NewInstrument(adsr), adsr}
	key.Off()
	return key
}

func (key *ReeseKey) Press() {
	key.adsr.Restart()
	key.adsr.Sustain()
	key.On()
}
func (key *ReeseKey) Release() {
	key.adsr.Release()
	key.OffIn(2 * ms)
}
func (key *ReeseKey) Freeze() {}

type PianoKey struct {
	*snd.Instrument

	freq float64

	osc, mod, phs    *snd.Oscil
	oscl, modl, phsl *snd.Oscil
	oscr, modr, phsr *snd.Oscil

	adsr0, adsr1 *snd.ADSR

	gain *snd.Gain

	dur    time.Duration
	reldur time.Duration

	frz *snd.Freeze
}

func NewPianoKey(idx int) Key {
	const phasefac float64 = 0.5063999999999971

	k := &PianoKey{}

	k.freq = notes[idx]
	k.mod = snd.NewOscil(sqsine, k.freq/2, nil)
	k.osc = snd.NewOscil(sawtooth, k.freq, k.mod)
	k.phs = snd.NewOscil(square, k.freq*phasefac, nil)
	k.osc.SetPhase(k.phs)

	freql := k.freq * math.Pow(2, -10.0/1200)
	k.modl = snd.NewOscil(sqsine, freql/2, nil)
	k.oscl = snd.NewOscil(sawtooth, freql, k.modl)
	k.phsl = snd.NewOscil(square, freql*phasefac, nil)
	k.oscl.SetPhase(k.phsl)

	freqr := k.freq * math.Pow(2, 10.0/1200)
	k.modr = snd.NewOscil(sqsine, freqr/2, nil)
	k.oscr = snd.NewOscil(sawtooth, freqr, k.modr)
	k.phsr = snd.NewOscil(square, freqr*phasefac, nil)
	k.oscr.SetPhase(k.phsr)

	oscmix := snd.NewMixer(k.osc, k.oscl, k.oscr)

	k.reldur = 1050 * ms
	k.dur = 280*ms + k.reldur
	k.adsr0 = snd.NewADSR(30*ms, 50*ms, 200*ms, k.reldur, 0.4, 1, oscmix)
	k.adsr1 = snd.NewADSR(1*ms, 278*ms, 1*ms, k.reldur, 0.4, 1, oscmix)
	adsrmix := snd.NewMixer(k.adsr0, k.adsr1)

	k.gain = snd.NewGain(snd.Decibel(-10).Amp(), adsrmix)

	k.Instrument = snd.NewInstrument(k.gain)
	k.Off()

	return k
}

func (key *PianoKey) Freeze() {
	key.On()
	key.frz = snd.NewFreeze(key.dur, key.gain)
	key.Instrument = snd.NewInstrument(key.frz)
	key.Off()
}

func (key *PianoKey) Press() {
	key.On()
	key.OffIn(key.dur)
	if key.frz == nil {
		key.adsr0.Restart()
		key.adsr1.Restart()
	} else {
		key.frz.Restart()
	}
}

func (key *PianoKey) Release() {
	if key.frz == nil {
		if key.adsr0.Release() && key.adsr1.Release() {
			key.OffIn(key.reldur)
		}
	}
}
