package main

import (
	"errors"
	"image"
	"image/color"
	"os"
)

type TextureType int

const (
	TextureTypeSolidColor TextureType = iota
	TextureTypeImage
)

type Texture struct {
	width, height   int
	widthF, heightF float64
	scale           float64
	color           color.RGBA
	pixels          []color.RGBA
	typ             TextureType
}

func NewColorTexture(c color.RGBA) *Texture {
	return &Texture{
		typ:   TextureTypeSolidColor,
		color: c,
	}
}

func NewImageTexture(img image.Image) (*Texture, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if !isPowerOfTwo(width) || !isPowerOfTwo(height) {
		return nil, errors.New("texture size is not a power of two")
	}

	t := &Texture{
		width:   width,
		height:  height,
		widthF:  float64(width),
		heightF: float64(height),
		pixels:  make([]color.RGBA, bounds.Dx()*bounds.Dy()),
		typ:     TextureTypeImage,
		scale:   1.0,
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
			t.pixels[y*width+x] = c
		}
	}

	return t, nil
}

func (t *Texture) SetScale(scale float64) {
	t.scale = scale
}

func (t *Texture) Sample(u, v float64) color.RGBA {
	switch t.typ {
	case TextureTypeSolidColor:
		return t.color
	case TextureTypeImage:
		x := int(u*t.scale*t.widthF) & (t.width - 1)
		y := int((1-v)*t.scale*t.heightF) & (t.height - 1)
		return t.pixels[y*t.width+x]
	default:
		return color.RGBA{255, 0, 255, 255}
	}
}

func LoadTextureFile(filename string) (*Texture, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return NewImageTexture(img)
}
