package snd

// TODO should mixer be stereo out?
// TODO perhaps this class is unnecessary, any sound could be a mixer
// if you can set multiple inputs, but might get confusing.
type Mixer struct {
	*snd

	ins    []Sound
	quiets []Sound
}

func NewMixer(ins ...Sound) *Mixer {
	m := &Mixer{
		snd: newSnd(nil),
		ins: ins,
	}
	return m
}

func (m *Mixer) Append(s Sound) {
	m.ins = append(m.ins, s)
}

// TODO temp method
func (m *Mixer) AppendQuiet(s Sound) {
	m.quiets = append(m.quiets, s)
}

func (m *Mixer) Prepare() {
	m.snd.Prepare()

	for _, in := range m.ins {
		in.Prepare()
	}

	// TODO tmp
	for _, quiet := range m.quiets {
		quiet.Prepare()
	}

	for i := range m.out {
		m.out[i] = 0
		if m.enabled {
			for _, in := range m.ins {
				m.out[i] += in.Output()[i]
			}
		}
	}
}
