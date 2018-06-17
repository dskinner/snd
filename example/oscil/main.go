package main

import (
	"log"
	"time"

	"dasa.cc/snd"
	"dasa.cc/snd/al"
)

// sn returns t seconds of sine wave with n periods at sample rate r.
// For example, 100 periods played back at 44100Hz produces a 441Hz sine wave.
// r must be evenly divisibable by n for result to be periodic.
// The only purpose for producing seconds of audio is to dispel any illusory
// thought that this changes how resampling works later.
func sn(n, r, t int) snd.Discrete {
	xs := make(snd.Discrete, r*t)
	xs.Sample(snd.SineFunc, float64(n)/float64(r), 0)
	return xs
}

func main() {
	master := snd.NewMixer()
	const buffers = 1
	if err := al.OpenDevice(buffers); err != nil {
		log.Fatal(err)
	}

	const n = 1024
	ms := make(snd.Discrete, n)
	ms.Sample(snd.SineFunc, 1.0/n, 0)
	mod := snd.NewOscil(ms, 0.25, nil)
	_ = mod

	sine := make(snd.Discrete, n)
	sine.Sample(snd.SineFunc, 1.0/n, 0)

	// some examples
	// var tmp snd.Discrete
	// _ = tmp

	// downsamples sine into tmp, halving resolution
	// tmp = make(snd.Discrete, n/2)
	// tmp.Sample(sine.Index, n/2, 0)
	// sine = tmp

	// upsamples sine (from prev) into tmp, doubling resolution
	// tmp = make(snd.Discrete, n)
	// tmp.Sample(sine.Interpolate, n, 0)
	// sine = tmp

	// samples half of sine into tmp (makes for interesting effect)
	// tmp = make(snd.Discrete, n/2)
	// tmp.Sample(sine.Index, n, 0)
	// sine = tmp

	// produce t seconds of 441Hz sine wave at different sampling rates
	// and use oscillator as clock for playback. All of these are
	// resampled on the fly during playback.
	const t = 4
	// osc := snd.NewOscil(sn(441, 44100, int(t)), 1./t, nil)

	//
	// dat := make(snd.Discrete, 256)
	// var phase float64
	// for i := 0; i < len(dat); i++ {
	// phase, dat[i] = snd.Err(i, phase)
	// fmt.Println(dat[i])
	// }
	// dat.Normalize()
	// return

	//
	// for i := 1; i < 44100; i++ {
	// osc.Prepare(uint64(i))
	// }
	// var dat snd.Discrete
	// for _, x := range snd.Phasedata {
	// dat = append(dat, x-256)
	// }
	// dat.Normalize()

	//
	// dat2 := make(snd.Discrete, 1024)
	// dat2.Sample(dat.Index, 1024, 0)

	// mod = snd.NewOscil(dat2, 0.1, nil)
	// osc = snd.NewOscil(ms, 440, nil)
	// osc.SetAmp(1, mod)
	// return

	// osc := snd.NewOscil(sn(50, 22050, int(t)), 1./t, nil)
	// osc := snd.NewOscil(sn(25, 11025, int(t)), 1./t, nil)

	// var a snd.Discrete = sn(25, 11025, t) // 4 seconds 441Hz sine @ 11025Hz
	// b := make(snd.Discrete, 44100*t)      // space for 4 seconds @ 44100Hz
	// b.Sample(a.Interpolate, 44100*t, 0)   // upsample 441Hz sine @ 11025Hz to 44100Hz
	// osc := snd.NewOscil(b, 1./t, nil)     // use osc as clock for looping playback

	osc := snd.NewOscil(sine, 440, mod)
	osc.SetAmp(0.5, nil)
	master.Append(osc)
	// osc2 := snd.NewOsc(sine, 0.5, 440)
	// master.Append(osc2)
	al.Start(master)
	for range time.Tick(time.Second) {
		log.Printf("underruns=%-4v buflen=%-4v tickavg=%-12s drift=%s\n",
			al.Underruns(), al.BufLen(), al.TickAverge(), al.DriftApprox())
	}
}
