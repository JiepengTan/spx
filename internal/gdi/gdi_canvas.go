//go:build canvas
// +build canvas

package gdi

import (
	"github.com/goplus/canvas"
	"github.com/goplus/spx/internal/gdi/font"
	"github.com/hajimehoshi/ebiten/v2"
)

type Font = font.Font

// -------------------------------------------------------------------------------------

// TextRender represents a text rendering engine.
type TextRender struct {
	fnt      *canvas.Font
	tm       *TextMetrics
	img      *ebiten.Image
	maxWidth int
	width    int
	height   int
	dy       int
	dirty    bool
	text     string
	lines    []string
	Scale    float64
}

// -------------------------------------------------------------------------------------

type TextMetrics struct {
	ctx   canvas.Context2D
	cache map[rune]float64
}
