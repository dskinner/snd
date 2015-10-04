package main

import (
	"log"

	"dasa.cc/piano/snd"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"
)

const buffers = 8

var (
	sz size.Event

	oal = snd.NewOpenAL()
)

func onStart(ctx gl.Context) {
	if err := oal.OpenDevice(buffers); err != nil {
		log.Fatal(err)
	}

	harm := snd.Sine()
	osc0 := snd.NewOsc(harm, 523.251)
	osc1 := snd.NewOsc(harm, 659.255)
	osc2 := snd.NewOsc(harm, 783.991)
	mix := snd.NewMixer(osc0, osc1, osc2)
	pan := snd.NewPan(0, mix)

	oal.SetInput(pan)
	// oal.SetInput(snd.Slice{osc0, osc1, osc2, mix, pan})
	oal.Play()
}

func onStop() {
	oal.CloseDevice()
}

func onTouch(ev touch.Event) {
	log.Printf("touch.Event: %#v\n", ev)
}

func onPaint(ctx gl.Context) {
	ctx.ClearColor(0, 0, 0, 1)
	ctx.Clear(gl.COLOR_BUFFER_BIT)
}

func main() {
	app.Main(func(a app.App) {
		var (
			glctx   gl.Context
			visible bool
		)
		for ev := range a.Events() {
			switch ev := a.Filter(ev).(type) {
			case lifecycle.Event:
				switch ev.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					visible = true
					glctx = ev.DrawContext.(gl.Context)
					onStart(glctx)
				case lifecycle.CrossOff:
					visible = false
					onStop()
					oal.Destroy()
				}
			case touch.Event:
				onTouch(ev)
			case size.Event:
				sz = ev
			case paint.Event:
				onPaint(glctx)
				a.Publish()
				if visible {
					a.Send(paint.Event{})
				}
			}
		}
	})
}
