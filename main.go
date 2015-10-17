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

	sawtooth = snd.Sawtooth(512)
	square   = snd.Square(512)
	pulse    = snd.Pulse(4)
	sine     = snd.Sine()

	harm  = snd.Square(4)
	notes = snd.EqualTempermant(12, 440, 48)

	keys [12]*Key

	piano    *Piano
	pianomod *snd.ADSR
	pianowf  *snd.Waveform

	loop *snd.Loop

	mix   *snd.Mixer
	mixwf *snd.Waveform

	somemod snd.Oscillator
	someosc snd.Oscillator

	// 0.42659999999999976
	// 0.4275
	// 0.50429
	// 0.2488
	phasefac float64 = 0.4275
)

type Key struct {
	*snd.Instrument
	osc     *snd.Oscil
	adsr0   *snd.ADSR
	adsr1   *snd.ADSR
	release time.Duration

	freq float64
}

// func NewKeyOsc(freq float64) *Key {
// osc := snd.NewOscil(harm, freq, nil)
// nst := snd.NewInstrument(osc)
// nst.Off()
// return &Key{nst, osc, nil, 0, freq}
// }

func NewKey(idx int) *Key {
	freq := notes[idx]
	c := 15.0
	freq1 := freq - ((notes[idx] - notes[idx-1]) / 100 * c)
	freq2 := freq + ((notes[idx+1] - notes[idx]) / 100 * c)
	// adsr release and also offin value for instrument
	release := 1350 * ms

	// TODO http://www.soundonsound.com/sos/nov02/articles/synthsecrets1102.asp
	osc0 := snd.NewOscil(sawtooth, freq, nil) // snd.NewOscil(sine, 2, nil))
	osc0.SetPhase(1, snd.NewOscil(square, freq*phasefac, nil))
	osc1 := snd.NewOscil(sawtooth, freq1, nil)
	osc1.SetPhase(1, snd.NewOscil(square, freq1*phasefac, nil))
	osc2 := snd.NewOscil(sawtooth, freq2, nil)
	osc2.SetPhase(1, snd.NewOscil(square, freq2*phasefac, nil))
	oscmix := snd.NewMixer(osc0, osc1, osc2)
	// oscgain := snd.NewGain(snd.Decibel(-5).Amp(), oscmix)

	adsr0 := snd.NewADSR(30*ms, 50*ms, 200*ms, release, 0.4, 1, oscmix)
	adsr1 := snd.NewADSR(0*ms, 280*ms, 0*ms, release, 0.4, 1, oscmix)
	adsrmix := snd.NewMixer(adsr0, adsr1)
	adsrgain := snd.NewGain(snd.Decibel(-8).Amp(), adsrmix)
	nst := snd.NewInstrument(adsrgain)
	nst.Off()

	return &Key{
		Instrument: nst,
		osc:        osc0,
		adsr0:      adsr0,
		adsr1:      adsr1,
		release:    release,
		freq:       freq,
	}
}

func (key *Key) Press() {
	key.On()
	key.OffIn(key.adsr0.Dur())
	// if key.adsr != nil {
	// key.adsr.Sustain()
	key.adsr0.Restart()
	key.adsr1.Restart()
	// }
}

func (key *Key) Release() {
	// if key.adsr != nil {
	if key.adsr0.Release() {
		key.adsr1.Release()
		key.OffIn(key.release)
	}
	// } else {
	// key.Off()
	// }
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

	keymix := snd.NewMixer()
	for i := range keys {
		keys[i] = NewKey(51 + i) // notes[51] is Major C
		keymix.Append(keys[i])
	}
	keylp := snd.NewLowPass(773, keymix)

	dly := snd.NewDelay(29*time.Millisecond, keylp)
	tap0 := snd.NewTap(19*time.Millisecond, dly)
	tap1 := snd.NewTap(13*time.Millisecond, dly)
	tap2 := snd.NewTap(7*time.Millisecond, dly)
	cmb := snd.NewComb(snd.Decibel(-3).Amp(), 11*time.Millisecond, snd.NewMixer(dly, tap0, tap1, tap2))
	dlymix := snd.NewMixer(cmb, keylp)
	loop = snd.NewLoop(5*time.Second, dlymix)
	loopmix := snd.NewMixer(dlymix, loop)

	// lp := snd.NewLowPass(800, loopmix)
	master := snd.NewMixer(loopmix)
	mixwf, err = snd.NewWaveform(ctx, 4, master)
	if err != nil {
		log.Fatal(err)
	}

	pan := snd.NewPan(0, mixwf)

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
	// piano.Sound.SetAmp(1, snd.NewDamp(time.Second, nil))

	pianowf, err = snd.NewWaveform(ctx, 1, piano)
	if err != nil {
		log.Fatal(err)
	}
	// pianowf.Align(-0.999)

	// experimental bits
	somemod = snd.NewOscil(harm, 2, nil)
	// interesting frequencies with first key
	// 520, 695
	someosc = snd.NewOscil(harm, 695, nil)

	// mtrosc := snd.NewOscil(harm, 220, nil)
	// mtrdmp := snd.NewDamp(snd.BPM(80).Dur(), mtrosc)
	// mtrdmp1 := snd.NewDamp(snd.BPM(100).Dur(), mtrosc)
	// mtrdrv := snd.NewDrive(snd.BPM(100).Dur(), mtrosc)
	// mtrmix := snd.NewMixer(mtrdmp, mtrdmp1, mtrdrv)
	// master.Append(mtrmix)

	//
	al.AddSource(pan)
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
			// loop.Record()
		case touch.TypeMove:
			dt := (ev.Y - lastY)
			lastY = ev.Y
			pm := phasefac - (float64(dt) * 0.0001)
			if pm <= 0 {
				pm = 0.0001
			}
			log.Println("phasefac", pm)
			phasefac = pm
			for _, key := range keys {
				key.osc.SetPhase(1, snd.NewOscil(square, key.freq*pm, nil))
			}
			al.Notify()
			// freq := lastFreq - float64(dt)
			// if freq > 20000 {
			// freq = 20000
			// }
			// if freq < 0 {
			// freq = 0
			// }
			// someosc.SetFreq(freq, nil) // somemod
			// log.Println("freq:", freq)
			// lastFreq = freq
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
