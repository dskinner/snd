package snd

// TODO should mixer be stereo out?
// TODO perhaps this class is unnecessary, any sound could be a mixer
// if you can set multiple inputs, but might get confusing.
type Mixer struct {
	*mono
	ins []Sound
}

func NewMixer(ins ...Sound) *Mixer {
	return &Mixer{
		mono: newmono(nil),
		ins:  ins,
	}
}

func (m *Mixer) Append(s Sound) {
	m.ins = append(m.ins, s)
}

func (m *Mixer) Prepare(tc uint64) (ok bool) {
	if ok = m.mono.Prepare(tc); !ok {
		return
	}
	for _, in := range m.ins {
		in.Prepare(tc)
	}
	for i := range m.out {
		m.out[i] = 0
		if !m.off {
			for _, in := range m.ins {
				m.out[i] += in.Sample(i)
			}
		}
	}
	return
}
