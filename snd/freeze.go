package snd

import "time"

type Freeze struct {
	*mono
	sig Discrete
	pos int
	nfr int
}

func NewFreeze(d time.Duration, in Sound) *Freeze {
	frz := &Freeze{mono: newmono(nil), nfr: Dtof(d, DefaultSampleRate)}
	n := frz.nfr
	for n == 0 || n&(n-1) != 0 {
		n++
	}
	frz.sig = make(Discrete, n)
	inps := GetInputs(in)
	dp := new(Dispatcher)
	for i := 0; i < n; i += DefaultBufferLen {
		dp.Dispatch(1, inps...)
		for j := 0; j < DefaultBufferLen; j++ {
			frz.sig[i+j] = in.Sample(j)
		}
	}
	return frz
}

func (frz *Freeze) Pos() int { return frz.pos }

func (frz *Freeze) SetPos(pos int) { frz.pos = pos }

func (frz *Freeze) Prepare(uint64) {
	for i := range frz.out {
		if frz.off {
			frz.out[i] = 0
		} else {
			frz.out[i] = frz.sig[frz.pos]
			frz.pos++
			if frz.pos == frz.nfr {
				frz.pos = 0
			}
		}
	}
}
