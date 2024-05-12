package main

import (
	"image/color"
)

type Texture interface {
	Sample(u, v float64) color.RGBA
}

type ImageTexture struct {
	Width, Height int
	Pixels        []color.RGBA
}

func (t *ImageTexture) Sample(u, v float64) color.RGBA {
	x := (int(u*float64(t.Width)) + t.Width) % t.Width
	y := (int(v*float64(t.Height)) + t.Height) % t.Height
	return t.Pixels[y*t.Width+x]
}

type SolidTexture struct {
	Color color.RGBA
}

func (t *SolidTexture) Sample(u, v float64) color.RGBA {
	return t.Color
}
