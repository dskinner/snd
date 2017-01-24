package main

import (
	"log"
	"time"

	"dasa.cc/snd"
	"dasa.cc/snd/al"
)

var (
	notes    = snd.EqualTempermant(12, 440, 48)
	sawsine  = snd.SawtoothSynthesis(8)
	triangle = snd.Triangle()
	bpm      = snd.BPM(80)
)

func rhythm(freq float64) snd.Sound {
	osc := snd.NewOscil(sawsine, freq, snd.NewOscil(triangle, 4, nil))
	dmp0 := snd.NewDamp(bpm.Dur(), osc)
	quin := snd.BPM(float64(bpm) * 1.25).Dur()
	dmp1 := snd.NewDamp(quin, osc)
	drv := snd.NewDrive(quin, osc)
	mix := snd.NewMixer(dmp0, dmp1, drv)
	adsr := snd.NewADSR(250*time.Millisecond, 500*time.Millisecond, 300*time.Millisecond, 400*time.Millisecond, 0.85, 1.0, mix)
	adsr.Sustain()
	return adsr
}

func main() {
	master := snd.NewMixer()
	gain := snd.NewGain(snd.Decibel(-6).Amp(), master)
	const buffers = 1
	if err := al.OpenDevice(buffers); err != nil {
		log.Fatal(err)
	}
	al.Start(gain)

	mix := snd.NewMixer(rhythm(notes[51]), rhythm(notes[58]), rhythm(notes[60])) // C5 G5 A5
	lowpass := snd.NewLowPass(773, mix)
	mixgain := snd.NewGain(snd.Decibel(-3).Amp(), lowpass)

	dly := snd.NewDelay(29*time.Millisecond, mixgain)
	tap0 := snd.NewTap(19*time.Millisecond, dly)
	tap1 := snd.NewTap(13*time.Millisecond, dly)
	tap2 := snd.NewTap(7*time.Millisecond, dly)
	cmb := snd.NewComb(snd.Decibel(-3).Amp(), 11*time.Millisecond, snd.NewMixer(dly, tap0, tap1, tap2))
	reverb := snd.NewLowPass(2000, cmb)

	master.Append(mixgain, reverb)
	al.Notify()

	for range time.Tick(time.Second) {
		log.Printf("underruns=%-4v buflen=%-4v tickavg=%-12s drift=%s\n",
			al.Underruns(), al.BufLen(), al.TickAverge(), al.DriftApprox())
	}
}
