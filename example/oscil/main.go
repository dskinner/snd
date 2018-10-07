package main

import (
	"log"
	"time"

	"dasa.cc/snd"
	"dasa.cc/snd/al"
)

func main() {
	master := snd.NewMixer()
	const buffers = 1
	if err := al.OpenDevice(buffers); err != nil {
		log.Fatal(err)
	}
	al.Start(master)

	sine := snd.Sine()
	// // mod is a modulator; try replacing the nil argument to the oscillator with this.
	// // mod := snd.NewOscil(sine, 2, nil)
	osc := snd.NewOscil(sine, 440, nil) // oscillator

	master.Append(osc)
	al.Notify()

	for range time.Tick(time.Second) {
		log.Printf("underruns=%-4v buflen=%-4v tickavg=%-12s drift=%s\n",
			al.Underruns(), al.BufLen(), al.TickAverge(), al.DriftApprox())
	}
}
