package x11draw

import (
	"image"
	"image/color"
)

var (
	BGRAModel = color.ModelFunc(bgraModel)
)

func bgraModel(c color.Color) color.Color {
	if _, ok := c.(BGRAColor); ok {
		return c
	}
	r, g, b, a := c.RGBA()
	return BGRAColor{uint8(b >> 8), uint8(g >> 8), uint8(r >> 8), uint8(a >> 8)}
}

// RGBA represents a traditional 32-bit alpha-premultiplied color, having 8
// bits for each of red, green, blue and alpha.
//
// An alpha-premultiplied color component C has been scaled by alpha (A), so
// has valid values 0 <= C <= A.
type BGRAColor struct {
	B, G, R, A uint8
}

func (c BGRAColor) RGBA() (r, g, b, a uint32) {
	b = uint32(c.B)
	b |= b << 8
	g = uint32(c.G)
	g |= g << 8
	r = uint32(c.R)
	r |= r << 8
	a = uint32(c.A)
	a |= a << 8
	return
}

// BGRA is an in-memory image whose At method returns [color.BGRA] values.
type BGRA struct {
	// Pix holds the image's pixels, in B, G, R, A order. The pixel at
	// (x, y) starts at Pix[(y-Rect.Min.Y)*Stride + (x-Rect.Min.X)*4].
	Pix []uint8
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
}

func (p *BGRA) ColorModel() color.Model { return color.ModelFunc(bgraModel) }

func (p *BGRA) Bounds() image.Rectangle { return p.Rect }

func (p *BGRA) At(x, y int) color.Color {
	return p.BGRAAt(x, y)
}

func (p *BGRA) BGRAAt(x, y int) BGRAColor {
	if !(image.Point{x, y}.In(p.Rect)) {
		return BGRAColor{}
	}
	i := p.PixOffset(x, y)
	s := p.Pix[i : i+4 : i+4] // Small cap improves performance, see https://golang.org/issue/27857
	return BGRAColor{s[0], s[1], s[2], s[3]}
}

func (p *BGRA) RGBAAt(x, y int) color.RGBA {
	if !(image.Point{x, y}.In(p.Rect)) {
		return color.RGBA{}
	}
	i := p.PixOffset(x, y)
	s := p.Pix[i : i+4 : i+4] // Small cap improves performance, see https://golang.org/issue/27857
	return color.RGBA{s[2], s[1], s[0], s[3]}
}

// PixOffset returns the index of the first element of Pix that corresponds to
// the pixel at (x, y).
func (p *BGRA) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*4
}

func (p *BGRA) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.PixOffset(x, y)
	c1 := BGRAModel.Convert(c).(BGRAColor)
	s := p.Pix[i : i+4 : i+4] // Small cap improves performance, see https://golang.org/issue/27857
	s[0] = c1.B
	s[1] = c1.G
	s[2] = c1.R
	s[3] = c1.A
}

func (p *BGRA) SetBGRA64(x, y int, c color.RGBA64) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.PixOffset(x, y)
	s := p.Pix[i : i+4 : i+4] // Small cap improves performance, see https://golang.org/issue/27857
	s[2] = uint8(c.R >> 8)
	s[1] = uint8(c.G >> 8)
	s[1] = uint8(c.B >> 8)
	s[3] = uint8(c.A >> 8)
}

/*
func (p *BGRA) SetBGRA(x, y int, c color.BGRA) {
	if !(Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.PixOffset(x, y)
	s := p.Pix[i : i+4 : i+4] // Small cap improves performance, see https://golang.org/issue/27857
	s[0] = c.R
	s[1] = c.G
	s[2] = c.B
	s[3] = c.A
}
*/

// SubImage returns an image representing the portion of the image p visible
// through r. The returned value shares pixels with the original image.
func (p *BGRA) SubImage(r image.Rectangle) image.Image {
	r = r.Intersect(p.Rect)
	// If r1 and r2 are Rectangles, r1.Intersect(r2) is not guaranteed to be inside
	// either r1 or r2 if the intersection is empty. Without explicitly checking for
	// this, the Pix[i:] expression below can panic.
	if r.Empty() {
		return &BGRA{}
	}
	i := p.PixOffset(r.Min.X, r.Min.Y)
	return &BGRA{
		Pix:    p.Pix[i:],
		Stride: p.Stride,
		Rect:   r,
	}
}

// Opaque scans the entire image and reports whether it is fully opaque.
func (p *BGRA) Opaque() bool {
	if p.Rect.Empty() {
		return true
	}
	i0, i1 := 3, p.Rect.Dx()*4
	for y := p.Rect.Min.Y; y < p.Rect.Max.Y; y++ {
		for i := i0; i < i1; i += 4 {
			if p.Pix[i] != 0xff {
				return false
			}
		}
		i0 += p.Stride
		i1 += p.Stride
	}
	return true
}

// NewBGRA returns a new [BGRA] image with the given bounds.
func NewBGRA(r image.Rectangle) *BGRA {
	return &BGRA{
		Pix:    make([]uint8, r.Dx()*r.Dy()*4),
		Stride: 4 * r.Dx(),
		Rect:   r,
	}
}

// Scroll implements gfx.Scroller
func (p *BGRA) Scroll(amount int) {
	switch {
	case amount == 0:
		return
	case amount > 0:
		if amount > p.Rect.Dy() {
			amount = p.Rect.Dy()
		}
		copy(p.Pix, p.Pix[p.Stride*amount:])
	case amount < 0:
		amount *= -1
		if amount > p.Rect.Dy() {
			amount = p.Rect.Dy()
		}
		reverseCopy(p.Pix[p.Stride*amount:], p.Pix[:len(p.Pix)-p.Stride*amount])
	}
}

func (p *BGRA) RegionScroll(region image.Rectangle, amount int) {
	region = p.Rect.Intersect(region)
	if region.Empty() || amount == 0 {
		return
	}
	// if amount is positive or negative, copy lines forwards or backwards

	var start, end int
	if amount > 0 {
		for y := region.Min.Y; y < (region.Max.Y - amount); y++ {
			start = p.Stride*y + region.Min.X*4
			end = p.Stride*y + region.Max.X*4

			copy(p.Pix[start:end], p.Pix[start+amount*p.Stride:end+amount*p.Stride])
		}
		return
	}

	// negative scrolling (scrolling up)
	for y := region.Max.Y; y > (region.Min.Y - amount); y-- {
		start = p.Stride*y + region.Min.X*4
		end = p.Stride*y + region.Max.X*4

		copy(p.Pix[start:end], p.Pix[start+amount*p.Stride:end+amount*p.Stride])
	}
}

// Fill implements gfx.Filler. Whereever BGRA overlaps with 'where', set those
// pixels to color c.
func (p *BGRA) Fill(where image.Rectangle, c color.Color) {
	// get c as native color
	nc := bgraModel(c).(BGRAColor)

	where = p.Bounds().Intersect(where)

	if where.Empty() {
		return
	}

	pixLine := make([]uint8, 4*where.Bounds().Dx())
	for i := 0; i <= where.Bounds().Dx(); i++ {
		copy(pixLine[i:i+4:i+4], []uint8{nc.B, nc.G, nc.R, nc.A})
	}

	// first try a naïve for loop--we'll optimize later
	for y := where.Min.Y; y < where.Max.Y; y++ {
		i := p.PixOffset(where.Bounds().Min.X, y)
		copy(p.Pix[i:i+len(pixLine):i+len(pixLine)], pixLine)
	}
}

func reverseCopy[E any](dst, src []E) {
	for i := len(src) - 1; i >= 0; i-- {
		dst[i] = src[i]
	}
}
