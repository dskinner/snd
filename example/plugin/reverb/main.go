package main

import "C"

import (
	"time"

	"dasa.cc/snd"
)

func Reverb(in snd.Sound) snd.Sound {
	dly := snd.NewDelay(29*time.Millisecond, in)
	tap0 := snd.NewTap(19*time.Millisecond, dly)
	tap1 := snd.NewTap(13*time.Millisecond, dly)
	tap2 := snd.NewTap(7*time.Millisecond, dly)
	cmb := snd.NewComb(snd.Decibel(-3).Amp(), 11*time.Millisecond, snd.NewMixer(dly, tap0, tap1, tap2))
	return snd.NewLowPass(2000, cmb)
}
