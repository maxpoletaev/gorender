package main

import (
	"image/color"
	"sort"
)

const (
	ZBufferMax = -1000000
)

type FrameBuffer struct {
	Width   int
	Height  int
	Pixels  []color.RGBA
	ZBuffer []float32
}

func NewFrameBuffer(width, height int) *FrameBuffer {
	return &FrameBuffer{
		Width:   width,
		Height:  height,
		Pixels:  make([]color.RGBA, width*height),
		ZBuffer: make([]float32, width*height),
	}
}

func (fb *FrameBuffer) Pixel(x, y int, c color.RGBA, z float32) {
	idx := y*fb.Width + x
	if idx < 0 || idx >= len(fb.Pixels) {
		return
	}

	fb.Pixels[idx] = c
	fb.ZBuffer[idx] = z
}

func (fb *FrameBuffer) Rect(x, y, width, height int, c color.RGBA, z float32) {
	if x >= fb.Width || y >= fb.Height {
		return
	}
	for py := y; py < y+height; py++ {
		for px := x; px < x+width; px++ {
			fb.Pixel(px, py, c, z)
		}
	}
}

func (fb *FrameBuffer) DotGrid(c color.RGBA, step int) {
	for y := step; y < fb.Height; y += step {
		for x := step; x < fb.Width; x += step {
			fb.Pixel(x, y, c, ZBufferMax)
		}
	}
}

func (fb *FrameBuffer) Fill(c color.RGBA) {
	fb.Pixels[0] = c
	fb.ZBuffer[0] = ZBufferMax

	for i := 1; i < len(fb.Pixels); i *= 2 {
		copy(fb.Pixels[i:], fb.Pixels[:i])
		copy(fb.ZBuffer[i:], fb.ZBuffer[:i])
	}
}

func (fb *FrameBuffer) Line(x0, y0 int, x1, y1 int, c color.RGBA, z float32) {
	dx := x1 - x0
	dy := y1 - y0

	dxAbs := dx
	if dxAbs < 0 {
		dxAbs = -dxAbs
	}

	dyAbs := dy
	if dyAbs < 0 {
		dyAbs = -dyAbs
	}

	var sideLength int
	if dxAbs > dyAbs {
		sideLength = dxAbs
	} else {
		sideLength = dyAbs
	}

	xStep := float32(dx) / float32(sideLength)
	yStep := float32(dy) / float32(sideLength)
	curX, curY := float32(x0), float32(y0)

	for i := 0; i <= sideLength; i++ {
		fb.Pixel(int(curX), int(curY), c, z)
		curX += xStep
		curY += yStep
	}
}

func (fb *FrameBuffer) triangleTopHalf(x0, y0 int, x1, y1 int, x2, y2 int, c color.RGBA, z float32) {
	if y0 == y1 {
		return
	}

	slope1 := float32(x1-x0) / float32(y1-y0)
	slope2 := float32(x2-x0) / float32(y2-y0)
	xStart := float32(x0)
	xEnd := float32(x0)

	for y := y0; y <= y1; y++ {
		fb.Line(int(xStart), y, int(xEnd), y, c, z)
		xStart += slope1
		xEnd += slope2
	}
}

func (fb *FrameBuffer) triangleBottomHalf(x0, y0 int, x1, y1 int, x2, y2 int, c color.RGBA, z float32) {
	if y1 == y2 {
		return
	}

	slope1 := float32(x2-x0) / float32(y2-y0)
	slope2 := float32(x2-x1) / float32(y2-y1)
	xStart := float32(x2)
	xEnd := float32(x2)

	for y := y2; y > y0; y-- {
		fb.Line(int(xStart), y, int(xEnd), y, c, z)
		xStart -= slope1
		xEnd -= slope2
	}
}

func (fb *FrameBuffer) Triangle(x0, y0 int, x1, y1 int, x2, y2 int, c color.RGBA, z float32) {
	verts := []struct{ x, y int }{{x0, y0}, {x1, y1}, {x2, y2}}

	sort.Slice(verts, func(i, j int) bool {
		return verts[i].y < verts[j].y
	})

	// Sorted points
	x0, y0 = verts[0].x, verts[0].y
	x1, y1 = verts[1].x, verts[1].y
	x2, y2 = verts[2].x, verts[2].y

	// Midpoint
	my := verts[1].y
	mx := int(float32((x2-x0)*(y1-y0))/float32(y2-y0)) + x0

	// Flat-bottom/flat-top algorithm
	fb.triangleTopHalf(x0, y0, x1, y1, mx, my, c, z)
	fb.triangleBottomHalf(x1, y1, mx, my, x2, y2, c, z)
}
