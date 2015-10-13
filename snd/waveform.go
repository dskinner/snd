package snd

import (
	"fmt"
	"log"
	"math"

	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/gl"
)

// TODO this is intended to graphically represent sound using opengl
// but the package is "snd". It doesn't make much sense to require
// go mobile gl to build snd (increasing complexity of portability)
// so move this to a subpkg requiring explicit importing.

// TODO should remove *Buffer and move that functionality directly into here.
type Waveform struct {
	Sound

	program  gl.Program
	position gl.Attrib
	color    gl.Uniform
	buf      gl.Buffer

	outs    [][]float64
	samples []float64

	align    bool
	alignamp float64
	aligned  []float64

	verts []float32

	data []byte
}

// TODO just how many samples do we want/need to display something useful?
func NewWaveform(ctx gl.Context, n int, in Sound) (*Waveform, error) {
	wf := &Waveform{Sound: in}

	wf.outs = make([][]float64, n)
	for i := range wf.outs {
		wf.outs[i] = make([]float64, in.BufferLen()*in.Channels())
	}
	wf.samples = make([]float64, in.BufferLen()*in.Channels()*n)
	wf.aligned = make([]float64, in.BufferLen()*in.Channels()*n)

	wf.verts = make([]float32, len(wf.aligned)*3)
	wf.data = make([]byte, len(wf.verts)*4)

	var err error
	wf.program, err = glutil.CreateProgram(ctx, vertexShader, fragmentShader)
	if err != nil {
		return nil, fmt.Errorf("error creating GL program: %v", err)
	}

	// create and alloc hw buf
	wf.buf = ctx.CreateBuffer()
	ctx.BindBuffer(gl.ARRAY_BUFFER, wf.buf)
	ctx.BufferData(gl.ARRAY_BUFFER, make([]byte, len(wf.aligned)*12), gl.STREAM_DRAW)

	wf.position = ctx.GetAttribLocation(wf.program, "position")
	wf.color = ctx.GetUniformLocation(wf.program, "color")
	return wf, nil
}

func (wf *Waveform) Align(amp float64) {
	wf.align = true
	wf.alignamp = amp
}

func (wf *Waveform) Prepare(tc uint64) {
	if wf.Sound != nil {
		wf.Sound.Prepare(tc)
	}

	// cycle outputs
	out := wf.outs[0]
	for i := 0; i+1 < len(wf.outs); i++ {
		wf.outs[i] = wf.outs[i+1]
	}
	for i, x := range wf.Sound.Samples() {
		out[i] = x
	}
	wf.outs[len(wf.outs)-1] = out

	//
	for i, out := range wf.outs {
		idx := i * len(out)
		sl := wf.samples[idx : idx+len(out)]
		for j, x := range out {
			sl[j] = x
		}
	}

	// for i, out := range wf.outs {
	// for j, x := range out {
	// wf.samples[(i+1)*j] = x
	// }
	// }
}

func (wf *Waveform) Paint(ctx gl.Context, xps, yps, width, height float32) {
	// TODO this is racey and samples can be in the middle of changing
	// move the slice copy to Prepare and sync with playback, or feed over chan
	// TODO assumes mono

	var (
		xstep float32 = width / float32(len(wf.samples))
		xpos  float32 = xps
	)

	samples := wf.samples[:]

	// TODO check out https://en.wikipedia.org/wiki/Matched_filter
	// TODO would be nice if Buffer could return samples matching a pattern
	// and get rid of this logic in here
	if wf.align {
		// naive equivalent-time sampling
		// TODO if audio and graphics were disjoint, a proper equiv-time smpl might be all we really need?
		var mt int = -1
		for i, x := range samples {
			if equals(x, wf.alignamp) {
				mt = i
				break
			}
		}

		if mt == -1 {
			log.Println("failed to locate trigger amp")
			return
		}

		for i, x := range samples[mt:] {
			wf.aligned[i] = x
		}
		samples = wf.aligned
	}

	for i, x := range samples {
		// make user aware of clipping, even if inaudible.
		if x > 1 {
			x = 1
		} else if x < -1 {
			x = -1
		}

		wf.verts[i*3] = float32(xpos)
		wf.verts[i*3+1] = yps + (height * float32((x+1)/2))
		wf.verts[i*3+2] = 0
		xpos += xstep
	}

	for i, x := range wf.verts {
		u := math.Float32bits(x)
		wf.data[4*i+0] = byte(u >> 0)
		wf.data[4*i+1] = byte(u >> 8)
		wf.data[4*i+2] = byte(u >> 16)
		wf.data[4*i+3] = byte(u >> 24)
	}

	ctx.UseProgram(wf.program)
	ctx.Uniform4f(wf.color, 1, 1, 1, 1)

	// update hw buf and draw
	ctx.BindBuffer(gl.ARRAY_BUFFER, wf.buf)
	ctx.EnableVertexAttribArray(wf.position)
	ctx.VertexAttribPointer(wf.position, 3, gl.FLOAT, false, 0, 0)
	ctx.BufferSubData(gl.ARRAY_BUFFER, 0, wf.data)
	ctx.DrawArrays(gl.LINE_STRIP, 0, len(wf.aligned))
	ctx.DisableVertexAttribArray(wf.position)
}

const vertexShader = `#version 100
attribute vec4 position;
void main() {
  gl_Position = position;
}`

const fragmentShader = `#version 100
precision mediump float;
uniform vec4 color;
void main() {
  gl_FragColor = color;
}`
