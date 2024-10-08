package gdi

import (
	"image"
	"image/draw"
)

type Image struct {
	ebiImg *image.Rectangle
	img    *image.RGBA
}

func NewImageSize(width, height int) Image {
	return Image{
		ebiImg: &image.Rectangle{Max: image.Point{X: width, Y: height}},
		img:    nil,
	}
}

func NewImageFrom(img image.Image) Image {
	rgba, ok := img.(*image.RGBA)
	if !ok {
		bounds := img.Bounds()
		bounds.Sub(bounds.Min)
		rgba = image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, img, img.Bounds().Min, draw.Src)
	}
	return Image{&rgba.Rect, rgba}
}

func (i Image) Origin() *image.RGBA {
	return i.img
}

func (i Image) IsValid() bool {
	return i.ebiImg != nil
}

func (i Image) Bounds() image.Rectangle {
	return i.ebiImg.Bounds()
}

func (i Image) Size() (width, height int) {
	return i.ebiImg.Max.X, i.ebiImg.Max.Y
}

func (i Image) SubImage(rect image.Rectangle) Image {
	panic("not implemented")
	return Image{}
}
