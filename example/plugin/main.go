//go:generate go build -buildmode=plugin ./reverb

package main

import (
	"log"
	"plugin"
	"time"

	"dasa.cc/signal"
	"dasa.cc/snd"
	"dasa.cc/snd/al"
)

// TODO consider a similar type for package snd to compliment type Discrete.
// type ProcFunc func(in snd.Sound) snd.Sound

var reverb func(in snd.Sound) snd.Sound

func init() {
	p, err := plugin.Open("reverb.so")
	if err != nil {
		log.Fatalf("requires go1.8 and linux; try running `go generate` first: %v", err)
	}
	v, err := p.Lookup("Reverb")
	if err != nil {
		log.Fatal(err)
	}
	reverb = v.(func(in snd.Sound) snd.Sound)
}

func main() {
	master := snd.NewMixer()
	gain := snd.NewGain(snd.Decibel(-3).Amp(), master)
	const buffers = 1
	if err := al.OpenDevice(buffers); err != nil {
		log.Fatal(err)
	}
	al.Start(gain)

	dur := snd.BPM(80).Dur()
	mod := snd.NewOscil(signal.Square(), 40, nil)
	osc := snd.NewOscil(signal.Sine(), 440, mod)
	dmp := snd.NewDamp(dur, osc)

	loop := snd.NewLoop(dur, dmp)
	loop.Record()

	master.Append(reverb(loop))
	al.Notify()

	for range time.Tick(time.Second) {
		log.Printf("underruns=%-4v buflen=%-4v tickavg=%-12s drift=%s\n",
			al.Underruns(), al.BufLen(), al.TickAverge(), al.DriftApprox())
	}
}
