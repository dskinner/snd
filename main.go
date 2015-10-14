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

// TODO organ ?
// https://en.wikipedia.org/wiki/Synthesizer
// http://www.soundonsound.com/sos/jan04/articles/synthsecrets.htm
// TODO piano
// http://www.soundonsound.com/sos/nov02/articles/synthsecrets1102.asp

const buffers = 2

var (
	sz size.Event

	fps       int
	lastpaint = time.Now()

	ms = time.Millisecond

	sawtooth = snd.Sawtooth(4)
	square   = snd.Square(4)
	pulse    = snd.Pulse(4)
	sine     = snd.Sine()

	harm  = snd.Square(4)
	notes = snd.EqualTempermant(12, 440, 48)

	keys [12]*Key

	piano    *Piano
	pianomod *snd.ADSR
	pianowf  *snd.Waveform

	loop snd.Looper

	mix   *snd.Mixer
	mixwf *snd.Waveform

	somemod snd.Oscillator
	someosc snd.Oscillator
)

type Key struct {
	*snd.Instrument
	adsr    *snd.ADSR
	release time.Duration
	total   time.Duration

	freq float64
}

func NewKeyOsc(freq float64) *Key {
	osc := snd.Osc(harm, freq, nil)
	instr := snd.NewInstrument(osc)
	instr.Off()

	return &Key{instr, nil, 0, 0, freq}
}

func NewKey(freq float64) *Key {
	// adsr release and also offin value for instrument
	release := 350 * ms

	// TODO http://www.soundonsound.com/sos/nov02/articles/synthsecrets1102.asp
	osc0 := snd.Osc(sawtooth, freq, snd.Osc(sine, 2, nil))
	osc0.SetPhase(1, snd.Osc(square, freq*0.4, nil))
	comb := snd.NewComb(0.8, 10*ms, osc0)
	adsr := snd.NewADSR(50*ms, 500*ms, 100*ms, release, 0.4, 1, comb)

	instr := snd.NewInstrument(adsr)
	instr.Off()

	return &Key{
		Instrument: instr,
		adsr:       adsr,
		release:    release,
		total:      1000 * ms,
		freq:       freq,
	}
}

func (key *Key) Press() {
	key.On()
	key.OffIn(key.total)
	if key.adsr != nil {
		// key.adsr.Sustain()
		key.adsr.Restart()
	}
}

func (key *Key) Release() {
	if key.adsr != nil {
		if key.adsr.Release() {
			key.OffIn(key.release)
		}
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

	mix = snd.NewMixer()
	for i := range keys {
		keys[i] = NewKey(notes[51+i]) // notes[51] is Major C
		mix.Append(keys[i])
	}

	loop = snd.Loop(5*time.Second, mix)
	mixloop := snd.NewMixer(mix, loop)

	lp := snd.NewLowPass(1500, mixloop)

	mixwf, err = snd.NewWaveform(ctx, 4, lp)
	if err != nil {
		log.Fatal(err)
	}

	pan := snd.NewPan(0, mixwf)
	al.AddSource(pan)

	// piano graphics

	// pianomod2 := snd.Osc(harm, 88, nil)
	// pianomod2.SetAmp(1, nil)
	// pianomod2.SetBufferLen(1024)

	// pianomod1 := snd.Osc(harm, 44, pianomod2)
	// pianomod1.SetAmp(1, nil)
	// pianomod1.SetBufferLen(1024)

	// pianomod = snd.NewADSR(
	// 200*time.Millisecond,
	// 50*time.Millisecond,
	// 400*time.Millisecond,
	// 350*time.Millisecond,
	// 0.7, 1,
	// pianomod1)
	// pianomod.SetAmp(1, nil)
	// pianomod.SetBufferLen(1024)

	piano = NewPiano()
	// piano.Sound.SetAmp(1, pianomod)

	pianowf, err = snd.NewWaveform(ctx, 1, piano)
	if err != nil {
		log.Fatal(err)
	}
	// pianowf.Align(-0.999)

	// experimental bits
	somemod = snd.Osc(harm, 2, nil)
	// interesting frequencies with first key
	// 520, 695
	someosc = snd.Osc(harm, 695, nil)
	// mix.Append(someosc)

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

	pianowf.Prepare(1)
	switch sz.Orientation {
	case size.OrientationPortrait:
		pianowf.Paint(ctx, -1, -1, 2, 0.5)
		mixwf.Paint(ctx, -1, 0.25, 2, 0.5)
	default:
		pianowf.Paint(ctx, -1, -1, 2, 1)
		mixwf.Paint(ctx, -1, 0.25, 2, 0.5)
	}

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
			// TODO needs ui control
			loop.Record()
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
			log.Println("freq:", freq)
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

	// mult := 1.0
	// if x := len(touchseq); x != 0 {
	// mult /= float64(x + 1) // +1 for someosc
	// }
	// gain.SetMult(mult)

	// if len(touchseq) == 4 {
	// whackypiano.Do(func() {
	// piano.SetAmp(1, pianomod)
	// go func() {
	// time.Sleep(5 * time.Second)
	// pianomod.Release()
	// time.Sleep(200 * time.Millisecond)
	// piano.SetAmp(1, nil)
	// whackypiano = sync.Once{}
	// }()
	// })
	// }
}

func main() {
	app.Main(func(a app.App) {

		logdbg := time.NewTicker(time.Second)
		go func() {
			for range logdbg.C {
				log.Printf("fps=%-4v underruns=%-4v buflen=%-4v tickavg=%-12s drift=%s\n",
					fps, al.Underruns(), al.BufLen(), al.TickAverge(), al.DriftApprox())
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
				onPaint(glctx)
				a.Publish()
				if visible {
					a.Send(paint.Event{})
				}
			}
		}
	})
}
