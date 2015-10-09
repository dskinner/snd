package snd

type UnitMode int

const (
	UnitSample UnitMode = iota
	UnitStep
	UnitRamp
)

type Unit struct {
	*snd
	amp  float64
	step float64
	mode UnitMode
}

func NewUnit(amp float64, step float64, mode UnitMode) *Unit {
	return &Unit{
		snd:  newSnd(nil),
		amp:  amp,
		step: step,
		mode: mode,
	}
}

func (u *Unit) Prepare() {
	u.snd.Prepare()
	for i := range u.out {
		if u.enabled {
			u.out[i] = u.amp
			switch u.mode {
			case UnitSample:
				u.enabled = false
			case UnitRamp:
				u.out[i] += u.step
			}
		} else {
			u.out[i] = 0
		}
	}
}
