package snd

import (
	"sort"
	"sync"
)

// TODO most of this probably doesn't need to be exposed
// but those details can be worked out once additional audio
// drivers are supported.

type Dispatcher struct{ sync.WaitGroup }

// Dispatch blocks until all inputs are prepared.
func (dp *Dispatcher) Dispatch(tc uint64, inps ...*Input) {
	wt := inps[0].wt
	for _, inp := range inps {
		if inp.wt != wt {
			dp.Wait()
			wt = inp.wt
		}
		dp.Add(1)
		go func(sd Sound, tc uint64) {
			sd.Prepare(tc)
			dp.Done()
		}(inp.sd, tc)
	}
	dp.Wait()
}

// This is a game I do not want to play. No guarantees can really be made about a slice
// of inputs in a goroutine b/c if a single input takes unusually long, it's still in the way
// of others in that routine.
func (dp *Dispatcher) DispatchSlice(tc uint64, sl [][]*Input) {
	const ln = 4
	for _, p := range sl {
		n := len(p) / ln
		i := 0
		for ; i < n; i++ {
			inps := p[i*ln : i*ln+ln]
			dp.Add(1)
			go func(tc uint64, inps ...*Input) {
				for _, inp := range inps {
					inp.sd.Prepare(tc)
				}
				dp.Done()
			}(tc, inps...)
		}
		// end of slice
		inps := p[i*ln:]
		dp.Add(1)
		go func(tc uint64, inps ...*Input) {
			for _, inp := range inps {
				inp.sd.Prepare(tc)
			}
			dp.Done()
		}(tc, inps...)

		//
		dp.Wait()
	}
}

type Input struct {
	sd Sound
	wt int
}

type ByWT []*Input

func (a ByWT) Len() int           { return len(a) }
func (a ByWT) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByWT) Less(i, j int) bool { return a[i].wt > a[j].wt }

func (a ByWT) Slice() (sl [][]*Input) {
	if len(a) == 0 {
		return nil
	}
	wt := a[0].wt
	i := 0
	for j, p := range a {
		if p.wt != wt {
			sl = append(sl, a[i:j])
			i = j
			wt = p.wt
		}
	}
	return append(sl, a[i:])
}

func GetInputs(sd Sound) []*Input {
	inps := []*Input{{sd, 0}}
	getinputs(sd, 1, &inps)
	sort.Sort(ByWT(inps))
	return inps
}

// TODO janky func
func getinputs(sd Sound, wt int, out *[]*Input) {
	for _, in := range sd.Inputs() {
		if in == nil {
			continue
		}
		at := -1
		for i, p := range *out {
			if p.sd == in {
				if p.wt >= wt {
					return // object has or will be traversed on different path
				}
				at = i
				break
			}
		}
		if at != -1 {
			(*out)[at].sd = in
			(*out)[at].wt = wt
		} else {
			*out = append(*out, &Input{in, wt})
		}
		getinputs(in, wt+1, out)
	}
}

type Job struct {
	sd Sound
	tc uint64
}

type Worker1 struct {
	pool chan chan Job
	jobs chan Job
	quit chan struct{}
}

func NewWorker1(pool chan chan Job) Worker1 {
	return Worker1{pool, make(chan Job), make(chan struct{})}
}

func (w Worker1) Start(wg *sync.WaitGroup) {
	go func() {
		for {
			w.pool <- w.jobs
			select {
			case job := <-w.jobs:
				job.sd.Prepare(job.tc)
				wg.Done()
			case <-w.quit:
				return
			}
		}
	}()
}

func (w Worker1) Stop() {
	go func() { w.quit <- struct{}{} }()
}

type Dispatcher1 struct {
	sync.WaitGroup
	workers chan chan Job
}

func NewDispatcher1(n int) *Dispatcher1 {
	return &Dispatcher1{workers: make(chan chan Job, n)}
}

func (dp *Dispatcher1) Run() {
	for i := 0; i < cap(dp.workers); i++ {
		NewWorker1(dp.workers).Start(&dp.WaitGroup)
	}
}

func (dp *Dispatcher1) Dispatch(sd Sound, tc uint64) {
	dp.Add(1)
	(<-dp.workers) <- Job{sd, tc}
}

type Dispatcher2 struct {
	sync.WaitGroup
	jobs chan Job
}

func NewDispatcher2() *Dispatcher2 {
	return &Dispatcher2{jobs: make(chan Job)}
}

func (dp *Dispatcher2) Run(n int) {
	for i := 0; i < n; i++ {
		go func() {
			for job := range dp.jobs {
				job.sd.Prepare(job.tc)
				dp.Done()
			}
		}()
	}
}

func (dp *Dispatcher2) Dispatch(tc uint64, inps ...*Input) {
	wt := inps[0].wt
	for _, inp := range inps {
		if inp.wt != wt {
			dp.Wait()
			wt = inp.wt
		}
		dp.Add(1)
		dp.jobs <- Job{inp.sd, tc}
	}
	dp.Wait()
}
