package snd

// TODO should mixer be stereo out?
// TODO perhaps this class is unnecessary, any sound could be a mixer
// if you can set multiple inputs, but might get confusing.
// TODO consider embedding Gain type
type Mixer struct {
	*mono
	ins []Sound
}

func NewMixer(ins ...Sound) *Mixer   { return &Mixer{newmono(nil), ins} }
func (mix *Mixer) Append(s ...Sound) { mix.ins = append(mix.ins, s...) }
func (mix *Mixer) Empty()            { mix.ins = nil }
func (mix *Mixer) Inputs() []Sound   { return mix.ins }

func (mix *Mixer) Prepare(uint64) {
	for i := range mix.out {
		mix.out[i] = 0
		if !mix.off {
			for _, in := range mix.ins {
				mix.out[i] += in.Sample(i)
			}
		}
	}
}

// TODO don't use this. provides a temporary workaround to spatialization
// effects openal forces onto mono sources. This is exported for dasa.cc/snd/al
// to use until a larger refactoring takes place.
type StereoMixer struct {
	*stereo
	*Buf
}

func NewStereoMixer(in Sound) *StereoMixer {
	return &StereoMixer{newstereo(in), &Buf{Discrete(in.Samples()), 0}}
}

func (mix *StereoMixer) Prepare(uint64) {
	// for i, x := range mix.in.Samples() {
	for i := 0; i < len(mix.out)/2; i++ {
		x := mix.Buf.Read()
		if mix.l.off {
			mix.l.out[i] = 0
		} else {
			mix.l.out[i] = x
		}
		if mix.r.off {
			mix.r.out[i] = 0
		} else {
			mix.r.out[i] = x
		}
		mix.out[i*2] = mix.l.out[i]
		mix.out[i*2+1] = mix.r.out[i]
	}
}
