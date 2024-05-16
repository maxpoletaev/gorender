package main

import (
	"image/color"
)

type Texture struct {
	Width, Height int
	Color         color.RGBA
	Pixels        []color.RGBA
}

func (t *Texture) Sample(u, v float64) color.RGBA {
	if t.Pixels == nil {
		return t.Color
	}

	x := abs(int(u*float64(t.Width))+t.Width) % t.Width
	y := abs(int(v*float64(t.Height))+t.Height) % t.Height
	px := t.Pixels[y*t.Width+x]

	if px.A == 0 {
		return t.Color
	}

	return px
}
