package snd

import (
	"fmt"
	"log"
	"math"
	"time"

	"golang.org/x/mobile/exp/audio/al"
)

// TODO move this to a snd/openal package and then an application can just
// import (
//   "dasa.cc/snd"
//   _ "dasa.cc/snd/openal"
// )
// ...
//   dev, err := snd.OpenDevice("openal")
//
type OpenAL struct {
	source  al.Source
	buffers []al.Buffer
	bufidx  int

	input  Sound
	output []byte

	quit chan struct{}
}

func NewOpenAL() *OpenAL {
	return &OpenAL{
		input:  NewMixer(),
		output: make([]byte, DefaultSampleSize*2*2), // fake stereo!
		quit:   make(chan struct{}),
	}
}

// TODO really kind of a Output given this is the "device"
func (oal *OpenAL) Input() Sound      { return oal.input }
func (oal *OpenAL) SetInput(in Sound) { oal.input = in }

func (oal *OpenAL) OpenDevice(buflen int) error {
	if buflen == 0 || buflen&(buflen-1) != 0 {
		return fmt.Errorf("snd: buflen(%v) not a power of 2", buflen)
	}

	if err := al.OpenDevice(); err != nil {
		return fmt.Errorf("snd: open device failed: %s", err)
	}

	s := al.GenSources(1)
	if code := al.Error(); code != 0 {
		return fmt.Errorf("snd: generate source failed [err=%v]", code)
	}
	oal.source = s[0]

	// TODO force buffers to be power of 2 once allowed to be configured
	b := al.GenBuffers(buflen)
	if code := al.Error(); code != 0 {
		return fmt.Errorf("snd: generate buffers failed [err=%v]", code)
	}
	oal.buffers = b

	return nil
}

func (oal *OpenAL) CloseDevice() {
	al.DeleteBuffers(oal.buffers)
	al.DeleteSources(oal.source)
	al.CloseDevice()
}

func (oal *OpenAL) prepare() {
	//
	oal.input.Prepare()

	// TODO don't just assume stereo sound
	// encode final signal
	for i, x := range oal.input.Output() {
		n := int16(math.MaxInt16*x) / reduce
		oal.output[2*i] = byte(n)
		oal.output[2*i+1] = byte(n >> 8)
	}

	// queue
	oal.bufidx = (oal.bufidx + 1) % len(oal.buffers)

	// TODO on first proc, buffers aren't queued so this causes error.
	oal.source.UnqueueBuffers([]al.Buffer{oal.buffers[oal.bufidx]})
	if code := al.Error(); code != 0 {
		log.Printf("snd: unqueue buffer failed [id=%v] [err=%v]\n", oal.bufidx, code)
	}

	oal.buffers[oal.bufidx].BufferData(al.FormatStereo16, oal.output, int32(DefaultSampleRate))
	if code := al.Error(); code != 0 {
		log.Printf("snd: buffer data failed [id=%v] [err=%v]\n", oal.bufidx, code)
	}

	oal.source.QueueBuffers([]al.Buffer{oal.buffers[oal.bufidx]})
	if code := al.Error(); code != 0 {
		log.Printf("snd: queue buffer failed [id=%v] [err=%v]\n", oal.bufidx, code)
	}
}

func (oal *OpenAL) Play() {
	go func() {
		d := time.Duration(float64(DefaultSampleSize) / DefaultSampleRate * float64(time.Second) * float64(len(oal.buffers)))
		log.Println("snd: latency:", d)
		tick := time.Tick(d / 2)

		for i := 0; i < len(oal.buffers); i++ {
			oal.prepare()
		}

		// forward := true

		for {
			select {
			case <-oal.quit:
				return
			case <-tick:
				if code := al.DeviceError(); code != 0 {
					log.Printf("snd: unknown device error [err=%v]\n", code)
				}
				if code := al.Error(); code != 0 {
					log.Printf("snd: unknown error [err=%v]\n", code)
				}

				// sl := oal.input.(Slice)
				// pan := sl[len(sl)-1].(*Pan)
				// if forward {
				// pan.amt += 0.01
				// forward = pan.amt < 1
				// } else {
				// pan.amt -= 0.01
				// forward = pan.amt < -1
				// }

				// log.Println("proc:", oal.source.BuffersProcessed())
				// log.Println("queu:", oal.source.BuffersQueued())
				// log.Println("offs:", oal.source.OffsetByte())
				// log.Println("offs:", oal.source.OffsetSample())
				// log.Println("----")

				// TODO could maybe detect buffer underruns and automatically increase
				// num buffers on the fly

				for i := oal.source.BuffersProcessed(); i > 0; i-- {
					oal.prepare()
				}

				switch oal.source.State() {
				case al.Initial:
					al.PlaySources(oal.source)
				case al.Playing:
				case al.Paused:
				case al.Stopped:
					al.PlaySources(oal.source)
				}
			}
		}
	}()
}

func (oal *OpenAL) Destroy() {
	if oal == nil {
		return
	}
	close(oal.quit)
}
