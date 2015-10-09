package snd

type Oscillator interface {
	Sound
	SetFreq(freq float64, freqmod Sound)
}

type osc2 struct {
}

type osc struct {
	*snd

	// TODO how much of this can I just make exported?
	// yes, freq and harm might not be thread safe exactly
	// but what's the worse that could happen if it was swapped out?
	h Harm

	freq    float64
	freqmod Sound

	idx float64
}

func Osc(h Harm, freq float64, freqmod Sound) Oscillator {
	return &osc{
		snd:     newSnd(nil),
		h:       h,
		freq:    freq,
		freqmod: freqmod,
	}
}

func (osc *osc) Freq(i int) float64 {
	if osc.freqmod != nil {
		return osc.freq * osc.freqmod.Output()[i]
	}
	return osc.freq
}

func (osc *osc) SetFreq(freq float64, freqmod Sound) {
	osc.freq = freq
	osc.freqmod = freqmod
}

func (osc *osc) Prepare() {
	osc.snd.Prepare()
	if osc.freqmod != nil {
		osc.freqmod.Prepare()
	}

	var (
		inc float64
		l   float64 = float64(len(osc.h))
		f   float64 = l / osc.snd.sr
	)

	for i := range osc.snd.out {
		if osc.enabled {
			freq := osc.Freq(i)

			osc.snd.out[i] = osc.snd.amp * osc.h[int(osc.idx)]
			inc = freq * f
			osc.idx += inc

			for osc.idx >= l {
				osc.idx -= l
			}
			for osc.idx < 0 {
				osc.idx += l
			}
		} else {
			osc.out[i] = 0
		}
	}
}
