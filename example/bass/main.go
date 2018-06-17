package main

import (
	"log"
	"math"
	"time"

	"dasa.cc/snd"
	"dasa.cc/snd/al"
)

var (
	notes = snd.EqualTempermant(12, 440, 48)

	freq0, freq1 float64
	freq2, freq3 float64

	freq4, freq5 float64
	freq6, freq7 float64
)

func main() {
	master := snd.NewMixer()
	const buffers = 1
	if err := al.OpenDevice(buffers); err != nil {
		log.Fatal(err)
	}
	al.Start(master)

	sine := snd.Sawtooth()

	freq0 = notes[15] * math.Pow(2, 30.0/1200)
	freq2 = notes[15] * math.Pow(2, -30.0/1200)

	freq1 = notes[7] * math.Pow(2, 30.0/1200)
	freq3 = notes[7] * math.Pow(2, -30.0/1200)

	osc0 := snd.NewOscil(sine, freq0, nil)
	osc0.SetAmp(snd.Decibel(-12).Amp(), nil)
	osc1 := snd.NewOscil(sine, freq2, nil)
	osc1.SetAmp(snd.Decibel(-12).Amp(), nil)

	freq4 = notes[27] * math.Pow(2, 10.0/1200)
	freq6 = notes[27] * math.Pow(2, -10.0/1200)

	freq5 = notes[19] * math.Pow(2, 10.0/1200)
	freq7 = notes[19] * math.Pow(2, -10.0/1200)

	osc2 := snd.NewOscil(sine, freq4, nil)
	osc2.SetAmp(snd.Decibel(-15).Amp(), nil)
	osc3 := snd.NewOscil(sine, freq6, nil)
	osc3.SetAmp(snd.Decibel(-15).Amp(), nil)

	go func() {
		for range time.Tick(750 * time.Millisecond) {
			freq0, freq1 = freq1, freq0
			freq2, freq3 = freq3, freq2
			osc0.SetFreq(freq0, nil)
			osc1.SetFreq(freq2, nil)

			freq4, freq5 = freq5, freq4
			freq6, freq7 = freq7, freq6
			osc2.SetFreq(freq4, nil)
			osc3.SetFreq(freq6, nil)
		}
	}()

	mix := snd.NewMixer(osc0, osc1, osc2, osc3)

	adsr := snd.NewADSR(50*time.Millisecond, 450*time.Millisecond, 200*time.Millisecond, 50*time.Millisecond, 0.5, 1, mix)
	master.Append(adsr)

	al.Notify()

	for range time.Tick(time.Second) {
		log.Printf("underruns=%-4v buflen=%-4v tickavg=%-12s drift=%s\n",
			al.Underruns(), al.BufLen(), al.TickAverge(), al.DriftApprox())
	}
}
