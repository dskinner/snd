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
	names []string
	snds  []Sound
	outs  [][]float64
}

func newplttr() *plttr {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	plt := &plttr{Plot: p}
	plt.X.Min, plt.X.Max = 0, 1
	plt.Y.Min, plt.Y.Max = -1, 1
	return plt
}

func (plt *plttr) add(n string, s Sound) {
	plt.snds = append(plt.snds, s)
	plt.names = append(plt.names, n)
}

func (plt *plttr) proc(iter int) {
	plt.outs = make([][]float64, len(plt.snds))
	for i := 1; i <= iter; i++ {
		for j, s := range plt.snds {
			for _, inp := range GetInputs(s) {
				inp.sd.Prepare(uint64(i))
			}
			plt.outs[j] = append(plt.outs[j], s.Samples()...)
		}
	}

	vs := make([]interface{}, len(plt.outs)*2)
	for i, out := range plt.outs {
		vs[i*2] = plt.names[i]
		vs[i*2+1] = xyer(out)
	}

	if err := plotutil.AddLines(plt.Plot, vs...); err != nil {
		panic(err)
	}
}

func (plt *plttr) save(name string) error {
	if err := os.MkdirAll("plots", 0755); err != nil {
		panic(err)
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
	plt := newplttr()
	osc := NewOscil(Sine(), 440, nil)
	plt.add("Sine 440Hz", osc)
	plt.proc(4)
	if err := plt.save("oscil.png"); err != nil {
		t.Fatal(err)
	}
}

func TestPlotPhaser(t *testing.T) {
	plt := newplttr()
	sawtooth := Sawtooth(4)
	square := Square(4)

	osc0 := NewOscil(sawtooth, 440, nil)
	plt.add("oscil", osc0)

	osc1 := NewOscil(sawtooth, 440, nil)
	osc1.SetPhase(1, NewOscil(square, 440*0.8, nil))
	plt.add("phased", osc1)

	plt.proc(8)
	if err := plt.save("phaser.png"); err != nil {
		t.Fatal(err)
	}
}

func TestPlotADSR(t *testing.T) {
	plt := newplttr()
	ms := time.Millisecond
	env := NewADSR(50*ms, 100*ms, 150*ms, 200*ms, 0.3, 0.6, nil)
	plt.add("adsr", env)

	plt.proc(256)
	if err := plt.save("adsr.png"); err != nil {
		t.Fatal(err)
	}
}

func TestLowPass(t *testing.T) {
	plt := newplttr()

	mix0 := NewMixer(NewOscil(Sine(), 520, nil), NewOscil(Sine(), 440, nil))
	plt.add("Mix Sine [520Hz, 440Hz]", mix0)

	mix1 := NewMixer(NewOscil(Sine(), 520, nil), NewOscil(Sine(), 440, nil))
	lp := NewLowPass(500, mix1)
	plt.add("Low Pass [500Hz]", lp)

	plt.proc(4)
	if err := plt.save("lowpass.png"); err != nil {
		t.Fatal(err)
	}
}
