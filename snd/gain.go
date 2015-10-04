package snd

type Gain struct {
	*snd

	mult float64
}

func NewGain(mult float64) *Gain {
	return &Gain{
		snd:  newSnd(),
		mult: mult,
	}
}

func (g *Gain) Prepare() {
	g.snd.Prepare()
	for i, x := range g.in.Output() {
		g.out[i] = x * g.mult
	}
}
