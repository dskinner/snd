package main

import (
	"log"
	"sync"
	"time"

	"dasa.cc/piano/snd"
	"dasa.cc/piano/snd/al"

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

	ms = time.Millisecond

	harm  = snd.Sine()
	notes = snd.EqualTempermant(12, 440, 48)

	adsrs [12]*snd.ADSR

	piano    *Piano
	pianomod *snd.ADSR
	pianowf  *snd.Waveform

	someosc *snd.Osc
	somemod *snd.Osc
	somewf  *snd.Waveform
)

func onStart(ctx gl.Context) {
	ctx.Enable(gl.DEPTH_TEST)
	ctx.Enable(gl.BLEND)
	ctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	if err := al.OpenDevice(buffers); err != nil {
		log.Fatal(err)
	}

	mix := snd.NewMixer()

	for i := range adsrs {
		// notes[51] is Major C
		osc := snd.NewOsc(harm, notes[51+i], snd.NewOsc(harm, 2, nil))
		cmb := snd.NewComb(40*ms, 0.8, osc)
		adsr := snd.NewADSR(400*ms, 150*ms, 400*ms, 350*ms, 0.7, 1, cmb)
		adsr.SetEnabled(false)
		adsr.SetLoop(false)
		mix.Append(adsr)
		adsrs[i] = adsr
	}

	somemod = snd.NewOsc(harm, 2, nil)
	someosc = snd.NewOsc(harm, 0, somemod)
	mix.Append(someosc)

	piano = NewPiano()
	pianobuf := snd.NewBuffer(4, piano)
	mix.AppendQuiet(pianobuf)

	mixbuf := snd.NewBuffer(4, snd.NewFilter(1800, 250, mix))

	pan := snd.NewPan(0, mixbuf)
	al.AddSource(pan)

	var err error
	pianowf, err = snd.NewWaveform(ctx, pianobuf)
	if err != nil {
		log.Fatal(err)
	}
	// pianowf.Align(-0.999)

	somewf, err = snd.NewWaveform(ctx, mixbuf)
	if err != nil {
		log.Fatal(err)
	}
}

func onStop() {
	al.CloseDevice()
}

func onPaint(ctx gl.Context) {
	ctx.ClearColor(0, 0, 0, 1)
	ctx.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	pianowf.Paint(ctx, -1, -1, 2, 1)
	somewf.Paint(ctx, -1, 0.25, 2, 0.5)
}

var (
	lastFreq float64 = 440
	lastY    float32

	touchseq = make(map[touch.Sequence]int)

	whackypiano sync.Once
)

func onTouch(ev touch.Event) {
	idx := piano.KeyAt(ev, sz)
	if idx == -1 {
		// top half
		switch ev.Type {
		case touch.TypeBegin:
			lastY = ev.Y
		case touch.TypeMove:
			dt := (ev.Y - lastY)
			lastY = ev.Y
			freq := lastFreq - float64(dt)
			if freq > 20000 {
				freq = 20000
			}
			if freq < 0 {
				freq = 0
			}
			someosc.SetFreq(freq, nil) // somemod
			lastFreq = freq
		}
		return
	}

	switch ev.Type {
	case touch.TypeBegin:
		adsrs[idx].Sustain()
		adsrs[idx].Restart()
		adsrs[idx].SetEnabled(true)
		touchseq[ev.Sequence] = idx
	case touch.TypeMove:

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

	if len(touchseq) == 4 {
		whackypiano.Do(func() {
			pianomod = snd.NewADSR(
				200*time.Millisecond,
				50*time.Millisecond,
				400*time.Millisecond,
				350*time.Millisecond,
				0.7, 1,
				snd.NewOsc(harm, 44, snd.NewOsc(harm, 88, nil)))
			piano.SetAmp(1, pianomod)
			go func() {
				time.Sleep(5 * time.Second)
				pianomod.SetLoop(false)
				pianomod.Release()
				time.Sleep(200 * time.Millisecond)
				piano.SetAmp(1, nil)
				whackypiano = sync.Once{}
			}()
		})
	}
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
				}
			case touch.Event:
				onTouch(ev)
			case size.Event:
				sz = ev
			case paint.Event:
				onPaint(glctx)
				al.Tick()
				a.Publish()
				if visible {
					a.Send(paint.Event{})
				}
			}
		}
	})
}
