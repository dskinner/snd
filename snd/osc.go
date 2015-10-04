package snd

type Osc struct {
	*snd

	// TODO how much of this can I just make exported?
	// yes, freq and harm might not be thread safe exactly
	// but what's the worse that could happen if it was swapped out?
	h   Harm
	fr  float64
	idx float64
}

func NewOsc(h Harm, fr float64) *Osc {
	return &Osc{
		snd: newSnd(),
		h:   h,
		fr:  fr,
	}
}

func (osc *Osc) Freq() float64 {
	return osc.fr
}

func (osc *Osc) SetFreq(fr float64) {
	osc.fr = fr
}

func (osc *Osc) Prepare() {
	osc.snd.Prepare()

	var (
		inc float64
		l   float64 = float64(len(osc.h))
		f   float64 = l / osc.snd.sr
	)

	for i := range osc.snd.out {
		osc.snd.out[i] = osc.snd.amp * osc.h[int(osc.idx)]
		inc = osc.fr * f
		osc.idx += inc

		for osc.idx >= l {
			osc.idx -= l
		}
		for osc.idx < 0 {
			osc.idx += l
		}
	}
}
