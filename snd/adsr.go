package snd

import "time"

type ADSR struct {
	*mono

	// time taken for signal to change from 0 to maxamp.
	attack float64
	// time taken for signal to change from maxamp to susamp.
	decay float64
	// time spent locking signal at susamp.
	sustain float64
	// time taken after sustain period for signal to change from susamp to 0.
	release float64

	// sustain amplitude multiplier
	susamp float64
	// max amplitutde multiplier after rise
	maxamp float64

	// sustaining locks envelope by pausing count during sustain period.
	sustaining bool

	// track time in samples
	count float64
	// envelope duration
	duration float64
}

func NewADSR(attack, decay, sustain, release time.Duration, susamp, maxamp float64, in Sound) *ADSR {
	sr := DefaultSampleRate
	if in != nil {
		sr = in.SampleRate()
	}
	return &ADSR{
		mono:     newmono(in),
		attack:   sr * float64(attack) / float64(time.Second),
		decay:    sr * float64(decay) / float64(time.Second),
		sustain:  sr * float64(sustain) / float64(time.Second),
		release:  sr * float64(release) / float64(time.Second),
		duration: sr * float64(attack+decay+sustain+release) / float64(time.Second),
		susamp:   susamp,
		maxamp:   maxamp,
	}
}

// Sustain locks envelope when sustain period is reached.
func (env *ADSR) Sustain() { env.sustaining = true }

// Release immediately releases envelope from anywhere and starts release period.
// TODO don't enter release if already in release
func (env *ADSR) Release() bool {
	env.sustaining = false
	if env.count < (env.duration - env.release) {
		env.count = env.duration - env.release + 1
		return true
	}
	return false
}

// Restart resets envelope to start from attack period.
func (env *ADSR) Restart() {
	env.count = 0
}

// TODO timing attack appears to be off. The value tested for is 200ms but result is 168ms.
// var (
// t          time.Time
// printStart sync.Once
// printEnd   sync.Once
// )

func (env *ADSR) Prepare(tc uint64) {
	if env.tc == tc {
		return
	}
	env.tc = tc

	for i := range env.out {
		if env.off {
			if env.in != nil {
				env.in.Prepare(tc)
			}

			amp := 1.0
			if env.count == env.duration {
				env.count = 0
			}
			if env.count < env.attack {
				amp = env.maxamp / env.attack * env.count
				// printStart.Do(func() {
				// t = time.Now()
				// log.Println("attack start", env.amp)
				// })
			}
			if env.count >= env.attack && env.count < (env.attack+env.decay) {
				// printEnd.Do(func() {
				// log.Println("attack finished", time.Now().Sub(t))
				// })
				amp = ((env.susamp-env.maxamp)/env.decay)*(env.count-env.attack) + env.maxamp
			}
			if env.count >= (env.attack+env.decay) && env.count <= (env.duration-env.release) {
				amp = env.susamp
			}
			if env.count > (env.duration - env.release) {
				if env.sustaining {
					amp = env.susamp
					env.count--
				} else {
					amp = ((0-env.susamp)/env.release)*(env.count-(env.duration-env.release)) + env.susamp
				}
			}
			env.count++
			if env.in == nil {
				env.out[i] = amp
			} else {
				env.out[i] = amp * env.in.Samples()[i]
			}
		} else {
			env.out[i] = 0
		}
	}
}
