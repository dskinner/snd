package main

import (
	"fmt"
	"io"
	"os"

	"dasa.cc/signal"
	"dasa.cc/snd"
	"github.com/gen2brain/malgo"
	// "dasa.cc/snd/al"
	// "dasa.cc/snd/miniaudio"
)

func start(pl *snd.Player) func() {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("LOG %v", message)
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// defer func() {
	// 	_ = ctx.Uninit()
	// 	ctx.Free()
	// }()

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)
	deviceConfig.Playback.Format = malgo.FormatF32
	deviceConfig.Playback.Channels = pl.Channels()
	deviceConfig.SampleRate = pl.SampleRate()
	deviceConfig.Alsa.NoMMap = 1

	// This is the function that's used for sending more data to the device for playback.
	onSamples := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		io.ReadFull(pl, pOutputSample)
	}

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: onSamples,
	}
	device, err := malgo.InitDevice(ctx.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// defer device.Uninit()

	err = device.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return func() {
		device.Uninit()
		_ = ctx.Uninit()
		ctx.Free()
	}
}

func main() {
	master := snd.NewMixer()
	sine := signal.Sine()
	mod := snd.NewOscil(sine, 2, nil)
	osc := snd.NewOscil(sine, 220, mod)
	master.Append(snd.NewGain(snd.Decibel(-15).Amp(), osc))
	pl := snd.NewPlayer(master)

	close := start(pl)
	fmt.Println("Press Enter to quit...")
	fmt.Scanln()
	close()

	// miniaudio.Start(master)

	// if err := al.OpenDevice(1); err != nil {
	// 	log.Fatal(err)
	// }
	// al.Start(master)
	// for range time.Tick(time.Second) {
	// 	log.Printf("underruns=%-4v buflen=%-4v tickavg=%-12s drift=%s\n",
	// 		al.Underruns(), al.BufLen(), al.TickAverge(), al.DriftApprox())
	// }
}
