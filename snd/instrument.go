package snd

import "time"

type Instrument struct {
	*snd

	snds []Sound

	// time in samples
	count float64
	offat float64
}

func NewInstrument(in Sound) *Instrument {
	return &Instrument{
		snd: newSnd(in),
	}
}

func (inst *Instrument) Manage(s Sound) {
	inst.snds = append(inst.snds, s)
}

func (inst *Instrument) Off() {
	inst.enabled = false
	for _, s := range inst.snds {
		s.SetEnabled(false)
	}
}

func (inst *Instrument) OffIn(d time.Duration) {
	inst.offat = inst.count + (inst.SampleRate() * float64(d) / float64(time.Second))
}

func (inst *Instrument) On() {
	inst.offat = -1 // cancels any previous offat if not reached
	inst.enabled = true
	for _, s := range inst.snds {
		s.SetEnabled(true)
	}
}

func (inst *Instrument) Prepare() {
	inst.snd.Prepare()
	for i := range inst.out {
		if inst.enabled {
			inst.out[i] = inst.in.Output()[i]
		} else {
			inst.out[i] = 0
		}
		// TODO overflows at some point i suppose
		inst.count++
		if inst.count == inst.offat {
			inst.Off()
		}
	}
}
