package snd

import (
	"fmt"
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
	env.p1 = env.f1 + env.p0
	env.p2 = env.f2 + env.p1 + env.p0
	env.p3 = env.f3 + env.p2 + env.p1 + env.p0
	return env
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
			case env.pn < env.p1:
				amp = env.maxamp - (env.maxamp-env.susamp)/env.f1*(env.pn-env.p0)
			case env.pn < env.p2:
				amp = env.susamp
			case env.sustaining:
				amp = env.susamp
				env.pn--
			case env.pn < env.p3:
				amp = ((-env.susamp)/env.f3)*(env.pn-env.p2) + env.susamp
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
