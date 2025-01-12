package drawtool

import (
	"image"
	"image/color"
	"image/draw"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// Define a white color
var colWhite = color.RGBA{255, 255, 255, 255}

// DrawTool holds the point and drawer needed for text drawing.
type DrawTool struct {
	point  fixed.Point26_6
	drawer font.Drawer
}

// NewDrawTool initializes and returns a new DrawTool instance.
func NewDrawTool(img *image.RGBA, face font.Face) *DrawTool {
	// Set the starting point for drawing text
	point := fixed.Point26_6{
		X: fixed.I(0),
		Y: fixed.I(0),
	}

	// Configure the font drawer with the chosen font and color
	drawer := font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{C: colWhite},
		Face: face,
		Dot:  point,
	}

	return &DrawTool{
		point:  point,
		drawer: drawer,
	}
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

// Bresenham draws a line between (x0, y0) and (x1, y1)
func Bresenham(img draw.Image, color color.Color, x0, y0, x1, y1 int) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	var sx, sy int
	if x0 < x1 {
		sx = 1
	} else {
		sx = -1
	}
	if y0 < y1 {
		sy = 1
	} else {
		sy = -1
	}
	err := dx - dy

	var e2 int
	for {
		img.Set(x0, y0, color)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 = 2 * err
		if e2 > -dy {
			err = err - dy
			x0 = x0 + sx
		}
		if e2 < dx {
			err = err + dx
			y0 = y0 + sy
		}
	}
}

// DrawString draws the provided text at the current point using the configured drawer.
func (dt *DrawTool) DrawString(text string) {
	dt.drawer.DrawString(text)
}

// SetPoint allows setting a new drawing point.
func (dt *DrawTool) SetPoint(x, y int) {
	dt.point = fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}
	dt.drawer.Dot = dt.point
}
