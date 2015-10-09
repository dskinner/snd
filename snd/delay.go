package snd

import "time"

type Delay struct {
	*snd
	dur  float64
	dout []float64

	wpos, rpos int
}

func NewDelay(dur time.Duration, in Sound) *Delay {
	delay := &Delay{
		snd:  newSnd(in),
		wpos: 0,
		rpos: 1,
	}

	delay.dur = float64(dur) / float64(time.Second)
	delay.dout = make([]float64, int(in.SampleRate()*delay.dur))
	return delay
}

func (d *Delay) Prepare() {
	d.snd.Prepare()
	for i := range d.out {
		if d.enabled {
			// read
			d.out[i] = d.dout[d.rpos]
			d.rpos = (d.rpos + 1) % len(d.dout)
			// write
			d.dout[d.wpos] = d.in.Output()[i]
			d.wpos = (d.wpos + 1) % len(d.dout)
		} else {
			d.out[i] = 0
		}
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
		if c.enabled {
			// read
			c.out[i] = c.dout[c.rpos]
			c.rpos = (c.rpos + 1) % len(c.dout)
			// write
			c.dout[c.wpos] = c.in.Output()[i] + c.out[i]*c.gain
			c.wpos = (c.wpos + 1) % len(c.dout)
		} else {
			c.out[i] = 0
		}
	}
}

type Tap struct {
	*snd
	delay *Delay
	dur   float64
	rpos  int
}

func NewTap(dur time.Duration, delay *Delay) *Tap {
	d := float64(dur) / float64(time.Second)
	if d > delay.dur {
		d = delay.dur
	}
	return &Tap{
		snd:   newSnd(delay),
		dur:   d,
		delay: delay,
		rpos:  delay.rpos + int(float64(len(delay.dout))-(d*delay.SampleRate())),
	}
}

func (tap *Tap) Prepare() {
	tap.snd.Prepare()
	for i := range tap.out {
		// read
		tap.out[i] = tap.delay.dout[tap.rpos]
		tap.rpos = (tap.rpos + 1) % len(tap.delay.dout)
	}
}
