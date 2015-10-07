package snd

// TODO should mixer be stereo out?
type Mixer struct {
	ins []Sound
	out []float64

	enabled bool

	quiets []Sound
}

func NewMixer(ins ...Sound) *Mixer {
	m := &Mixer{
		ins: ins,
		out: make([]float64, DefaultSampleSize),
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

func (m *Mixer) Enabled() bool     { return m.enabled }
func (m *Mixer) SetEnabled(b bool) { m.enabled = b }

func (m *Mixer) Output() []float64 {
	return m.out
}

func (m *Mixer) Prepare() {
	for _, in := range m.ins {
		in.Prepare()
	}

	// TODO tmp
	for _, quiet := range m.quiets {
		quiet.Prepare()
	}

	for i := range m.out {
		m.out[i] = 0
		for _, in := range m.ins {
			if in.Enabled() {
				m.out[i] += in.Output()[i]
			}
		}
		m.out[i] /= float64(len(m.ins))
	}
}
