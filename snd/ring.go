package snd

// ring modulator
type Ring struct {
	*snd
	in0, in1 Sound
}

func NewRing(in0, in1 Sound) *Ring {
	return &Ring{
		snd: newSnd(nil),
		in0: in0,
		in1: in1,
	}
}

func (ring *Ring) Prepare() {
	ring.in0.Prepare()
	ring.in1.Prepare()
	for i := range ring.out {
		if ring.enabled {
			ring.out[i] = ring.in0.Output()[i] * ring.in1.Output()[i]
		} else {
			ring.out[i] = 0
		}
	}
}
