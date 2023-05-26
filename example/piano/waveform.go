package main

import (
	"dasa.cc/material"
	"dasa.cc/material/glutil"
	"dasa.cc/snd"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/gl"
)

type Waveform struct {
	*material.Material
	snd.Sound

	prg  glutil.Program
	vbuf glutil.FloatBuffer
	ibuf glutil.UintBuffer
	uc   gl.Uniform
	ap   gl.Attrib

	uw, uv, up gl.Uniform

	verts   []float32
	indices []uint32
	outs    [][]float64
	samples []float64
}

// TODO just how many samples do we want/need to display something useful?
func NewWaveform(ctx gl.Context, n int, in snd.Sound) (*Waveform, error) {
	wf := &Waveform{Material: env.NewMaterial(ctx), Sound: in}
	wf.Drawer = wf.Draw

	wf.outs = make([][]float64, n)
	for i := range wf.outs {
		wf.outs[i] = make([]float64, len(in.Samples())*in.Channels())
	}
	wf.samples = make([]float64, len(in.Samples())*in.Channels()*n)
	wf.verts = make([]float32, len(wf.samples)*2)

	wf.indices = make([]uint32, len(wf.verts))
	for i := range wf.indices {
		wf.indices[i] = uint32((i + 1) / 2)
	}
	wf.indices[len(wf.indices)-1] = wf.indices[len(wf.indices)-2]

	wf.prg.CreateAndLink(ctx,
		glutil.ShaderAsset(gl.VERTEX_SHADER, "basic-vert.glsl"),
		glutil.ShaderAsset(gl.FRAGMENT_SHADER, "basic-frag.glsl"))

	wf.vbuf = glutil.NewFloatBuffer(ctx, wf.verts, gl.STREAM_DRAW)
	wf.ibuf = glutil.NewUintBuffer(ctx, wf.indices, gl.STATIC_DRAW)
	wf.uw = wf.prg.Uniform(ctx, "world")
	wf.uv = wf.prg.Uniform(ctx, "view")
	wf.up = wf.prg.Uniform(ctx, "proj")
	wf.uc = wf.prg.Uniform(ctx, "color")
	wf.ap = wf.prg.Attrib(ctx, "position")
	return wf, nil
}

func (wf *Waveform) Prepare(tc uint64) {
	wf.Sound.Prepare(tc)

	// cycle outputs
	out := wf.outs[0]
	for i := 0; i+1 < len(wf.outs); i++ {
		wf.outs[i] = wf.outs[i+1]
	}
	for i, x := range wf.Sound.Samples() {
		out[i] = x
	}
	wf.outs[len(wf.outs)-1] = out

	// sample
	for i, out := range wf.outs {
		idx := i * len(out)
		sl := wf.samples[idx : idx+len(out)]
		for j, x := range out {
			sl[j] = x
		}
	}
}

func (wf *Waveform) Draw(ctx gl.Context, view, proj f32.Mat4) {
	world := wf.World()
	xstep := float32(1) / float32(len(wf.samples))
	xpos := float32(0)

	// TODO assumes mono
	for i, x := range wf.samples {
		// clip
		if x > 1 {
			x = 1
		} else if x < -1 {
			x = -1
		}

		wf.verts[i*2] = float32(xpos)
		wf.verts[i*2+1] = float32(x+1) / 2
		xpos += xstep
	}

	ctx.LineWidth(2)
	wf.prg.Use(ctx)
	wf.prg.Mat4(ctx, wf.uw, *world)
	wf.prg.Mat4(ctx, wf.uv, view)
	wf.prg.Mat4(ctx, wf.up, proj)
	r, g, b, a := env.Palette().Accent.RGBA()
	wf.prg.U4f(ctx, wf.uc, r, g, b, a)
	wf.vbuf.Bind(ctx)
	wf.vbuf.Update(ctx, wf.verts)
	wf.ibuf.Bind(ctx)
	wf.prg.Pointer(ctx, wf.ap, 2)
	wf.ibuf.Draw(ctx, wf.prg, gl.LINES)
}
