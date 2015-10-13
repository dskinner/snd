package snd

type Oscillator interface {
	Sound
	Freq(pos int) float64
	SetFreq(hz float64, mod Sound)
	Amp(pos int) float64
	SetAmp(mult float64, mod Sound)
	Phase(pos int) float64
	SetPhase(amt float64, mod Sound)
}

type osc struct {
	*mono

	// TODO how much of this can I just make exported?
	// yes, freq and harm might not be thread safe exactly
	// but what's the worse that could happen if it was swapped out?
	h   Harm
	idx float64

	freq    float64
	freqmod Sound

	amp    float64
	ampmod Sound

	phase    float64
	phasemod Sound
}

func Osc(h Harm, freq float64, freqmod Sound) Oscillator {
	return &osc{
		mono:    newmono(nil),
		h:       h,
		freq:    freq,
		freqmod: freqmod,
		amp:     DefaultAmpMult,
	}
}

func (osc *osc) Freq(i int) float64 {
	if osc.freqmod != nil {
		return osc.freq * osc.freqmod.Sample(i)
	}
	return osc.freq
}

func (osc *osc) SetFreq(hz float64, mod Sound) {
	osc.freq = hz
	osc.freqmod = mod
}

func (osc *osc) Amp(i int) float64 {
	if osc.ampmod != nil {
		return osc.amp * osc.ampmod.Sample(i)
	}
	return osc.amp
}

func (osc *osc) SetAmp(mult float64, mod Sound) {
	osc.amp = mult
	osc.ampmod = mod
}

func (osc *osc) Phase(i int) float64 {
	if osc.phasemod != nil {
		return float64(len(osc.h)) * osc.phasemod.Sample(i)
	}
	return 0
}

func (osc *osc) SetPhase(amt float64, mod Sound) {
	osc.phase = amt
	osc.phasemod = mod
}

func (osc *osc) Prepare(tc uint64) {
	if tc == osc.tc {
		return
	}
	osc.tc = tc

	if osc.freqmod != nil {
		osc.freqmod.Prepare(tc)
	}
	if osc.ampmod != nil {
		osc.ampmod.Prepare(tc)
	}
	if osc.phasemod != nil {
		osc.phasemod.Prepare(tc)
	}

	var (
		l float64 = float64(len(osc.h))
		f float64 = l / osc.sr
	)

	for i := range osc.out {
		if osc.off {
			osc.out[i] = 0
		} else {
			freq := osc.freq
			if osc.freqmod != nil {
				freq *= osc.freqmod.Sample(i)
			}

			amp := osc.amp
			if osc.ampmod != nil {
				amp *= osc.ampmod.Sample(i)
			}

			idx := 0
			if osc.phasemod != nil {
				idx = int(osc.idx+l*osc.phasemod.Sample(i)) & int(l-1)
			} else {
				idx = int(osc.idx) & int(l-1)
			}

			osc.out[i] = amp * osc.h[idx]
			osc.idx += freq * f
		}
	}
}
