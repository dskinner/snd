package snd

import "time"

type Delay struct {
	*snd
	dout []float64

	wpos, rpos int
}

func NewDelay(dur time.Duration, in Sound) *Delay {
	return &Delay{
		snd:  newSnd(in),
		dout: make([]float64, int(in.SampleRate()*float64(dur)/float64(time.Second))),
		wpos: 0,
		rpos: 1,
	}
}

func (d *Delay) Prepare() {
	d.snd.Prepare()
	for i := range d.out {
		// read
		d.out[i] = d.dout[d.rpos]
		d.rpos = (d.rpos + 1) % len(d.dout)
		// write
		d.dout[d.wpos] = d.in.Output()[i]
		d.wpos = (d.wpos + 1) % len(d.dout)
	}
}

type Comb struct {
	*Delay
	gain float64
}

func NewComb(dur time.Duration, gain float64, in Sound) *Comb {
	return &Comb{NewDelay(dur, in), gain}
}

func (c *Comb) Prepare() {
	c.snd.Prepare()
	for i := range c.out {
		// read
		c.out[i] = c.dout[c.rpos]
		c.rpos = (c.rpos + 1) % len(c.dout)
		// write
		c.dout[c.wpos] = c.in.Output()[i] + c.out[i]*c.gain
		c.wpos = (c.wpos + 1) % len(c.dout)
	}
}
