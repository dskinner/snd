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

	source  al.Source
	buffers []al.Buffer
	bufidx  int

	format uint32
	in     snd.Sound
	out    []byte

	underruns uint64
}

func OpenDevice(buflen int) error {
	if err := al.OpenDevice(); err != nil {
		return fmt.Errorf("snd/al: open device failed: %s", err)
	}
	if buflen == 0 || buflen&(buflen-1) != 0 {
		return fmt.Errorf("snd/al: buflen(%v) not a power of 2", buflen)
	}
	hwa = &openal{buflen: buflen}
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

	b := al.GenBuffers(hwa.buflen)
	if code := al.Error(); code != 0 {
		return fmt.Errorf("snd/al: generate buffers failed [err=%v]", code)
	}
	hwa.buffers = b

	for i := 0; i < len(hwa.buffers); i++ {
		prepare()
	}

	d := time.Duration(float64(in.FrameLen()) / in.SampleRate() * float64(time.Second) * float64(len(hwa.buffers)))
	log.Println("snd: latency:", d)
	// tick := time.Tick(d / 2)

	return nil
}

func Tick() {
	if code := al.DeviceError(); code != 0 {
		log.Printf("snd/al: unknown device error [err=%v]\n", code)
	}
	if code := al.Error(); code != 0 {
		log.Printf("snd/al: unknown error [err=%v]\n", code)
	}

	for i := hwa.source.BuffersProcessed(); i > 0; i-- {
		prepare()
	}

	switch hwa.source.State() {
	case al.Initial:
		al.PlaySources(hwa.source)
	case al.Playing:
	case al.Paused:
	case al.Stopped:
		hwa.underruns++
		log.Println("buffer underrun")
		al.PlaySources(hwa.source)
	}
}

func prepare() {
	//
	// time.Sleep(500 * time.Millisecond)

	// t := time.Now()
	hwa.in.Prepare()
	// log.Println(time.Now().Sub(t))

	// TODO don't just assume stereo sound
	// encode final signal
	for i, x := range hwa.in.Output() {
		n := int16(math.MaxInt16*x) / reduce
		hwa.out[2*i] = byte(n)
		hwa.out[2*i+1] = byte(n >> 8)
	}

	// queue
	hwa.bufidx = (hwa.bufidx + 1) % len(hwa.buffers)

	// TODO on first proc, buffers aren't queued so this causes error.
	hwa.source.UnqueueBuffers([]al.Buffer{hwa.buffers[hwa.bufidx]})
	if code := al.Error(); code != 0 {
		log.Printf("snd: unqueue buffer failed [id=%v] [err=%v]\n", hwa.bufidx, code)
	}

	hwa.buffers[hwa.bufidx].BufferData(hwa.format, hwa.out, int32(hwa.in.SampleRate()))
	if code := al.Error(); code != 0 {
		log.Printf("snd: buffer data failed [id=%v] [err=%v]\n", hwa.bufidx, code)
	}

	hwa.source.QueueBuffers([]al.Buffer{hwa.buffers[hwa.bufidx]})
	if code := al.Error(); code != 0 {
		log.Printf("snd: queue buffer failed [id=%v] [err=%v]\n", hwa.bufidx, code)
	}
}
