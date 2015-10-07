package main

import (
	"log"
	"time"

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

	osc *snd.Osc
	wfs []*snd.Waveform

	oscs  [12]*snd.Osc
	adsrs [12]*snd.ADSR
)

func onStart(ctx gl.Context) {
	ctx.Enable(gl.DEPTH_TEST)
	ctx.Enable(gl.BLEND)
	ctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	if err := oal.OpenDevice(buffers); err != nil {
		log.Fatal(err)
	}

	harm := snd.Sine()

	oscs[0] = snd.NewOsc(harm, 523.251)
	oscs[1] = snd.NewOsc(harm, 554.365)
	oscs[2] = snd.NewOsc(harm, 587.330)
	oscs[3] = snd.NewOsc(harm, 622.254)
	oscs[4] = snd.NewOsc(harm, 659.255)
	oscs[5] = snd.NewOsc(harm, 698.456)
	oscs[6] = snd.NewOsc(harm, 739.989)
	oscs[7] = snd.NewOsc(harm, 783.991)
	oscs[8] = snd.NewOsc(harm, 830.609)
	oscs[9] = snd.NewOsc(harm, 880)
	oscs[10] = snd.NewOsc(harm, 932.328)
	oscs[11] = snd.NewOsc(harm, 987.767)

	for i, osc := range oscs {
		adsrs[i] = snd.NewADSR(200*time.Millisecond, 50*time.Millisecond, 400*time.Millisecond, 350*time.Millisecond, 0.8, 1, osc)
		adsrs[i].SetEnabled(false)
	}

	mix := snd.NewMixer()
	for _, adsr := range adsrs {
		mix.Append(adsr)
	}

	// osc = snd.NewOsc(harm, 440)
	// adsr := snd.NewADSR(200*time.Millisecond, 50*time.Millisecond, 400*time.Millisecond, 350*time.Millisecond, 0.5, 1, osc)
	// mix := snd.NewMixer(adsr)

	piano := snd.NewPiano()
	bufpiano := snd.NewBuffer(4, piano)
	mix.AppendQuiet(bufpiano)

	buf := snd.NewBuffer(4, mix)

	pan := snd.NewPan(0, buf)
	oal.SetInput(pan)
	oal.Play()

	wf0, err := snd.NewWaveform(ctx, bufpiano)
	if err != nil {
		log.Fatal(err)
	}
	wf0.Align(-0.999)

	wf1, err := snd.NewWaveform(ctx, buf)
	if err != nil {
		log.Fatal(err)
	}

	wfs = append(wfs, wf0, wf1)
}

func onStop() {
	oal.CloseDevice()
}

var touchseq = make(map[touch.Sequence]int)

func onTouch(ev touch.Event) {
	idx := int(ev.X / float32(sz.WidthPx) * 12)
	if idx > 11 {
		idx = 11
	}

	switch ev.Type {
	case touch.TypeBegin:
		adsrs[idx].Sustain()
		adsrs[idx].Restart()
		adsrs[idx].SetEnabled(true)
		touchseq[ev.Sequence] = idx
	case touch.TypeMove:
		// dt := (ev.Y - lastY)
		// lastY = ev.Y
		// freq := osc.Freq() - float64(dt)
		// if freq > 20000 {
		// freq = 20000
		// }
		// if freq < 0 {
		// freq = 0
		// }
		// osc.SetFreq(freq)
		if lastidx, ok := touchseq[ev.Sequence]; ok {
			if idx != lastidx {
				adsrs[lastidx].Release()
				adsrs[idx].Sustain()
				adsrs[idx].Restart()
				adsrs[idx].SetEnabled(true)
				touchseq[ev.Sequence] = idx
			}
		}
	case touch.TypeEnd:
		adsrs[idx].Release()
		delete(touchseq, ev.Sequence)
	default:
	}
}

func onPaint(ctx gl.Context) {
	ctx.ClearColor(0, 0, 0, 1)
	ctx.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	wfs[0].Paint(ctx, -1, 1, -1, 0)
	wfs[1].Paint(ctx, -1, 1, 0, 1)
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
