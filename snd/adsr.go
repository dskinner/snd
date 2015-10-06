package snd

import "time"

type ADSR struct {
	*snd

	// time taken for signal to change from 0 to maxamp.
	attack float64
	// time taken for signal to change from maxamp to sustain.
	decay float64
	// time taken after sustain period for signal to change from sustain to 0.
	release float64

	// sustain amplitude multiplier
	sustain float64
	// max amplitutde multiplier after rise
	maxamp float64

	// should sustain
	sustaining bool

	// track time in samples
	count float64
	// envelope duration
	duration float64
}

func NewADSR(attack, decay, release, duration time.Duration, sustain, maxamp float64, in Sound) *ADSR {
	return &ADSR{
		snd:      newSnd(in),
		attack:   DefaultSampleRate * float64(attack) / float64(time.Second),
		decay:    DefaultSampleRate * float64(decay) / float64(time.Second),
		release:  DefaultSampleRate * float64(release) / float64(time.Second),
		duration: DefaultSampleRate * float64(duration) / float64(time.Second),
		sustain:  sustain,
		maxamp:   maxamp,
	}
}

// Sustain locks envelope when sustain phase is reached.
func (env *ADSR) Sustain() { env.sustaining = true }

// Release immediately releases envelope from anywhere and starts release phase.
func (env *ADSR) Release() {
	env.sustaining = false
	env.count = env.duration - env.release + 1
}

// Restart resets envelope to start from attach phase.
func (env *ADSR) Restart() {
	env.count = 0
}

func (env *ADSR) Prepare() {
	env.snd.Prepare()

	for i, x := range env.in.Output() {
		if env.count == env.duration {
			env.count = 0
		}
		if env.count < env.attack {
			env.amp = env.maxamp / env.attack * env.count
		}
		if env.count >= env.attack && env.count < (env.attack+env.decay) {
			env.amp = ((env.sustain-env.maxamp)/env.decay)*(env.count-env.attack) + env.maxamp
		}
		if env.count >= (env.attack+env.decay) && env.count <= (env.duration-env.release) {
			env.amp = env.sustain
		}
		if env.count > (env.duration - env.release) {
			if env.sustaining {
				env.amp = env.sustain
				env.count--
			} else {
				env.amp = ((0-env.sustain)/env.release)*(env.count-(env.duration-env.release)) + env.sustain
			}
		}
		env.count++
		env.out[i] = env.amp * x
	}
}
