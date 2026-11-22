package main

import (
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"math"
	"strconv"
	"strings"

	"golang.org/x/image/vector"
)

// drawing is the mark as assets/mark.svg draws it: filled shapes, back to front, in a square.
type drawing struct {
	side   float64 // the viewBox's side; its corner is 0 0
	shapes []shape
}

// shape is one filled outline. A path is a list of closed curves; every curve is cubics, a line
// being the cubic whose handles sit on its ends, so drawing and writing have one case.
type shape struct {
	fill   string // #RRGGBB, as written
	curves [][]cubic
}

type point struct{ x, y float64 }

type cubic struct{ p0, p1, p2, p3 point }

// readDrawing reads the part of SVG the mark is drawn in: <path> with absolute M, L, C and Z, and
// <ellipse>, each with a fill. Anything else fails. An element or attribute this does not draw
// (a stroke, a transform, a gradient an editor added) would be in the source and in none of the
// logos, and the two would differ with nothing saying so.
func readDrawing(r io.Reader) (drawing, error) {
	var d drawing
	dec := xml.NewDecoder(r)
	depth := 0
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return d, err
		}
		switch t := tok.(type) {
		case xml.EndElement:
			depth--
		case xml.StartElement:
			depth++
			attrs := map[string]string{}
			for _, a := range t.Attr {
				if a.Name.Space == "xmlns" || a.Name.Local == "xmlns" {
					continue
				}
				attrs[a.Name.Local] = a.Value
			}
			switch {
			case depth == 1 && t.Name.Local == "svg":
				if d.side, err = squareOf(attrs["viewBox"]); err != nil {
					return d, err
				}
				delete(attrs, "viewBox")
			case depth == 2 && t.Name.Local == "path":
				s := shape{fill: attrs["fill"]}
				if s.curves, err = curvesOf(attrs["d"]); err != nil {
					return d, fmt.Errorf("path %q: %w", attrs["d"], err)
				}
				d.shapes = append(d.shapes, s)
				delete(attrs, "fill")
				delete(attrs, "d")
			case depth == 2 && t.Name.Local == "ellipse":
				var v [4]float64
				for i, k := range []string{"cx", "cy", "rx", "ry"} {
					if v[i], err = strconv.ParseFloat(attrs[k], 64); err != nil {
						return d, fmt.Errorf("ellipse: %s %q is no number", k, attrs[k])
					}
					delete(attrs, k)
				}
				d.shapes = append(d.shapes, shape{fill: attrs["fill"], curves: [][]cubic{ellipse(v[0], v[1], v[2], v[3])}})
				delete(attrs, "fill")
			default:
				return d, fmt.Errorf("<%s> at depth %d: the mark is <path> and <ellipse> directly in <svg>", t.Name.Local, depth)
			}
			for k := range attrs {
				return d, fmt.Errorf("<%s %s>: an attribute the logos would not carry", t.Name.Local, k)
			}
		}
	}
	if len(d.shapes) == 0 {
		return d, fmt.Errorf("it draws nothing")
	}
	return d, nil
}

func squareOf(viewBox string) (float64, error) {
	f := strings.Fields(viewBox)
	if len(f) != 4 || f[0] != "0" || f[1] != "0" || f[2] != f[3] {
		return 0, fmt.Errorf("viewBox %q is not a square from 0 0", viewBox)
	}
	side, err := strconv.ParseFloat(f[2], 64)
	if err != nil || side <= 0 {
		return 0, fmt.Errorf("viewBox %q is not a square from 0 0", viewBox)
	}
	return side, nil
}

// curvesOf reads a path's d: M x y, then L x y and C x1 y1 x2 y2 x y, then Z, as often as it has
// outlines.
func curvesOf(d string) ([][]cubic, error) {
	var out [][]cubic
	var cur []cubic
	var start, at point
	open := false
	f := strings.Fields(strings.NewReplacer(",", " ", "M", " M ", "L", " L ", "C", " C ", "Z", " Z ").Replace(d))
	for i := 0; i < len(f); {
		op := f[i]
		i++
		n := map[string]int{"M": 1, "L": 1, "C": 3, "Z": 0}
		want, ok := n[op]
		if !ok {
			return nil, fmt.Errorf("%q: only absolute M, L, C and Z", op)
		}
		if i+2*want > len(f) {
			return nil, fmt.Errorf("%s: too few numbers", op)
		}
		var p [3]point
		for j := range want {
			x, errX := strconv.ParseFloat(f[i], 64)
			y, errY := strconv.ParseFloat(f[i+1], 64)
			if errX != nil || errY != nil {
				return nil, fmt.Errorf("%s: %q %q is no point", op, f[i], f[i+1])
			}
			p[j] = point{x, y}
			i += 2
		}
		if open != (op != "M") {
			return nil, fmt.Errorf("%s: an outline is M, its segments, Z", op)
		}
		switch op {
		case "M":
			start, at, open = p[0], p[0], true
		case "L":
			cur = append(cur, cubic{at, at, p[0], p[0]})
			at = p[0]
		case "C":
			cur = append(cur, cubic{at, p[0], p[1], p[2]})
			at = p[2]
		case "Z":
			if at != start {
				cur = append(cur, cubic{at, at, start, start})
			}
			out, cur, open = append(out, cur), nil, false
		}
	}
	if open {
		return nil, fmt.Errorf("the last outline has no Z")
	}
	return out, nil
}

// ellipse is four quarter arcs, each the cubic that is closest to one.
func ellipse(cx, cy, rx, ry float64) []cubic {
	const k = 0.5522847498307936 // 4/3 (sqrt 2 - 1)
	e, s, w, n := point{cx + rx, cy}, point{cx, cy + ry}, point{cx - rx, cy}, point{cx, cy - ry}
	return []cubic{
		{e, point{cx + rx, cy + k*ry}, point{cx + k*rx, cy + ry}, s},
		{s, point{cx - k*rx, cy + ry}, point{cx - rx, cy + k*ry}, w},
		{w, point{cx - rx, cy - k*ry}, point{cx - k*rx, cy - ry}, n},
		{n, point{cx + k*rx, cy - ry}, point{cx + rx, cy - k*ry}, e},
	}
}

// form is one way the mark is shown: on a tile or on nothing, in its colours or in one.
type form struct {
	tile string // the square's ground, #RRGGBB; "" is none
	one  string // every shape in this colour; "" is each in its own
}

// margin is the tile's side over the drawing's: a tile has a tenth of itself around the mark, the
// room a platform's rounding and its mask take. A mark on nothing fills its square: 16 px of it is
// 16 px of mark.
const margin = 1.25

// place is where a point of the drawing falls in a square of size.
func (d drawing) place(f form, size float64) (scale, offset float64) {
	scale = size / d.side
	if f.tile != "" {
		scale /= margin
	}
	return scale, (size - float64(d.side*scale)) / 2
}

// steps is the lines a cubic is drawn as. The mark's largest arc, a quarter of a disk at 512 px,
// is within a fiftieth of a pixel of its chords at this count.
const steps = 64

// raster draws the form at size pixels. Curves are flattened here and not by the rasterizer, and
// every product is rounded before it is added: left to the compiler, arm64 fuses a multiply and
// an add into one rounding, and the same drawing would make other bytes there than on amd64.
// At these sizes the rasterizer itself works in fixed point.
func (d drawing) raster(f form, size int) (*image.RGBA, error) {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	if f.tile != "" {
		c, err := rgbOf(f.tile)
		if err != nil {
			return nil, err
		}
		draw.Draw(img, img.Bounds(), &image.Uniform{c}, image.Point{}, draw.Src)
	}
	scale, offset := d.place(f, float64(size))
	to := func(p point) (float32, float32) {
		return float32(float64(p.x*scale) + offset), float32(float64(p.y*scale) + offset)
	}
	for _, s := range d.shapes {
		fill := s.fill
		if f.one != "" {
			fill = f.one
		}
		c, err := rgbOf(fill)
		if err != nil {
			return nil, err
		}
		z := vector.NewRasterizer(size, size)
		for _, curve := range s.curves {
			z.MoveTo(to(curve[0].p0))
			for _, b := range curve {
				for i := 1; i <= steps; i++ {
					z.LineTo(to(b.at(float64(i) / steps)))
				}
			}
			z.ClosePath()
		}
		z.Draw(img, img.Bounds(), &image.Uniform{c}, image.Point{})
	}
	return img, nil
}

func (b cubic) at(t float64) point {
	u := 1 - t
	w0, w1, w2, w3 := float64(u*u)*u, 3*float64(float64(u*u)*t), 3*float64(float64(u*t)*t), float64(t*t)*t
	return point{
		float64(w0*b.p0.x) + float64(w1*b.p1.x) + float64(w2*b.p2.x) + float64(w3*b.p3.x),
		float64(w0*b.p0.y) + float64(w1*b.p1.y) + float64(w2*b.p2.y) + float64(w3*b.p3.y),
	}
}

// svg writes the form as a file: the drawing's own coordinates, moved onto a tile by a transform.
func (d drawing) svg(f form) []byte {
	var b strings.Builder
	side := num(d.side)
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %s %s">`+"\n", side, side)
	indent := "\t"
	if f.tile != "" {
		scale, offset := d.place(f, d.side)
		fmt.Fprintf(&b, "\t"+`<rect width="%s" height="%s" fill="%s"/>`+"\n", side, side, f.tile)
		fmt.Fprintf(&b, "\t"+`<g transform="translate(%s %s) scale(%s)">`+"\n", num(offset), num(offset), num(scale))
		indent = "\t\t"
	}
	for _, s := range d.shapes {
		fill := s.fill
		if f.one != "" {
			fill = f.one
		}
		fmt.Fprintf(&b, `%s<path fill="%s" d="`, indent, fill)
		for i, curve := range s.curves {
			if i > 0 {
				b.WriteByte(' ')
			}
			fmt.Fprintf(&b, "M%s %s", num(curve[0].p0.x), num(curve[0].p0.y))
			for _, c := range curve {
				fmt.Fprintf(&b, " C%s %s %s %s %s %s", num(c.p1.x), num(c.p1.y), num(c.p2.x), num(c.p2.y), num(c.p3.x), num(c.p3.y))
			}
			b.WriteString(" Z")
		}
		b.WriteString(`"/>` + "\n")
	}
	if f.tile != "" {
		b.WriteString("\t</g>\n")
	}
	b.WriteString("</svg>\n")
	return []byte(b.String())
}

// num is a coordinate to a thousandth of a unit, which at 512 px is a hundredth of a pixel.
func num(v float64) string {
	return strconv.FormatFloat(math.Round(v*1000)/1000+0, 'f', -1, 64) // +0: no "-0"
}

func rgbOf(hex string) (color.RGBA, error) {
	if len(hex) != 7 || hex[0] != '#' {
		return color.RGBA{}, fmt.Errorf("%q is no #rrggbb", hex)
	}
	v, err := strconv.ParseUint(hex[1:], 16, 32)
	if err != nil {
		return color.RGBA{}, fmt.Errorf("%q is no #rrggbb", hex)
	}
	return color.RGBA{uint8(v >> 16), uint8(v >> 8), uint8(v), 255}, nil
}
