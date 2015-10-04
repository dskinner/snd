package snd

import "math"

var (
	onesqrt2 = 1 / math.Sqrt(2)

	// panpos [panres*2]float64
	panres float64 = 512
	panpos [1024]float64
)

func init() {
	for i := range panpos {
		n := float64(i)/panres - 1
		panpos[i] = onesqrt2 * (1 - n) / math.Sqrt(1+(n*n))
	}
}

// TODO refactor this mess
func getpanpos(amt float64, ch int) float64 {
	if -1 <= amt && amt <= 1 {
		if ch == 0 {
			return panpos[int(panres*(1+amt))]
		}
		return panpos[int(panres*(1-amt))]
	} else if amt < -1 { // consider making this == or removing entirely bc; amt ∈ [-1..1]
		if ch == 0 {
			return panpos[0]
		}
		return 0
	} else {
		if ch == 0 {
			return 0
		}
		return panpos[0]
	}
}

type Pan struct {
	*stereosnd

	amt float64
}

func NewPan(amt float64, in Sound) *Pan {
	return &Pan{newStereosnd(in), amt}
}

// SetAmount takes an input pans it across two outputs by
// an amount given as -1 to 1.
func (p *Pan) SetAmount(amt float64) { p.amt = amt }

// Prepare interleaves the left and right channels.
func (p *Pan) Prepare() {
	p.stereosnd.Prepare()
	for i, x := range p.in.Output() {
		p.l.out[i] = x * getpanpos(p.amt, 0)
		p.r.out[i] = x * getpanpos(p.amt, 1)
		p.out[i*2] = p.l.out[i]
		p.out[i*2+1] = p.r.out[i]
	}
}
