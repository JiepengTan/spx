package gdi

import (
	"image"

	svgo "github.com/ajstarks/svgo"
)

// -------------------------------------------------------------------------------------

// Sprite type.
type Sprite image.RGBA

// -------------------------------------------------------------------------------------

// Canvas represents a gdi object.
type Canvas struct {
	*svgo.SVG
}

type SVG struct {
	*svgo.SVG
}
