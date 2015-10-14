package snd

import "time"

// bufc implements a circular buffer.
type bufc struct {
	xs   []float64
	r, w int
}

func newbufc(n int) *bufc { return &bufc{make([]float64, n), 1, 0} }

func (b *bufc) read() (x float64) {
	x = b.xs[b.r]
	b.r++
	if b.r == len(b.xs) {
		b.r = 0
	}
	return
}

func (b *bufc) write(x float64) (end bool) {
	b.xs[b.w] = x
	b.w++
	end = b.w == len(b.xs)
	if end {
		b.w = 0
	}
	return
}

type Delay struct {
	*mono
	line *bufc
}

func NewDelay(dur time.Duration, in Sound) *Delay {
	// TODO treat as time.Millisecond and then number of samples needed would be
	// 1/(dur/time.Millisecond)
	// probably throw away duration after that can be worked out as
	// len(dout) / in.Channels() / in.SampleRate()
	// ... maybe time to add dat Frames method!
	n := int(in.SampleRate() * float64(dur) / float64(time.Second))
	return &Delay{newmono(in), newbufc(n)}
}

func (dly *Delay) Prepare(uint64) {
	for i := range dly.out {
		if dly.off {
			dly.out[i] = 0
		} else {
			// if d.in != nil {
			// d.in.Prepare(tc)
			// }
			dly.out[i] = dly.line.read()
			dly.line.write(dly.in.Sample(i))
		}
	}
}

type Comb struct {
	*mono
	gain float64
	line *bufc
}

func NewComb(gain float64, dur time.Duration, in Sound) *Comb {
	n := int(in.SampleRate() * float64(dur) / float64(time.Second))
	return &Comb{newmono(in), gain, newbufc(n)}
}

func (cmb *Comb) Prepare(uint64) {
	for i := range cmb.out {
		if cmb.off {
			cmb.out[i] = 0
		} else {
			// if cmb.in != nil {
			// cmb.in.Prepare(tc)
			// }
			cmb.out[i] = cmb.line.read()
			cmb.line.write(cmb.in.Sample(i) + cmb.out[i]*cmb.gain)
		}
	}
}

// type Tap struct {
// *mono
// delay *Delay
// dur   float64
// rpos  int
// }

// func NewTap(dur time.Duration, delay *Delay) *Tap {
// d := float64(dur) / float64(time.Second)
// if d > delay.dur {
// d = delay.dur
// }
// return &Tap{
// mono:  newmono(delay),
// dur:   d,
// delay: delay,
// rpos:  delay.rpos + int(float64(len(delay.dout))-(d*delay.SampleRate())),
// }
// }

// func (tap *Tap) Prepare(tc uint64) {
// if tap.in != nil {
// tap.in.Prepare(tc)
// }
// for i := range tap.out {
// read
// tap.out[i] = tap.delay.dout[tap.rpos]
// tap.rpos = (tap.rpos + 1) % len(tap.delay.dout)
// }
// }

type Looper interface {
	Sound
	Record()
}

func Loop(dur time.Duration, in Sound) Looper {
	lp := &loop{mono: newmono(in)}
	lp.dur = float64(dur) / float64(time.Second)
	lp.dout = make([]float64, int(in.SampleRate()*lp.dur))
	return lp
}

type loop struct {
	*mono
	dur        float64
	dout       []float64
	wpos, rpos int

	count int
	rec   bool
}

func (lp *loop) Record() {
	lp.wpos = 0
	lp.rec = true
}

func (lp *loop) Prepare(uint64) {

	for i := range lp.out {
		if lp.rec {
			if lp.count < len(lp.dout) {
				// write
				lp.dout[lp.wpos] = lp.in.Samples()[i]
				lp.wpos = (lp.wpos + 1) % len(lp.dout)
				lp.count++
			} else {
				lp.rpos = 0
				lp.rec = false
			}
		} else {
			// read
			lp.out[i] = lp.dout[lp.rpos]
			lp.rpos = (lp.rpos + 1) % len(lp.dout)
		}
	}
}
