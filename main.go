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

const buffers = 2

var (
	sz size.Event

	fps       int
	lastpaint = time.Now()

	ms = time.Millisecond

	harm  = snd.Sine()
	notes = snd.EqualTempermant(12, 440, 48)

	keys [12]*Key

	piano    *Piano
	pianobuf *snd.Buffer
	pianomod *snd.ADSR
	pianowf  *snd.Waveform

	mix   *snd.Mixer
	mixwf *snd.Waveform

	someosc snd.Oscillator
	somemod snd.Oscillator
)

type Key struct {
	*snd.Instrument
	adsr    *snd.ADSR
	release time.Duration
}

func NewKeyOsc(freq float64) *Key {
	osc := snd.Osc(harm, freq, nil)
	instr := snd.NewInstrument(osc)
	instr.Manage(osc)
	instr.Off()
	return &Key{instr, nil, 0}
}

func NewKey(freq float64) *Key {
	// adsr release and also offin value for instrument
	release := 350 * ms

	osc := snd.Osc(harm, freq, snd.Osc(harm, 2, nil))
	comb := snd.NewComb(40*ms, 0.8, osc)
	adsr := snd.NewADSR(200*ms, 150*ms, 400*ms, release, 0.7, 1, comb)
	unit := snd.NewUnit(0.5, 0, snd.UnitStep)
	ring := snd.NewRing(unit, adsr)

	instr := snd.NewInstrument(ring)
	instr.Manage(osc)
	instr.Manage(comb)
	instr.Manage(adsr)
	instr.Manage(unit)
	instr.Manage(ring)
	instr.Off()

	key := &Key{
		Instrument: instr,
		adsr:       adsr,
		release:    release,
	}
	return key
}

func (key *Key) Press() {
	if key.adsr != nil {
		key.adsr.Sustain()
		key.adsr.Restart()
	}
	key.On()
}

func (key *Key) Release() {
	if key.adsr != nil {
		key.adsr.Release()
		key.OffIn(key.release)
	} else {
		key.Off()
	}
}

func onStart(ctx gl.Context) {
	ctx.Enable(gl.DEPTH_TEST)
	ctx.Enable(gl.BLEND)
	ctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	if err := al.OpenDevice(buffers); err != nil {
		log.Fatal(err)
	}

	var err error

	// piano key sounds
	mix = snd.NewMixer()
	for i := range keys {
		// notes[51] is Major C
		keys[i] = NewKeyOsc(notes[51+i])
		mix.Append(keys[i])
	}
	mixbuf := snd.NewBuffer(4, snd.NewFilter(1800, 250, mix))
	pan := snd.NewPan(0, mixbuf)
	al.AddSource(pan)
	mixwf, err = snd.NewWaveform(ctx, mixbuf)
	if err != nil {
		log.Fatal(err)
	}

	// piano graphics
	piano = NewPiano()
	pianobuf = snd.NewBuffer(4, piano)
	mix.AppendQuiet(pianobuf)
	pianowf, err = snd.NewWaveform(ctx, pianobuf)
	if err != nil {
		log.Fatal(err)
	}
	// pianowf.Align(-0.999)

	// experimental bits
	somemod = snd.Osc(harm, 2, nil)
	someosc = snd.Osc(harm, 0, nil)
	mix.Append(someosc)
	// somedelay := snd.NewDelay(time.Second, someosc)
	// for i := 1; i < 10; i++ {
	// dur := time.Duration(int64(13*i)) * ms
	// tap := snd.NewTap(dur, somedelay)
	// mix.Append(tap)
	// }
}

func onStop() {
	al.CloseDevice()
}

func onPaint(ctx gl.Context) {
	ctx.ClearColor(0, 0, 0, 1)
	ctx.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	pianowf.Paint(ctx, -1, -1, 2, 1)
	mixwf.Paint(ctx, -1, 0.25, 2, 0.5)

	now := time.Now()
	fps = int(time.Second / now.Sub(lastpaint))
	lastpaint = now
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
		// TODO drag finger off piano and it still plays, shouldn't return here
		// to allow TypeMove to figure out what to turn off
		return
	}

	switch ev.Type {
	case touch.TypeBegin:
		keys[idx].Press()
		touchseq[ev.Sequence] = idx
	case touch.TypeMove:
		// TODO drag finger off piano and it still plays, should stop
		if lastidx, ok := touchseq[ev.Sequence]; ok {
			if idx != lastidx {
				keys[lastidx].Release()
				keys[idx].Press()
				touchseq[ev.Sequence] = idx
			}
		}
	case touch.TypeEnd:
		keys[idx].Release()
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
				snd.Osc(harm, 44, snd.Osc(harm, 88, nil)))
			piano.SetAmp(1, pianomod)
			go func() {
				time.Sleep(5 * time.Second)
				// pianomod.SetLoop(false)
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

		logdbg := time.NewTicker(time.Second)
		go func() {
			for range logdbg.C {
				buflen, underruns := al.BufStats()
				preptime, prepcalls := al.PrepStats()
				prepavg := preptime / time.Duration(prepcalls)
				log.Printf("fps=%-5v underruns=%-6v prepavg=%-15s buflen=%v\n", fps, underruns, prepavg, buflen)
			}
		}()

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
					al.Start()
				case lifecycle.CrossOff:
					visible = false
					al.Stop()
					logdbg.Stop()
					onStop()
				}
			case touch.Event:
				onTouch(ev)
			case size.Event:
				sz = ev
			case paint.Event:
				// al.Tick()
				onPaint(glctx)
				a.Publish()
				if visible {
					a.Send(paint.Event{})
				}
			}
		}
	})
}
