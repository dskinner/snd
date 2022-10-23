package snd

import (
	"dasa.cc/signal"
)

// TODO consider functional options

type Oscil struct {
	*mono
	in signal.Discrete

	amp   float64
	freq  float64
	phase float64

	ampmod   Sound
	freqmod  Sound
	phasemod Sound
}

func NewOscil(in signal.Discrete, freq float64, freqmod Sound) *Oscil {
	return &Oscil{
		mono:    newmono(nil),
		in:      in,
		amp:     1,
		freq:    freq,
		freqmod: freqmod,
	}
}

func (osc *Oscil) SetFreq(hz float64, mod Sound) {
	osc.freq = hz
	osc.freqmod = mod
}

func (osc *Oscil) SetAmp(fac float64, mod Sound) {
	osc.amp = fac
	osc.ampmod = mod
}

func (osc *Oscil) SetPhase(mod Sound) {
	osc.phasemod = mod
}

func (osc *Oscil) Inputs() []Sound {
	return []Sound{osc.freqmod, osc.ampmod, osc.phasemod}
}

func (osc *Oscil) Prepare(tc uint64) {
	frame := int(tc-1) * len(osc.out)
	nfreq := osc.freq / osc.sr

	// phase := float64(frame) * nfreq

	for i := range osc.out {
		interval := nfreq
		if osc.freqmod != nil {
			interval *= osc.freqmod.Index(frame + i)
		}

		offset := 0.0
		if osc.phasemod != nil {
			offset = osc.phasemod.Index(frame + i)
		}

		amp := osc.amp
		if osc.ampmod != nil {
			amp *= osc.ampmod.Index(frame + i)
		}

		osc.out[i] = amp * osc.in.At(osc.phase+offset)
		osc.phase += interval
	}
}
