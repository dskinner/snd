package main

import (
	"dasa.cc/piano/snd"
	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
)

var plt = newplttr()

type plttr struct {
	*plot.Plot
	names []string
	snds  []snd.Sound
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

func (plt *plttr) add(n string, s snd.Sound) {
	plt.snds = append(plt.snds, s)
	plt.names = append(plt.names, n)
}

func (plt *plttr) proc(iter int) {
	plt.outs = make([][]float64, len(plt.snds))
	for i := 1; i <= iter; i++ {
		for j, s := range plt.snds {
			s.Prepare(uint64(i))
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

func xyer(out []float64) plotter.XYs {
	n := len(out)
	xys := make(plotter.XYs, n)
	for i, v := range out {
		xys[i].X = float64(i) / float64(n)
		xys[i].Y = v
	}
	return xys
}

var (
	sine     = snd.Sine()
	sawtooth = snd.Sawtooth(4)
	square   = snd.Square(8)
)

func t1() {
	osc0 := snd.Osc(sine, 520, nil)
	osc1 := snd.Osc(sine, 440, nil)
	mix := snd.NewMixer(osc0, osc1)
	lp := snd.NewLowPass(500, mix)

	plt.add("Sine [520Hz, 440Hz]", mix)
	plt.add("Low Pass [500Hz]", lp)
	plt.proc(8)
}

func t2() {
	osc0 := snd.Osc(sawtooth, 440, nil)
	osc1 := snd.Osc(sawtooth, 440, nil)
	osc1.SetPhase(1, snd.Osc(square, 440*0.8, nil))

	plt.add("0", osc0)
	plt.add("1", osc1)
	plt.proc(4)
}

func main() {
	t2()
	plt.Save(1000, 500, "plot.png")
}
