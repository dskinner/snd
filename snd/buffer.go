package snd

// Buffer is a crude buffer that accumulates signal during prepare
// but otherwise simply passes through input to output, making this
// pretty useless at the moment.
type Buffer struct {
	*snd

	outs    [][]float64
	samples []float64
}

func NewBuffer(n int, in Sound) *Buffer {
	buf := &Buffer{}
	buf.snd = newSnd(in)
	buf.outs = make([][]float64, n)
	for i := range buf.outs {
		buf.outs[i] = make([]float64, DefaultSampleSize)
	}
	buf.samples = make([]float64, DefaultSampleSize*n)
	return buf
}

func (buf *Buffer) Prepare() {
	buf.snd.Prepare()

	// cycle outputs
	out := buf.outs[0]
	for i := 0; i+1 < len(buf.outs); i++ {
		buf.outs[i] = buf.outs[i+1]
	}
	for i, x := range buf.in.Output() {
		out[i] = x
	}
	buf.outs[len(buf.outs)-1] = out
}

func (buf *Buffer) Output() []float64 { return buf.in.Output() }

func (buf *Buffer) Samples() []float64 {
	// TODO racey given how this method is called
	for i, out := range buf.outs {
		idx := i * DefaultSampleSize
		sl := buf.samples[idx : idx+DefaultSampleSize]
		for j, x := range out {
			sl[j] = x
		}
	}
	return buf.samples
}
