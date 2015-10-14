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

func (ring *ring) Prepare(tc uint64) (ok bool) {
	if ok = ring.mono.Prepare(tc); !ok {
		return
	}
	ring.in0.Prepare(tc)
	ring.in1.Prepare(tc)
	for i := range ring.out {
		if ring.off {
			ring.out[i] = 0
		} else {
			ring.out[i] = ring.in0.Samples()[i] * ring.in1.Samples()[i]
		}
	}
	return
}
