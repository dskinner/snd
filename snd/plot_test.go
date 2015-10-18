// +build plot

package snd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
)

type plttr struct {
	*plot.Plot
	nproc  int
	nlines int
}

func newplttr(nproc int) *plttr {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	plt := &plttr{Plot: p, nproc: nproc}
	plt.X.Min, plt.X.Max = 0, 1
	plt.Y.Min, plt.Y.Max = -1.5, 1.5
	return plt
}

func (plt *plttr) add(name string, sd Sound) {
	inps := GetInputs(sd)
	dp := new(Dispatcher)
	var out []float64
	for i := 1; i <= plt.nproc; i++ {
		dp.Dispatch(uint64(i), inps...)
		out = append(out, sd.Samples()...)
	}

	plt.addDiscrete(name, Discrete(out))
}

func (plt *plttr) addDiscrete(name string, sig Discrete) {
	// TODO there appears to be a bug in gonum plot where certain
	// dashed lines for a particular result will not render correctly.
	// Raw calls of plotutil are just tossed in here for now and avoids
	// dashed lines.
	l, err := plotter.NewLine(xyer([]float64(sig)))
	if err != nil {
		panic(err)
	}
	l.Color = plotutil.Color(plt.nlines)
	// l.Dashes = plotutil.Dashes(plt.nlines)
	// l.LineStyle.Width = 2

	plt.Add(l)
	plt.Legend.Add(name, l)

	plt.nlines++
}

func (plt *plttr) save(name string) error {
	plt.Add(plotter.NewGrid())
	if err := os.MkdirAll("plots", 0755); err != nil {
		return err
	}
	return plt.Save(1000, 500, filepath.Join("plots", name))
}

func xyer(out []float64) plotter.XYs {
	n := len(out)
	xys := make(plotter.XYs, n)
	for i, v := range out {
		xys[i].X = float64(i) / float64(n)
		xys[i].Y = v
	}
	return xys
}

func TestPlotOscil(t *testing.T) {
	plt := newplttr(4)
	// plt.add("Sine", NewOscil(Sine(), 440, nil))
	// plt.add("Triangle", NewOscil(Triangle(), 440, nil))
	// plt.add("Square", NewOscil(Square(), 440, nil))
	// plt.add("Sawtooth", NewOscil(SawtoothSynthesis(2, 0), 440, nil))
	// plt.add("Pulse", NewOscil(PulseSynthesis(64, 0), 440, nil))

	// sig := Sine()
	// for i := 2; i <= 2; i++ {
	// sig.Add(Sawtooth(), i, 0)
	// }
	// plt.add("Add()", NewOscil(sig, 440, nil))

	plt.addDiscrete("Sine", Sine())

	// plt.addDiscrete("Square Synth", SquareSynthesis(99, 0))
	// sig := Sine()
	// add odd partials to produce square via additive synthesis
	// for i := 3; i <= 99; i += 2 {
	// sig.Add(Sine(), i)
	// }
	// plt.addDiscrete("Square Add", sig)

	plt.addDiscrete("Sawtooth Synth", SawtoothSynthesis(64, 0))
	sig := Sine()
	for i := 2; i <= 64; i++ {
		sig.Add(Sine(), i)
	}
	plt.addDiscrete("Sawtooth Add", sig)

	// osc0 := NewOscil(Sine(), 220, nil)
	// osc0.Prepare(1)
	// sig0 := Discrete(osc0.Samples())
	// osc1 := NewOscil(Sine(), 440, nil)
	// osc1.Prepare(1)
	// sig1 := Discrete(osc1.Samples())
	// sig0.Add(sig1, 0)
	// plt.add("xtra", NewOscil(sig0, 660, nil))

	if err := plt.save("oscil.png"); err != nil {
		t.Fatal(err)
	}
}

func TestPlotPhaser(t *testing.T) {
	plt := newplttr(8)
	sawtooth := SawtoothSynthesis(4, 0)
	square := SquareSynthesis(4, 0)

	osc0 := NewOscil(sawtooth, 440, nil)
	plt.add("oscil", osc0)

	osc1 := NewOscil(sawtooth, 440, nil)
	osc1.SetPhase(1, NewOscil(square, 440*0.8, nil))
	plt.add("phased", osc1)

	if err := plt.save("phaser.png"); err != nil {
		t.Fatal(err)
	}
}

func TestPlotADSR(t *testing.T) {
	plt := newplttr(256)
	ms := time.Millisecond
	// adsr := NewADSR(50*ms, 100*ms, 150*ms, 200*ms, 0.3, 0.6, nil)
	adsr := NewADSR(10*ms, 5*ms, 400*ms, 350*ms, 0.4, 1, nil)
	plt.add("adsr", adsr)
	if err := plt.save("adsr.png"); err != nil {
		t.Fatal(err)
	}
}

func TestPlotLowPass(t *testing.T) {
	plt := newplttr(4)

	mix0 := NewMixer(NewOscil(Sine(), 520, nil), NewOscil(Sine(), 440, nil))
	plt.add("Mix Sine [520Hz, 440Hz]", mix0)

	mix1 := NewMixer(NewOscil(Sine(), 520, nil), NewOscil(Sine(), 440, nil))
	lp := NewLowPass(500, mix1)
	plt.add("Low Pass [500Hz]", lp)

	if err := plt.save("lowpass.png"); err != nil {
		t.Fatal(err)
	}
}

func TestPlotDelay(t *testing.T) {
	plt := newplttr(4)

	osc0 := NewOscil(Sine(), 440, nil)
	plt.add("Sine 440Hz", osc0)

	dly := NewDelay(ftod(DefaultBufferLen, DefaultSampleRate), NewOscil(Sine(), 440, nil))
	plt.add("Delay", dly)

	if err := plt.save("delay.png"); err != nil {
		t.Fatal(err)
	}
}

func TestPlotDamp(t *testing.T) {
	plt := newplttr(4)

	osc0 := NewOscil(Sine(), 440, nil)
	plt.add("440Hz", osc0)

	d := ftod(DefaultBufferLen*4, DefaultSampleRate)

	osc1 := NewOscil(Sine(), 440, nil)
	dmp := NewDamp(d, osc1)
	plt.add("Damped", dmp)

	plt.add("Force", NewDamp(d, newunit()))

	if err := plt.save("damp.png"); err != nil {
		t.Fatal(err)
	}
}

func TestPlotDrive(t *testing.T) {
	plt := newplttr(8)

	osc0 := NewOscil(Sine(), 440, nil)
	plt.add("440Hz", osc0)

	d := ftod(DefaultBufferLen*4, DefaultSampleRate)

	osc1 := NewOscil(Sine(), 440, nil)
	drv := NewDrive(d, osc1)
	plt.add("Driven", drv)

	plt.add("Force", NewDrive(d, newunit()))

	if err := plt.save("drive.png"); err != nil {
		t.Fatal(err)
	}
}
