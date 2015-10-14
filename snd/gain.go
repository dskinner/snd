package snd

type gain struct {
	*mono
	mult float64
}

func Gain(mult float64, in Sound) Sound { return &gain{newmono(in), mult} }

func (g *gain) SetMult(mult float64) { g.mult = mult }

func (g *gain) Prepare(tc uint64) {
	for i, x := range g.in.Samples() {
		g.out[i] = x * g.mult
	}
}
