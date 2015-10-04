package snd

// TODO should mixer be stereo out?
type Mixer struct {
	ins []Sound
	out []float64
}

func NewMixer(ins ...Sound) *Mixer {
	return &Mixer{
		ins: ins,
		out: make([]float64, DefaultSampleSize),
	}
}

func (m *Mixer) Append(s Sound) {
	m.ins = append(m.ins, s)
}

func (m *Mixer) Output() []float64 {
	return m.out
}

func (m *Mixer) Prepare() {
	for _, in := range m.ins {
		in.Prepare()
	}

	for i := range m.out {
		m.out[i] = 0
		for _, in := range m.ins {
			m.out[i] += in.Output()[i]
		}
		m.out[i] /= float64(len(m.ins))
	}
}
