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
	xStep := float64(dx) / float64(sideLength)
	yStep := float64(dy) / float64(sideLength)
	curX, curY := float64(x0), float64(y0)

	for i := 0; i <= sideLength; i++ {
		fb.Pixel(int(curX), int(curY), c)
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
	tileStartX, tileStartY, tileEndX, tileEndY int,
	texture *Texture,
	intensity float64,
) {
	// Find the bounding box of the triangle
	minX, maxX := min(x0, x1, x2), max(x0, x1, x2)
	minY, maxY := min(y0, y1, y2), max(y0, y1, y2)

	// Clip the bounding box to the tile boundaries
	minX, maxX = max(minX, tileStartX), min(maxX, tileEndX)
	minY, maxY = max(minY, tileStartY), min(maxY, tileEndY)

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
			// Compute barycentric coordinates for x, y
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

			zRec := -(alphaZ + betaZ + gammaZ)
			index := y*fb.Width + x

			if zRec >= fb.ZBuffer[index] {
				// Interpolate the texture coordinates
				u := (alphaZ*u0 + betaZ*u1 + gammaZ*u2) / zRec
				v := (alphaZ*v0 + betaZ*v1 + gammaZ*v2) / zRec

				c := faceColor
				if texture != nil {
					c = texture.Sample(u, v)
				}

				c = colorIntensity(c, intensity)
				fb.ZBuffer[index] = zRec
				fb.Pixels[index] = c
			}
		}
	}
}

func (fb *FrameBuffer) Triangle2(
	x0, y0 int, z0 float64, u0, v0 float64,
	x1, y1 int, z1 float64, u1, v1 float64,
	x2, y2 int, z2 float64, u2, v2 float64,
	tileStartX, tileStartY, tileEndX, tileEndY int,
	texture *Texture,
	intensity float64,
) {
	// Find the bounding box of the triangle
	minX, maxX := min(x0, x1, x2), max(x0, x1, x2)
	minY, maxY := min(y0, y1, y2), max(y0, y1, y2)

	// Clip the bounding box to the tile boundaries
	minX, maxX = max(minX, tileStartX, 0), min(maxX, tileEndX, fb.Width-1)
	minY, maxY = max(minY, tileStartY, 0), min(maxY, tileEndY, fb.Height-1)

	// Ensure the intensity is in the range [0.2, 1.0]
	intensity = max(0.2, min(intensity, 1.0))

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

	// Apply top-left rule adjustment for the edge functions
	edgeAdjust := func(f, dx, dy int) int {
		if dy > 0 || (dy == 0 && dx > 0) {
			return f
		}
		return f - 1
	}

	// Adjust the initial edge function values based on the top-left rule
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
				alpha := float64(fx12) / float64(fx12+fx20+fx01)
				beta := float64(fx20) / float64(fx12+fx20+fx01)
				gamma := 1 - alpha - beta

				zRec := -(alpha/z0 + beta/z1 + gamma/z2)
				index := y*fb.Width + x

				if zRec >= fb.ZBuffer[index] {
					// Interpolate texture coordinates
					u := (alpha*u0z0 + beta*u1z1 + gamma*u2z2) / zRec
					v := (alpha*v0z0 + beta*v1z1 + gamma*v2z2) / zRec

					c := faceColor
					if texture != nil {
						c = texture.Sample(u, v)
					}

					c = colorIntensity(c, intensity)
					fb.ZBuffer[index] = zRec
					fb.Pixels[index] = c
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

func interpolate(
	y, y0, y1 int,
	x0, x1 int,
	z0rec, z1rec float64,
	u0, u1, v0, v1 float64,
) (int, float64, float64, float64) {
	t := float64(y-y0) / float64(y1-y0)
	x := int(float64(x0) + t*float64(x1-x0))
	z := z0rec + t*(z1rec-z0rec)
	u := u0 + t*(u1-u0)
	v := v0 + t*(v1-v0)
	return x, z, u, v
}

func (fb *FrameBuffer) Triangle3(
	x0, y0 int, z0 float64, u0, v0 float64,
	x1, y1 int, z1 float64, u1, v1 float64,
	x2, y2 int, z2 float64, u2, v2 float64,
	tileStartX, tileStartY, tileEndX, tileEndY int,
	texture *Texture,
	intensity float64,
) {
	// Sort vertices by y-coordinate
	if y0 > y1 {
		x0, y0, z0, x1, y1, z1 = x1, y1, z1, x0, y0, z0
		u0, v0, u1, v1 = u1, v1, u0, v0
	}
	if y0 > y2 {
		x0, y0, z0, x2, y2, z2 = x2, y2, z2, x0, y0, z0
		u0, v0, u2, v2 = u2, v2, u0, v0
	}
	if y1 > y2 {
		x1, y1, z1, x2, y2, z2 = x2, y2, z2, x1, y1, z1
		u1, v1, u2, v2 = u2, v2, u1, v1
	}

	// Ensure the intensity is in the range [0.2, 1.0]
	intensity = max(0.2, min(intensity, 1.0))

	// Calculate reciprocals for depth and texture mapping
	z0rec, z1rec, z2rec := 1/z0, 1/z1, 1/z2

	// Draw scanlines from y0 to y1 (top half of the triangle)
	if y1 > y0 {
		for y := y0; y <= y1; y++ {
			if y < tileStartY || y >= tileEndY {
				continue
			}

			xStart, zStart, uStart, vStart := interpolate(y, y0, y1, x0, x1, z0rec, z1rec, u0, u1, v0, v1)
			xEnd, zEnd, uEnd, vEnd := interpolate(y, y0, y2, x0, x2, z0rec, z2rec, u0, u2, v0, v2)

			if xStart > xEnd {
				xStart, xEnd, zStart, zEnd = xEnd, xStart, zEnd, zStart
				uStart, uEnd, vStart, vEnd = uEnd, uStart, vEnd, vStart
			}

			for x := xStart; x <= xEnd; x++ {
				if x < tileStartX || x >= tileEndX {
					continue
				}

				t := float64(x-xStart) / float64(xEnd-xStart)
				z := -(zStart + t*(zEnd-zStart))
				index := y*fb.Width + x

				if z >= fb.ZBuffer[index] {
					u := uStart + t*(uEnd-uStart)
					v := vStart + t*(vEnd-vStart)

					c := faceColor
					if texture != nil {
						c = texture.Sample(u, v)
					}

					c = colorIntensity(c, intensity)
					fb.ZBuffer[index] = z
					fb.Pixels[index] = c
				}
			}
		}
	}

	// Draw scanlines from y1 to y2 (bottom half of the triangle)
	if y2 > y1 {
		for y := y1; y <= y2; y++ {
			if y < tileStartY || y >= tileEndY {
				continue
			}

			xStart, zStart, uStart, vStart := interpolate(y, y1, y2, x1, x2, z1rec, z2rec, u1, u2, v1, v2)
			xEnd, zEnd, uEnd, vEnd := interpolate(y, y0, y2, x0, x2, z0rec, z2rec, u0, u2, v0, v2)

			if xStart > xEnd {
				xStart, xEnd, zStart, zEnd = xEnd, xStart, zEnd, zStart
				uStart, uEnd, vStart, vEnd = uEnd, uStart, vEnd, vStart
			}

			for x := xStart; x <= xEnd; x++ {
				if x < tileStartX || x >= tileEndX {
					continue
				}

				t := float64(x-xStart) / float64(xEnd-xStart)
				z := -(zStart + t*(zEnd-zStart))
				index := y*fb.Width + x

				if z >= fb.ZBuffer[index] {
					u := uStart + t*(uEnd-uStart)
					v := vStart + t*(vEnd-vStart)

					c := faceColor
					if texture != nil {
						c = texture.Sample(u, v)
					}

					c = colorIntensity(c, intensity)
					fb.ZBuffer[index] = z
					fb.Pixels[index] = c
				}
			}
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

func (fb *FrameBuffer) CrossHair(c color.RGBA) {
	const size, offset = 5, 3
	x, y := fb.Width/2, fb.Height/2

	fb.Line(x-size, y, x-offset, y, c)
	fb.Line(x+offset, y, x+size, y, c)
	fb.Line(x, y-size, x, y-offset, c)
	fb.Line(x, y+offset, x, y+size, c)
}
