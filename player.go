package snd

import (
	"fmt"
	"io"
	"math"

	"github.com/gen2brain/malgo"
)

type pbufc struct {
	xs []float64
	w  int
}

func newpbufc(n int) *pbufc {
	return &pbufc{xs: make([]float64, n)}
}

func (b *pbufc) readf32(bin []byte) {
	n := len(bin) / 4
	for i, x := range b.xs[:n] {
		n := math.Float32bits((float32)(x))
		bin[4*i] = byte(n)
		bin[4*i+1] = byte(n >> 8)
		bin[4*i+2] = byte(n >> 16)
		bin[4*i+3] = byte(n >> 24)
	}
	copy(b.xs, b.xs[n:])
	b.w -= n
}

func (b *pbufc) readi16(bin []byte) {
	n := len(bin) / 2
	for i, x := range b.xs[:n] {
		n := int16(math.MaxInt16 * x)
		bin[2*i] = byte(n)
		bin[2*i+1] = byte(n >> 8)
	}
	copy(b.xs, b.xs[n:])
	b.w -= n
}

func (b *pbufc) write(xs []float64) {
	copy(b.xs[b.w:], xs)
	b.w += len(xs)
}

func (b *pbufc) nwrites(atsize int) int {
	return (len(b.xs) - b.w) / atsize
}

type Player struct {
	dp *Dispatcher
	in Sound
	tc uint64

	inputs []*Input

	line *pbufc

	uninit func()
}

func NewPlayer(in Sound) *Player {
	return &Player{
		dp:     new(Dispatcher),
		in:     in,
		inputs: GetInputs(in),
		line:   newpbufc(4096),
	}
}

func (p *Player) Notify() {
	if p.in != nil {
		p.inputs = GetInputs(p.in)
	}
}

func (p *Player) Channels() uint32   { return uint32(p.in.Channels()) }
func (p *Player) SampleRate() uint32 { return uint32(p.in.SampleRate()) }

// func binf32(xs []float32) []byte {
// 	return unsafe.Slice((*byte)(unsafe.Pointer(&xs[0])), 4*len(xs))
// }

func (p *Player) Read(bin []byte) (int, error) {
	nwrites := p.line.nwrites(DefaultBufferLen)
	// fmt.Printf("performing nwrites %v\n", nwrites)
	for i := 0; i < nwrites; i++ {
		p.tc++
		p.dp.Dispatch(p.tc, p.inputs...)
		p.line.write(p.in.Samples())
	}

	p.line.readf32(bin)

	return len(bin), nil
}

func (pl *Player) Start() error {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("LOG %v", message)
	})
	if err != nil {
		return err
	}

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)
	deviceConfig.Playback.Format = malgo.FormatF32
	deviceConfig.Playback.Channels = pl.Channels()
	deviceConfig.SampleRate = pl.SampleRate()
	deviceConfig.Alsa.NoMMap = 1

	// This is the function that's used for sending more data to the device for playback.
	onSamples := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		io.ReadFull(pl, pOutputSample)
	}

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: onSamples,
	}
	device, err := malgo.InitDevice(ctx.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		return err
	}

	err = device.Start()
	if err != nil {
		return err
	}

	pl.uninit = func() {
		device.Uninit()
		_ = ctx.Uninit()
		ctx.Free()
	}
	return nil
}

func (pl *Player) Stop() {
	var uninit func()
	uninit, pl.uninit = pl.uninit, uninit
	if uninit != nil {
		uninit()
	}
}
