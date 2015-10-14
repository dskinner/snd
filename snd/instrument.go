package snd

import "time"

type Instrument struct {
	*mono
	// time in samples
	count float64
	offat float64
}

func NewInstrument(in Sound) *Instrument {
	return &Instrument{mono: newmono(in)}
}

func (inst *Instrument) OffIn(d time.Duration) {
	inst.offat = inst.count + (inst.SampleRate() * float64(d) / float64(time.Second))
}

func (inst *Instrument) On() {
	inst.offat = -1 // cancels any previous offat if not reached
	inst.mono.On()
}

func (inst *Instrument) Prepare(uint64) {
	for i := range inst.out {
		if inst.off {
			inst.out[i] = 0
		} else {
			// if inst.in != nil {
			// inst.in.Prepare(tc)
			// }
			inst.out[i] = inst.in.Samples()[i]
		}

		inst.count++ // TODO overflows at some point i suppose
		if inst.count == inst.offat {
			inst.Off()
		}
	}
}
