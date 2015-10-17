package snd

import (
	"fmt"
	"math"
	"time"
)

type ADSR struct {
	*mono

	susamp, maxamp float64

	// sustaining locks envelope
	sustaining bool

	// number of frames for attack, decay, sustain, release
	f0, f1, f2, f3 float64

	// period end markers for attack, decay, sustain, release
	p0, p1, p2, p3 float64

	// current index in envelope
	pn float64
}

// NewADSR
// attack
// time taken for signal to change from 0 to maxamp.
// decay
// time taken for signal to change from maxamp to susamp.
// sustain
// time spent locking signal at susamp.
// release
// time taken after sustain period for signal to change from susamp to 0.
func NewADSR(attack, decay, sustain, release time.Duration, susamp, maxamp float64, in Sound) *ADSR {
	env := &ADSR{
		mono:   newmono(in),
		susamp: susamp,
		maxamp: maxamp,
	}
	sr := env.SampleRate()
	env.f0 = float64(dtof(attack, sr))
	env.f1 = float64(dtof(decay, sr))
	env.f2 = float64(dtof(sustain, sr))
	env.f3 = float64(dtof(release, sr))
	env.p0 = env.f0
	env.p1 = env.f1 + env.f0
	env.p2 = env.f2 + env.f1 + env.f0
	env.p3 = env.f3 + env.f2 + env.f1 + env.f0
	return env
}

func (env *ADSR) Dur() time.Duration {
	return ftod(int(env.p3), env.sr)
}

// Restart resets envelope to start from attack period.
func (env *ADSR) Restart() { env.pn = 0 }

// Sustain locks envelope when sustain period is reached.
func (env *ADSR) Sustain() { env.sustaining = true }

// Release immediately releases envelope from anywhere and starts release period.
func (env *ADSR) Release() (ok bool) {
	env.sustaining = false
	if ok = env.pn <= env.p2; ok {
		env.pn = env.p2 + 1
	}
	return
}

func (env *ADSR) Prepare(uint64) {
	for i := range env.out {
		if env.off {
			env.out[i] = 0
		} else {
			amp := 1.0

			switch {
			case env.pn < env.p0:
				amp = env.maxamp / env.f0 * env.pn
				// amp = env.maxamp * getdrivefac(env.pn/env.p0)
			case env.pn < env.p1:
				amp = env.maxamp - (env.maxamp-env.susamp)*(env.pn-env.p0)/env.f1
				// amp = env.susamp + (env.maxamp-env.susamp)*getdampfac((env.pn-env.p0)/env.f1)
			case env.pn < env.p2:
				amp = env.susamp
			case env.sustaining:
				amp = env.susamp
				env.pn--
			case env.pn < env.p3:
				// amp = env.susamp * ((env.p3 - env.pn) / env.f3)
				// interesting...
				// amp = env.susamp - env.susamp*getdampfac((env.p3-env.pn)/env.f3)
				//
				amp = env.susamp * getdrivefac((env.p3-env.pn)/env.f3)
			default:
				panic(fmt.Errorf("ADSR.pn=%v out of bounds", env.pn))
			}

			env.pn++
			if env.pn == env.p3 {
				env.pn = 0
			}

			if env.in == nil {
				env.out[i] = amp
			} else {
				env.out[i] = amp * env.in.Sample(i)
			}
		}
	}
}

// exponential factors
var expfac [1024]float64

func init() {
	for i := range expfac {
		expfac[i] = math.Exp(twopi * -(float64(i) / float64(len(expfac))))
	}
}

func getdampfac(t float64) float64 {
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	return expfac[int(t*float64(len(expfac)-1))]
}

func getdrivefac(t float64) float64 {
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	n := len(expfac) - 1
	return expfac[n-int(t*float64(n))]
}

// Damp provides decay of a signal, damped signal.
type Damp struct {
	*mono
	amp float64
	pf  int // period frames
	pn  int // period index
}

func NewDamp(d time.Duration, in Sound) *Damp {
	sd := newmono(in)
	return &Damp{mono: sd, amp: 1, pf: dtof(d, sd.SampleRate())}
}

func (dmp *Damp) Prepare(uint64) {
	for i := range dmp.out {
		if dmp.off {
			dmp.out[i] = 0
		} else if dmp.in == nil {
			dmp.out[i] = getdampfac(float64(dmp.pn) / float64(dmp.pf))
		} else {
			dmp.out[i] = dmp.in.Sample(i) * getdampfac(float64(dmp.pn)/float64(dmp.pf))
		}
		dmp.pn++
		if dmp.pn == dmp.pf {
			dmp.pn = 0
		}
	}
}

// Drive gives a driven signal.
type Drive struct {
	*mono
	amp float64
	pf  int
	pn  int
}

func NewDrive(d time.Duration, in Sound) *Drive {
	sd := newmono(in)
	return &Drive{mono: sd, amp: 1, pf: dtof(d, sd.SampleRate())}
}

func (drv *Drive) Prepare(uint64) {
	for i := range drv.out {
		if drv.off {
			drv.out[i] = 0
		} else {
			drv.out[i] = drv.in.Sample(i) * getdrivefac(float64(drv.pn)/float64(drv.pf))
		}
		drv.pn++
		if drv.pn == drv.pf {
			drv.pn = 0
		}
	}
}
