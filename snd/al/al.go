package al

import (
	"fmt"
	"log"
	"math"
	"time"

	"dasa.cc/piano/snd"

	"golang.org/x/mobile/exp/audio/al"
)

var hwa *openal

type openal struct {
	buflen int

	source al.Source

	buffers []al.Buffer
	bufproc int // number of buffers processed
	bufidx  int // index into buffers

	format uint32
	in     snd.Sound
	out    []byte

	quit chan struct{}

	underruns uint64
	preptime  time.Duration
	prepcalls uint64
	tickcount uint64
}

func OpenDevice(buflen int) error {
	if err := al.OpenDevice(); err != nil {
		return fmt.Errorf("snd/al: open device failed: %s", err)
	}
	if buflen == 0 || buflen&(buflen-1) != 0 {
		return fmt.Errorf("snd/al: buflen(%v) not a power of 2", buflen)
	}
	hwa = &openal{buflen: buflen, prepcalls: 1}
	return nil
}

func CloseDevice() error {
	al.DeleteBuffers(hwa.buffers)
	al.DeleteSources(hwa.source)
	al.CloseDevice()
	hwa = nil
	return nil
}

func AddSource(in snd.Sound) error {
	switch in.Channels() {
	case 1:
		hwa.format = al.FormatMono16
	case 2:
		hwa.format = al.FormatStereo16
	default:
		return fmt.Errorf("snd/al: can't handle input with channels(%v)", in.Channels())
	}
	hwa.in = in
	hwa.out = make([]byte, in.FrameLen()*in.Channels()*2)

	s := al.GenSources(1)
	if code := al.Error(); code != 0 {
		return fmt.Errorf("snd/al: generate source failed [err=%v]", code)
	}
	hwa.source = s[0]

	// b := al.GenBuffers(hwa.buflen)
	// if code := al.Error(); code != 0 {
	// 	return fmt.Errorf("snd/al: generate buffers failed [err=%v]", code)
	// }
	// hwa.buffers = b

	// for i := 0; i < len(hwa.buffers); i++ {
	// 	incbufferidx()
	// 	queue()
	// }

	d := time.Duration(float64(in.FrameLen()) / in.SampleRate() * float64(time.Second) * float64(hwa.buflen))
	log.Println("snd/al: latency", d)

	return nil
}

func Start() {
	if hwa.quit != nil {
		panic("snd/al: hwa.quit not nil")
	}
	hwa.quit = make(chan struct{})
	go func() {
		d := time.Duration(float64(hwa.in.FrameLen()) / hwa.in.SampleRate() * float64(time.Second) * float64(hwa.buflen))
		tick := time.Tick(d) // / 2) // TODO i want to tick at half the rate!!!! or ill risk overruns all the time. this is for debugging only
		for {
			select {
			case <-hwa.quit:
				return
			case <-tick:
				Tick()
			}
		}
	}()
}

func Stop() {
	close(hwa.quit)
}

func Tick() {
	hwa.tickcount++

	if code := al.DeviceError(); code != 0 {
		log.Printf("snd/al: unknown device error [err=%v] [tick=%v]\n", code, hwa.tickcount)
	}
	if code := al.Error(); code != 0 {
		log.Printf("snd/al: unknown error [err=%v] [tick=%v]\n", code, hwa.tickcount)
	}

	// for i := hwa.source.BuffersProcessed(); i > 0; i-- {
	// start := time.Now()
	// incbufferidx()
	// if err := unqueue(); err != nil {
	// log.Println(err)
	// } else if err := queue(); err != nil {
	// log.Println(err)
	// }
	// hwa.preptime += time.Now().Sub(start)
	// hwa.prepcalls++
	// }

	// @@@ make new buffers everytime

	// TODO I'm ticking at half the rate of time to fill buffers!!! so i have use buflen/2 below
	start := time.Now()
	var bufs []al.Buffer
	if hwa.bufproc = int(hwa.source.BuffersProcessed()); hwa.bufproc >= hwa.buflen {
		// TODO advance hwa.bufidx somewhere
		// bufidx is last index we're aware of that started processing
		//
		// if we only advance by buflen, then hopefully buffersprocessed will report
		// half as much next time around allowing for buffers processed to be greater than
		// buflen.
		bufs = hwa.buffers[hwa.bufidx : hwa.bufidx+hwa.buflen]
		hwa.source.UnqueueBuffers(bufs)
		if code := al.Error(); code != 0 {
			log.Printf("snd/al: unqueue buffers failed [err=%v]\n", code)
		}
		hwa.bufidx = (hwa.bufidx + hwa.buflen) % len(hwa.buffers)
	} else {
		// let's make some more buffers since we're here to fill data
		// regardless of what openal says.
		bufs = al.GenBuffers(hwa.buflen)
		if code := al.Error(); code != 0 {
			log.Printf("snd/al: generate buffers failed [err=%v]", code)
		}
		hwa.buffers = append(hwa.buffers, bufs...)
	}

	for _, buf := range bufs {
		hwa.in.Prepare()
		for i, x := range hwa.in.Output() {
			n := int16(math.MaxInt16 * x)
			hwa.out[2*i] = byte(n)
			hwa.out[2*i+1] = byte(n >> 8)
		}

		buf.BufferData(hwa.format, hwa.out, int32(hwa.in.SampleRate()))
		if code := al.Error(); code != 0 {
			log.Printf("snd/al: buffer data failed [id=%v] [err=%v] [tick=%v]\n", hwa.bufidx, code, hwa.tickcount)
		}
	}

	hwa.source.QueueBuffers(bufs)
	if code := al.Error(); code != 0 {
		log.Printf("snd/al: queue buffer failed [id=%v] [err=%v] [tick=%v]\n", hwa.bufidx, code, hwa.tickcount)
	}
	hwa.preptime += time.Now().Sub(start)
	hwa.prepcalls++

	// log.Printf("buflen[%v] bufproc[%v] bufidx[%v]\n", len(hwa.buffers), hwa.bufproc, hwa.bufidx)
	// @@@

	switch hwa.source.State() {
	case al.Initial:
		al.PlaySources(hwa.source)
	case al.Playing:
	case al.Paused:
	case al.Stopped:
		hwa.underruns++
		al.PlaySources(hwa.source)
	}
}

func BufStats() (buflen int, underruns uint64) {
	return len(hwa.buffers), hwa.underruns
}

func PrepStats() (total time.Duration, calls uint64) {
	return hwa.preptime, hwa.prepcalls
}

func incbufferidx() {
	hwa.bufidx = (hwa.bufidx + 1) % len(hwa.buffers)
}

func unqueue() error {
	hwa.source.UnqueueBuffers([]al.Buffer{hwa.buffers[hwa.bufidx]})
	if code := al.Error(); code != 0 {
		return fmt.Errorf("snd/al: unqueue buffer failed [id=%v] [err=%v] [tick=%v]\n", hwa.bufidx, code, hwa.tickcount)
	}
	return nil
}

func queue() error {
	hwa.in.Prepare()
	for i, x := range hwa.in.Output() {
		n := int16(math.MaxInt16 * x)
		hwa.out[2*i] = byte(n)
		hwa.out[2*i+1] = byte(n >> 8)
	}

	hwa.buffers[hwa.bufidx].BufferData(hwa.format, hwa.out, int32(hwa.in.SampleRate()))
	if code := al.Error(); code != 0 {
		return fmt.Errorf("snd/al: buffer data failed [id=%v] [err=%v] [tick=%v]\n", hwa.bufidx, code, hwa.tickcount)
	}

	hwa.source.QueueBuffers([]al.Buffer{hwa.buffers[hwa.bufidx]})
	if code := al.Error(); code != 0 {
		return fmt.Errorf("snd/al: queue buffer failed [id=%v] [err=%v] [tick=%v]\n", hwa.bufidx, code, hwa.tickcount)
	}

	return nil
}
