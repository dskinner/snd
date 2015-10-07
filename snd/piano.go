package snd

type Piano struct {
	*snd

	keys []float64
	idx  int
}

func NewPiano() *Piano {
	wf := &Piano{}
	wf.snd = newSnd(nil)

	wf.keys = make([]float64, DefaultSampleSize*4)

	space := 16

	hasmn := []struct {
		left, right bool
	}{
		{false, true}, // C
		{true, true},  // D
		{true, false}, // E
		{false, true}, // F
		{true, true},  // G
		{true, true},  // A
		{true, false}, // B
	}

	nkeys := len(hasmn)
	mj := len(wf.keys) / nkeys

	// dinky piano signal
	for i := 0; i < len(wf.keys); i += 2 {
		key := i / mj
		// if key >= nkeys {
		// wf.keys[i] = -1
		// wf.keys[i+1] = -1
		// continue
		// }

		// marker for signal alignment
		if i <= space {
			wf.keys[i] = -0.999
			wf.keys[i+1] = -0.999
			continue
		} else if i >= (len(wf.keys) - space) {
			wf.keys[i] = -0.98
			wf.keys[i+1] = -0.98
			continue
		}

		j := i % mj
		if j <= space || (mj-j) >= (mj-space) {
			// spacing for major keys
			wf.keys[i] = -1
		} else if (j <= (mj/4) && hasmn[key].left) || (j >= mj-(mj/4) && hasmn[key].right) {
			// minor key
			wf.keys[i] = -0.3
		} else {
			// major key
			wf.keys[i] = 1
		}

		wf.keys[i+1] = -1
	}

	return wf
}

func (wf *Piano) Prepare() {
	for i := range wf.snd.out {
		wf.snd.out[i] = wf.keys[wf.idx]
		wf.idx = (wf.idx + 1) % len(wf.keys)
	}
}
