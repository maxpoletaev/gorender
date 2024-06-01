package main

import (
	"image/color"
)

type FrameBuffer struct {
	Width   int
	Height  int
	ZBuffer []float32
	Pixels  []color.RGBA
	Pixels2 []color.RGBA
}

func NewFrameBuffer(width, height int) *FrameBuffer {
	return &FrameBuffer{
		Width:   width,
		Height:  height,
		Pixels:  make([]color.RGBA, width*height),
		Pixels2: make([]color.RGBA, width*height),
		ZBuffer: make([]float32, width*height),
	}
}

func (fb *FrameBuffer) Pixel(x, y int, c color.RGBA) {
	idx := y*fb.Width + x
	if idx > 0 && idx < len(fb.Pixels) {
		fb.Pixels[idx] = c
	}
}

func (fb *FrameBuffer) SwapBuffers() {
	fb.Pixels, fb.Pixels2 = fb.Pixels2, fb.Pixels
}

func (fb *FrameBuffer) Clear(c color.RGBA) {
	fb.ZBuffer[0] = -1.0
	fb.Pixels[0] = c

	for i := 1; i < len(fb.Pixels); i *= 2 {
		copy(fb.Pixels[i:], fb.Pixels[:i])
		copy(fb.ZBuffer[i:], fb.ZBuffer[:i])
	}
}

func (fb *FrameBuffer) DotGrid(c color.RGBA, step int) {
	for y := step; y < fb.Height; y += step {
		for x := step; x < fb.Width; x += step {
			fb.Pixel(x, y, c)
		}
	}
}

func (fb *FrameBuffer) Rect(x, y, width, height int, c color.RGBA) {
	if x >= fb.Width || y >= fb.Height {
		return
	}
	for py := y; py < y+height; py++ {
		for px := x; px < x+width; px++ {
			fb.Pixel(px, py, c)
		}
	}
}

func (fb *FrameBuffer) Line(x0, y0 int, x1, y1 int, c color.RGBA) {
	dx := x1 - x0
	dy := y1 - y0

	sideLength := max(abs(dx), abs(dy))
	xStep := float32(dx) / float32(sideLength)
	yStep := float32(dy) / float32(sideLength)
	curX, curY := float32(x0), float32(y0)

	for i := 0; i <= sideLength; i++ {
		fb.Pixel(int(curX), int(curY), c)
		curX += xStep
		curY += yStep
	}
}

func colorIntensity(c color.RGBA, intensity float32) color.RGBA {
	return color.RGBA{
		R: uint8(float32(c.R) * intensity),
		G: uint8(float32(c.G) * intensity),
		B: uint8(float32(c.B) * intensity),
		A: c.A,
	}
}

func (fb *FrameBuffer) Triangle(
	x0, y0 int, z0 float32, u0, v0 float32,
	x1, y1 int, z1 float32, u1, v1 float32,
	x2, y2 int, z2 float32, u2, v2 float32,
	tileStartX, tileStartY, tileEndX, tileEndY int,
	intensA, intensB, intensC float32,
	texture *Texture,
) {
	// Find the bounding box of the triangle
	minX, maxX := min(x0, x1, x2), max(x0, x1, x2)
	minY, maxY := min(y0, y1, y2), max(y0, y1, y2)

	// Clip the bounding box to the tile boundaries
	minX, maxX = max(minX, tileStartX, 0), min(maxX, tileEndX, fb.Width-1)
	minY, maxY = max(minY, tileStartY, 0), min(maxY, tileEndY, fb.Height-1)

	// Calculate initial edge function values for the first pixel in the bounding box
	f01 := (y0-y1)*minX + (x1-x0)*minY + (x0*y1 - x1*y0)
	f12 := (y1-y2)*minX + (x2-x1)*minY + (x1*y2 - x2*y1)
	f20 := (y2-y0)*minX + (x0-x2)*minY + (x2*y0 - x0*y2)

	// Calculate the change in the edge function values when moving one pixel to the right and down
	f01dx := y0 - y1
	f01dy := x1 - x0
	f12dx := y1 - y2
	f12dy := x2 - x1
	f20dx := y2 - y0
	f20dy := x0 - x2

	// Top-left rule adjustment for the edge functions
	edgeAdjust := func(f, dx, dy int) int {
		if dy > 0 || (dy == 0 && dx > 0) {
			return f
		}
		return f - 1
	}

	f01 = edgeAdjust(f01, f01dx, f01dy)
	f12 = edgeAdjust(f12, f12dx, f12dy)
	f20 = edgeAdjust(f20, f20dx, f20dy)

	// Precalculate factors for uv interpolation
	v0z0 := v0 / z0
	u0z0 := u0 / z0
	u1z1 := u1 / z1
	v1z1 := v1 / z1
	u2z2 := u2 / z2
	v2z2 := v2 / z2

	// Iterate through the bounding box
	for y := minY; y <= maxY; y++ {
		fx01 := f01
		fx12 := f12
		fx20 := f20

		for x := minX; x <= maxX; x++ {
			// Check if the point is inside the triangle using the edge function values
			if fx01 < 0 && fx12 < 0 && fx20 < 0 {
				// Compute barycentric coordinates for x, y
				alpha := float32(fx12) / float32(fx12+fx20+fx01)
				beta := float32(fx20) / float32(fx12+fx20+fx01)
				gamma := 1 - alpha - beta

				zRec := -(alpha/z0 + beta/z1 + gamma/z2)
				index := y*fb.Width + x

				if zRec >= fb.ZBuffer[index] {
					// Interpolate texture coordinates
					u := (alpha*u0z0 + beta*u1z1 + gamma*u2z2) / zRec
					v := (alpha*v0z0 + beta*v1z1 + gamma*v2z2) / zRec

					// Interpolate light intensity
					intensity := alpha*intensA + beta*intensB + gamma*intensC

					c := faceColor
					if texture != nil {
						c = texture.Sample(u, v)
					}

					fb.ZBuffer[index] = zRec
					fb.Pixels[index] = colorIntensity(c, intensity)
				}
			}

			fx01 += f01dx
			fx12 += f12dx
			fx20 += f20dx
		}

		f01 += f01dy
		f12 += f12dy
		f20 += f20dy
	}
}

func blendRGBA(a, b color.RGBA, f float32) color.RGBA {
	cr := uint8(float32(a.R)*(1-f) + float32(b.R)*f)
	cg := uint8(float32(a.G)*(1-f) + float32(b.G)*f)
	cb := uint8(float32(a.B)*(1-f) + float32(b.B)*f)
	ca := uint8(float32(a.A)*(1-f) + float32(b.A)*f)
	return color.RGBA{cr, cg, cb, ca}
}

func (fb *FrameBuffer) Fog(fogStart, fogEnd float32, c color.RGBA) {
	for i := range fb.Pixels {
		depth := fb.ZBuffer[i]

		switch {
		case depth >= fogStart:
			// noop
		case depth <= fogEnd:
			fb.Pixels[i] = c
		default:
			f := 1 - ((fogEnd - depth) / (fogEnd - fogStart))
			fb.Pixels[i] = blendRGBA(fb.Pixels[i], c, f)
		}
	}
}

func (fb *FrameBuffer) CrossHair(c color.RGBA) {
	const size, offset = 5, 3
	x, y := fb.Width/2, fb.Height/2

	fb.Line(x-size, y, x-offset, y, c)
	fb.Line(x+offset, y, x+size, y, c)
	fb.Line(x, y-size, x, y-offset, c)
	fb.Line(x, y+offset, x, y+size, c)
}
