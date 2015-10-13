package snd

// ring modulator
type ring struct {
	*mono
	in0, in1 Sound
}

func Ring(in0, in1 Sound) Sound {
	return &ring{
		mono: newmono(nil),
		in0:  in0,
		in1:  in1,
	}
}

func (ring *ring) Prepare(tc uint64) {
	ring.in0.Prepare(tc)
	ring.in1.Prepare(tc)
	for i := range ring.out {
		if ring.off {
			ring.out[i] = ring.in0.Samples()[i] * ring.in1.Samples()[i]
		} else {
			ring.out[i] = 0
		}
	}
}
