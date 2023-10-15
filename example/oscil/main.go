package main

import (
	"fmt"
	"os"

	"dasa.cc/signal"
	"dasa.cc/snd"
)

var (
	Decibel = func(dB float64) float64 { return snd.Decibel(-15).Amp() } // TODO maybe make New* constructors take decibel instead of amplitude
	Gain    = snd.NewGain
	Mixer   = snd.NewMixer
	Oscil   = snd.NewOscil
	Sine    = signal.Sine()
)

func main() {
	mix := Mixer(
		Gain(Decibel(-15),
			Oscil(Sine, 220,
				Oscil(Sine, 2, nil))))

	pl := snd.NewPlayer(mix)
	if err := pl.Start(); err != nil {
		fmt.Printf("Failed to start: %v", err)
		os.Exit(1)
	}
	fmt.Println("Press Enter to quit...")
	fmt.Scanln()
	pl.Stop()
}
