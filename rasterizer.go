package main

import (
	"image/color"
)

type FrameBuffer struct {
	Width   int
	Height  int
	ZBuffer []float64
	Pixels  []color.RGBA
	Pixels2 []color.RGBA
}

func NewFrameBuffer(width, height int) *FrameBuffer {
	return &FrameBuffer{
		Width:   width,
		Height:  height,
		Pixels:  make([]color.RGBA, width*height),
		Pixels2: make([]color.RGBA, width*height),
		ZBuffer: make([]float64, width*height),
	}
}

func (fb *FrameBuffer) Pixel(x, y int, c color.RGBA, z float64) {
	idx := y*fb.Width + x

	if idx > 0 && idx < len(fb.Pixels) && z >= fb.ZBuffer[idx] {
		fb.Pixels[idx] = c
		fb.ZBuffer[idx] = z
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
			fb.Pixel(x, y, c, -1.0)
		}
	}
}

func (fb *FrameBuffer) Rect(x, y, width, height int, c color.RGBA) {
	if x >= fb.Width || y >= fb.Height {
		return
	}
	for py := y; py < y+height; py++ {
		for px := x; px < x+width; px++ {
			fb.Pixel(px, py, c, 1.0)
		}
	}
}

func (fb *FrameBuffer) Line(x0, y0 int, x1, y1 int, c color.RGBA) {
	dx := x1 - x0
	dy := y1 - y0

	sideLength := max(abs(dx), abs(dy))
	xStep := float64(dx) / float64(sideLength)
	yStep := float64(dy) / float64(sideLength)
	curX, curY := float64(x0), float64(y0)

	for i := 0; i <= sideLength; i++ {
		fb.Pixel(int(curX), int(curY), c, 1.0)
		curX += xStep
		curY += yStep
	}
}

func colorIntensity(c color.RGBA, intensity float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) * intensity),
		G: uint8(float64(c.G) * intensity),
		B: uint8(float64(c.B) * intensity),
		A: c.A,
	}
}

func (fb *FrameBuffer) Triangle(
	x0, y0 int, z0 float64, u0, v0 float64,
	x1, y1 int, z1 float64, u1, v1 float64,
	x2, y2 int, z2 float64, u2, v2 float64,
	texture *Texture,
	intensity float64,
) {
	// Find the bounding box of the triangle
	minX, maxX := min(x0, x1, x2), max(x0, x1, x2)+1
	minY, maxY := min(y0, y1, y2), max(y0, y1, y2)+1

	// Ensure the intensity is in the range [0.2, 1.0]
	intensity = max(0.2, min(intensity, 1.0))

	// Find the area of the triangle
	area := float64((y1-y2)*(x0-x2) + (x2-x1)*(y0-y2))
	areaRec := 1 / area

	// Precalculate factors for barycentric coordinates
	var (
		f1 = y1 - y2
		f2 = x2 - x1
		f3 = y2 - y0
		f4 = x0 - x2
	)

	// Calculate the barycentric coordinate deltas
	alphaDx := float64(f1) * areaRec
	betaDx := float64(f3) * areaRec

	// Depth reciprocals for texture mapping
	z0rec, z1rec, z2rec := 1/z0, 1/z1, 1/z2

	// Iterate through the bounding box and check if the pixel
	// is inside the triangle using barycentric coordinates.
	for y := minY; y <= maxY; y++ {
		alphaRow := float64(f1*(minX-x2)+f2*(y-y2)) * areaRec
		betaRow := float64(f3*(minX-x2)+f4*(y-y2)) * areaRec

		for x := minX; x <= maxX; x++ {
			alpha := alphaRow + alphaDx*float64(x-minX)
			beta := betaRow + betaDx*float64(x-minX)
			gamma := 1 - alpha - beta

			// Skip rendering if pixel is outside the triangle
			if alpha < 0 || beta < 0 || gamma < 0 {
				continue
			}

			alphaZ := alpha * z0rec
			betaZ := beta * z1rec
			gammaZ := gamma * z2rec

			// Interpolate the texture coordinates
			zRec := alphaZ + betaZ + gammaZ
			u := (alphaZ*u0 + betaZ*u1 + gammaZ*u2) / zRec
			v := (alphaZ*v0 + betaZ*v1 + gammaZ*v2) / zRec

			c := texture.Sample(u, v)
			c = colorIntensity(c, intensity)

			fb.Pixel(x, y, c, -zRec)
		}
	}
}

func blendRGBA(a, b color.RGBA, f float64) color.RGBA {
	cr := uint8(float64(a.R)*(1-f) + float64(b.R)*f)
	cg := uint8(float64(a.G)*(1-f) + float64(b.G)*f)
	cb := uint8(float64(a.B)*(1-f) + float64(b.B)*f)
	ca := uint8(float64(a.A)*(1-f) + float64(b.A)*f)
	return color.RGBA{cr, cg, cb, ca}
}

func (fb *FrameBuffer) Fog(fogStart, fogEnd float64, c color.RGBA) {
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
