//go:build !canvas
// +build !canvas

package gdi

import (
	"golang.org/x/image/font"

	"github.com/goplus/spx/internal/gdi/text"
)

type Font = font.Face

// -------------------------------------------------------------------------------------

// TextRender represents a text rendering engine.
type TextRender struct {
	*text.Render
}

// -------------------------------------------------------------------------------------
