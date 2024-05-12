package main

import (
	"image/color"
)

type FrameBuffer struct {
	Width  int
	Height int
	Pixels []color.RGBA
}

func NewFrameBuffer(width, height int) *FrameBuffer {
	return &FrameBuffer{
		Width:  width,
		Height: height,
		Pixels: make([]color.RGBA, width*height),
	}
}

func (fb *FrameBuffer) Pixel(x, y int, c color.RGBA) {
	idx := y*fb.Width + x
	if idx < 0 || idx >= len(fb.Pixels) {
		return
	}

	fb.Pixels[idx] = c
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

func (fb *FrameBuffer) DotGrid(c color.RGBA, step int) {
	for y := step; y < fb.Height; y += step {
		for x := step; x < fb.Width; x += step {
			fb.Pixel(x, y, c)
		}
	}
}

func (fb *FrameBuffer) Clear(c color.RGBA) {
	fb.Pixels[0] = c

	for i := 1; i < len(fb.Pixels); i *= 2 {
		copy(fb.Pixels[i:], fb.Pixels[:i])
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
		fb.Pixel(int(curX), int(curY), c)
		curX += xStep
		curY += yStep
	}
}

func barycentric(area float64, x0, y0, x1, y1, x2, y2, px, py int) (float64, float64, float64) {
	w0 := float64((y1-y2)*(px-x2)+(x2-x1)*(py-y2)) / area
	w1 := float64((y2-y0)*(px-x2)+(x0-x2)*(py-y2)) / area
	w2 := 1 - w0 - w1
	return w0, w1, w2
}

func (fb *FrameBuffer) TexturedTriangle(
	x0, y0 int, u0, v0 float64,
	x1, y1 int, u1, v1 float64,
	x2, y2 int, u2, v2 float64,
	texture Texture,
	intensity float64,
) {
	// Find the bounding box of the triangle
	minX, maxX := min(x0, x1, x2), max(x0, x1, x2)
	minY, maxY := min(y0, y1, y2), max(y0, y1, y2)

	// Ensure the intensity is in the range [0, 1]
	intensity = max(0.1, min(intensity, 0.95))

	// Find the area of the triangle
	area := float64((y1-y2)*(x0-x2) + (x2-x1)*(y0-y2))

	// Iterate through the bounding box and check if the pixel is inside the triangle
	// using barycentric coordinates (if any of the weights is negative, the pixel is
	// outside the triangle).
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			w0, w1, w2 := barycentric(area, x0, y0, x1, y1, x2, y2, x, y)
			if w0 < 0 || w1 < 0 || w2 < 0 {
				continue
			}

			// Interpolate the texture coordinates
			u := w0*u0 + w1*u1 + w2*u2
			v := w0*v0 + w1*v1 + w2*v2

			c := texture.Sample(u, v)
			fb.Pixel(x, y, colorIntensity(c, intensity))
		}
	}
}
