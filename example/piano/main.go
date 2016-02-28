package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"dasa.cc/material"
	"dasa.cc/material/icon"
	"dasa.cc/snd"
	"dasa.cc/snd/al"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"
)

const buffers = 1

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

var (
	env     = new(material.Environment)
	toolbar *material.Toolbar
	btnNext *material.FloatingActionButton
	btnkeys [12]*material.Button
	decor   *material.Material
	mixwf   *Waveform

	fps       int
	lastpaint = time.Now()
)

func init() {
	env.SetPalette(material.Palette{
		Primary: material.BlueGrey500,
		Dark:    material.BlueGrey100,
		Light:   material.BlueGrey900,
		Accent:  material.DeepOrangeA200,
	})
}

func onStart(ctx gl.Context) {
	ctx.Enable(gl.BLEND)
	ctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	if err := al.OpenDevice(buffers); err != nil {
		log.Fatal(err)
	}

	env.LoadIcons(ctx)
	env.LoadGlyphs(ctx)

	toolbar = env.NewToolbar(ctx)
	toolbar.SetText("piano")
	toolbar.SetTextColor(material.White)
	toolbar.Nav.SetIconColor(material.White)

	btnLoop := env.NewButton(ctx)
	toolbar.AddAction(btnLoop)
	btnLoop.SetIcon(icon.AvFiberSmartRecord)
	btnLoop.SetIconColor(material.White)
	btnLoop.OnPress = func() {
		if loop.IsOff() {
			btnLoop.SetIconColor(env.Palette().Accent)
			loop.Record()
		} else {
			btnLoop.SetIconColor(material.White)
			loop.Stop()
		}
	}

	btnMetronome := env.NewButton(ctx)
	toolbar.AddAction(btnMetronome)
	btnMetronome.SetIcon(icon.AvSlowMotionVideo)
	btnMetronome.SetIconColor(material.White)
	btnMetronome.OnPress = func() {
		if metronome.IsOff() {
			metronome.On()
		} else {
			metronome.Off()
		}
	}

	btnLowpass := env.NewButton(ctx)
	toolbar.AddAction(btnLowpass)
	btnLowpass.SetIcon(icon.AvSubtitles)
	btnLowpass.SetIconColor(material.White)
	btnLowpass.OnPress = func() {
		lowpass.SetPassthrough(!lowpass.Passthrough())
	}

	btnReverb := env.NewButton(ctx)
	toolbar.AddAction(btnReverb)
	btnReverb.SetIcon(icon.AvSurroundSound)
	btnReverb.SetIconColor(material.White)
	btnReverb.OnPress = func() {
		if reverb.IsOff() {
			reverb.On()
			keygain.SetAmp(snd.Decibel(-3).Amp())
		} else {
			reverb.Off()
			keygain.SetAmp(snd.Decibel(3).Amp())
		}
	}

	btnNext = env.NewFloatingActionButton(ctx)
	btnNext.Mini = true
	btnNext.SetColor(env.Palette().Accent)
	btnNext.SetIcon(icon.AvSkipNext)
	btnNext.OnPress = func() {
		sndbankpos = (sndbankpos + 1) % len(sndbank)
		go makekeys()
	}

	decor = env.NewMaterial(ctx)
	decor.SetColor(material.BlueGrey900)

	tseq := make(map[touch.Sequence]int)
	for i := range btnkeys {
		btnkeys[i] = env.NewButton(ctx)
		j := i
		btnkeys[i].OnTouch = func(ev touch.Event) {
			switch ev.Type {
			case touch.TypeBegin:
				keys[j].Press()
				tseq[ev.Sequence] = j
			case touch.TypeMove:
				// TODO drag finger off piano and it still plays, should stop
				if last, ok := tseq[ev.Sequence]; ok {
					if j != last {
						keys[last].Release()
						keys[j].Press()
						tseq[ev.Sequence] = j
					}
				}
			case touch.TypeEnd:
				keys[j].Release()
				delete(tseq, ev.Sequence)
			}
		}
	}

	var err error

	keymix = snd.NewMixer()
	go makekeys()
	lowpass = snd.NewLowPass(773, keymix)
	keygain = snd.NewGain(snd.Decibel(-3).Amp(), lowpass)

	dly := snd.NewDelay(29*time.Millisecond, keygain)
	tap0 := snd.NewTap(19*time.Millisecond, dly)
	tap1 := snd.NewTap(13*time.Millisecond, dly)
	tap2 := snd.NewTap(7*time.Millisecond, dly)
	cmb := snd.NewComb(snd.Decibel(-3).Amp(), 11*time.Millisecond, snd.NewMixer(dly, tap0, tap1, tap2))
	reverb = snd.NewLowPass(2000, cmb)
	dlymix := snd.NewMixer(reverb, keygain)

	loop = snd.NewLoopFrames(loopdur, dlymix)
	loop.SetBPM(bpm)
	loopmix := snd.NewMixer(dlymix, loop)

	master = snd.NewMixer(loopmix)
	mastergain = snd.NewGain(snd.Decibel(-6).Amp(), master)
	mixwf, err = NewWaveform(ctx, 2, mastergain)
	mixwf.SetColor(material.BlueGrey700)
	if err != nil {
		log.Fatal(err)
	}
	pan := snd.NewPan(0, mixwf)

	mtrosc := snd.NewOscil(sine, 440, nil)
	mtrdmp := snd.NewDamp(bpm.Dur(), mtrosc)
	metronome = snd.NewMixer(mtrdmp)
	metronome.Off()
	master.Append(metronome)

	al.AddSource(pan)
}

func onPaint(ctx gl.Context) {
	ctx.ClearColor(material.BlueGrey800.RGBA())
	ctx.Clear(gl.COLOR_BUFFER_BIT)
	env.Draw(ctx)

	now := time.Now()
	fps = int(time.Second / now.Sub(lastpaint))
	lastpaint = now
}

func onLayout(sz size.Event) {
	toolbar.Span(4, 4, 4)
	env.SetOrtho(sz)
	env.StartLayout()
	env.AddConstraints(btnNext.Constraints(env)...)
	env.AddConstraints(
		btnNext.EndIn(mixwf.Box, env.Grid.Gutter), btnNext.BottomIn(mixwf.Box, env.Grid.Gutter),
		mixwf.StartIn(env.Box, env.Grid.Gutter), mixwf.EndIn(decor.Box, 0),
		mixwf.Below(toolbar.Box, env.Grid.Gutter), mixwf.Above(decor.Box, env.Grid.Gutter), mixwf.Z(1),
	)

	wmjr := env.Grid.StepSize()*4/7 - 2
	wmnr := wmjr * 13.7 / 23.5
	hmjr := wmjr * 4
	hmnr := wmnr * 4

	prevmjr := func(n int) int {
		switch n {
		case 2:
			return 0
		case 4:
			return 2
		case 5:
			return 4
		case 7:
			return 5
		case 9:
			return 7
		case 11:
			return 9
		default:
			panic(fmt.Errorf("invalid prevmjr(%v)", n))
		}
	}
	for i, btn := range btnkeys {
		switch i {
		case 0:
			btn.SetColor(material.BlueGrey100)
			env.AddConstraints(
				btn.Width(wmjr), btn.Height(hmjr), btn.Z(4),
				btn.BottomIn(env.Box, env.Grid.Gutter), btn.StartIn(env.Box, env.Grid.Gutter))
		case 1, 3, 6, 8, 10:
			btn.SetColor(material.BlueGrey900)
			env.AddConstraints(
				btn.Width(wmnr), btn.Height(hmnr), btn.Z(6),
				btn.AlignTops(btnkeys[i-1].Box, 0), btn.StartIn(btnkeys[i-1].Box, wmjr-wmnr/2))
		default:
			btn.SetColor(material.BlueGrey100)
			env.AddConstraints(
				btn.Width(wmjr), btn.Height(hmjr), btn.Z(4),
				btn.BottomIn(env.Box, env.Grid.Gutter), btn.After(btnkeys[prevmjr(i)].Box, 2))
		}
	}
	env.AddConstraints(
		decor.Height(45), decor.Z(14), decor.Above(btnkeys[0].Box, 2),
		decor.StartIn(btnkeys[0].Box, 0), decor.EndIn(btnkeys[len(btnkeys)-1].Box, 0),
	)
	env.FinishLayout()
}

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		go func() {
			time.Sleep(10 * time.Second)
			pprof.StopCPUProfile()
		}()
	}

	app.Main(func(a app.App) {
		var logdbg *time.Ticker
		var glctx gl.Context
		for ev := range a.Events() {
			switch ev := a.Filter(ev).(type) {
			case lifecycle.Event:
				switch ev.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					logdbg = time.NewTicker(time.Second)
					go func() {
						for range logdbg.C {
							log.Printf("fps=%-4v underruns=%-4v buflen=%-4v tickavg=%-12s drift=%s\n",
								fps, al.Underruns(), al.BufLen(), al.TickAverge(), al.DriftApprox())
						}
					}()
					glctx = ev.DrawContext.(gl.Context)
					onStart(glctx)
					al.Start()
				case lifecycle.CrossOff:
					glctx = nil
					logdbg.Stop()
					al.Stop()
					al.CloseDevice()
				}
			case touch.Event:
				env.Touch(ev)
			case size.Event:
				if glctx == nil {
					a.Send(ev)
				} else {
					onLayout(ev)
				}
			case paint.Event:
				if glctx != nil {
					onPaint(glctx)
					a.Publish()
					a.Send(paint.Event{})
				}
			}
		}
	})
}
