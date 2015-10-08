package snd

import "time"

type ADSR struct {
	*snd

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

	// TODO get rid of this ...
	noloop bool
}

func NewADSR(attack, decay, sustain, release time.Duration, susamp, maxamp float64, in Sound) *ADSR {
	env := &ADSR{
		snd:     newSnd(in),
		attack:  DefaultSampleRate * float64(attack) / float64(time.Second),
		decay:   DefaultSampleRate * float64(decay) / float64(time.Second),
		sustain: DefaultSampleRate * float64(sustain) / float64(time.Second),
		release: DefaultSampleRate * float64(release) / float64(time.Second),
		// duration: DefaultSampleRate * float64(duration) / float64(time.Second),
		susamp: susamp,
		maxamp: maxamp,
	}
	env.duration = env.attack + env.decay + env.sustain + env.release
	return env
}

// Sustain locks envelope when sustain period is reached.
func (env *ADSR) Sustain() { env.sustaining = true }

// Release immediately releases envelope from anywhere and starts release period.
func (env *ADSR) Release() {
	env.sustaining = false
	env.count = env.duration - env.release + 1
}

// Restart resets envelope to start from attack period.
func (env *ADSR) Restart() {
	env.count = 0
}

func (env *ADSR) SetLoop(b bool) {
	env.noloop = !b
}

// TODO timing attack appears to be off. The value tested for is 200ms but result is 168ms.
// var (
// t          time.Time
// printStart sync.Once
// printEnd   sync.Once
// )

func (env *ADSR) Prepare() {
	env.snd.Prepare()

	for i, x := range env.in.Output() {
		if env.enabled {
			if env.count == env.duration {
				env.count = 0
				if env.noloop {
					env.SetEnabled(false)
					return
				}
			}
			if env.count < env.attack {
				env.amp = env.maxamp / env.attack * env.count
				// printStart.Do(func() {
				// t = time.Now()
				// log.Println("attack start", env.amp)
				// })
			}
			if env.count >= env.attack && env.count < (env.attack+env.decay) {
				// printEnd.Do(func() {
				// log.Println("attack finished", time.Now().Sub(t))
				// })
				env.amp = ((env.susamp-env.maxamp)/env.decay)*(env.count-env.attack) + env.maxamp
			}
			if env.count >= (env.attack+env.decay) && env.count <= (env.duration-env.release) {
				env.amp = env.susamp
			}
			if env.count > (env.duration - env.release) {
				if env.sustaining {
					env.amp = env.susamp
					env.count--
				} else {
					env.amp = ((0-env.susamp)/env.release)*(env.count-(env.duration-env.release)) + env.susamp
				}
			}
			env.count++
			env.out[i] = env.amp * x
		} else {
			env.out[i] = 0
		}
	}
}
