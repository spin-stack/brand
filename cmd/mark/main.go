// Command mark cuts the spin-stack mark - the stack of disks - out of assets/mark.png, the mark as
// it was drawn: 1024 px on a clear ground, the stack some 450 of them. (The organization's avatar
// is the same mark shrunk to 460 px on the navy with the wordmark under it, the stack 240 px: cut
// from it, every logo larger than that, or on a 2x screen, was an enlargement.)
// It writes the mark at the sizes the places that show it use, into logos/: on the brand's navy,
// square, for each to round as it rounds a tile, and clear, for a light ground. Every size is the
// drawing reduced, never enlarged: a size the drawing is too small for fails.
//
// With -check it writes nothing and fails when logos/ is not what the drawing makes.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strconv"

	"github.com/BurntSushi/toml"
	"golang.org/x/image/draw"
)

// logo is one file logos/ holds: the mark at a size something shows it at, on the navy (a tile)
// or on nothing (clear).
type logo struct {
	size  int
	clear bool
}

var sizes = map[string]logo{
	"mark-512.png":         {512, false}, // the link preview, a dark hero
	"mark-256.png":         {256, false}, // a dark header
	"mark-clear-512.png":   {512, true},  // on any ground: a light page, a light hero
	"mark-clear-256.png":   {256, true},  // a light header
	"apple-touch-icon.png": {180, false}, // a phone's home screen
	"favicon-64.png":       {64, false},  // a browser's tab, whose ground is the browser's
	"favicon-32.png":       {32, false},
}

// margin is the square's side over the stack's longer one: a fifth of it around the stack.
const margin = 1.22

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "mark:", err)
		os.Exit(1)
	}
}

func run() error {
	src := flag.String("mark", "assets/mark.png", "the mark as drawn, on a clear ground")
	palette := flag.String("palette", "palette.toml", "the palette, whose brand navy is the tile's ground")
	out := flag.String("out", "logos", "where the mark is written")
	check := flag.Bool("check", false, "write nothing; fail if logos/ is not what the drawing makes")
	flag.Parse()

	navy, err := ground(*palette)
	if err != nil {
		return err
	}
	f, err := os.Open(*src)
	if err != nil {
		return err
	}
	defer f.Close()
	drawing, err := png.Decode(f)
	if err != nil {
		return err
	}
	stack, err := stackOf(drawing)
	if err != nil {
		return fmt.Errorf("%s: %w", *src, err)
	}

	onNavy := shadowOn(drawing, stack, navy)

	var stale []string
	for name, l := range sizes {
		src := onNavy
		if l.clear {
			src = drawing
		}
		img, err := cut(src, stack, l.size, l.clear, navy)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		var b bytes.Buffer
		if err := (&png.Encoder{CompressionLevel: png.BestCompression}).Encode(&b, img); err != nil {
			return err
		}
		path := filepath.Join(*out, name)
		if have, err := os.ReadFile(path); err == nil && bytes.Equal(have, b.Bytes()) {
			continue
		}
		stale = append(stale, name)
		if *check {
			continue
		}
		if err := os.MkdirAll(*out, 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, b.Bytes(), 0o644); err != nil {
			return err
		}
	}
	if *check && len(stale) > 0 {
		return fmt.Errorf("logos/ is not what %s makes, run `task mark`: %v", *src, stale)
	}
	return nil
}

// ground is the palette's brand navy, the tile's ground and the dark page's.
func ground(palette string) (color.RGBA, error) {
	var p struct{ Brand map[string]string }
	if _, err := toml.DecodeFile(palette, &p); err != nil {
		return color.RGBA{}, err
	}
	hex := p.Brand["navy"]
	if len(hex) != 7 || hex[0] != '#' {
		return color.RGBA{}, fmt.Errorf("%s: brand.navy %q is no #rrggbb", palette, hex)
	}
	v, err := strconv.ParseUint(hex[1:], 16, 32)
	if err != nil {
		return color.RGBA{}, fmt.Errorf("%s: brand.navy %q is no #rrggbb", palette, hex)
	}
	return color.RGBA{uint8(v >> 16), uint8(v >> 8), uint8(v), 255}, nil
}

// stackOf is the box of what the drawing draws: every pixel past a faint alpha, which leaves out
// the drawing's near-invisible haze.
func stackOf(drawing image.Image) (image.Rectangle, error) {
	b := drawing.Bounds()
	box := image.Rectangle{}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if _, _, _, a := drawing.At(x, y).RGBA(); a>>8 > 16 {
				box = box.Union(image.Rect(x, y, x+1, y+1))
			}
		}
	}
	if box.Empty() {
		return box, fmt.Errorf("it draws nothing: no pixel is more than faintly opaque")
	}
	return box, nil
}

// shadowOn is the drawing for the tile: its shadow is a light grey ring, drawn for a light ground,
// which on the navy reads as a halo; a shadow is never lighter than what it falls on. The ring is
// the grey in the stack's bottom quarter - the white top disk is grey too, but at the top - and
// becomes the navy with a sixteenth of the ring's light in it, a rim as the avatar's was.
func shadowOn(drawing image.Image, stack image.Rectangle, navy color.RGBA) image.Image {
	const light = 0.06
	out := image.NewNRGBA(drawing.Bounds())
	draw.Draw(out, out.Bounds(), drawing, drawing.Bounds().Min, draw.Src)
	band := stack.Min.Y + stack.Dy()*3/4
	for y := band; y < stack.Max.Y; y++ {
		for x := stack.Min.X; x < stack.Max.X; x++ {
			c := out.NRGBAAt(x, y)
			hi, lo := max(c.R, c.G, c.B), min(c.R, c.G, c.B)
			if c.A == 0 || hi-lo > 24 || hi < 96 {
				continue
			}
			mix := func(n, g uint8) uint8 { return uint8(float64(n) + light*(float64(g)-float64(n)) + 0.5) }
			out.SetNRGBA(x, y, color.NRGBA{mix(navy.R, c.R), mix(navy.G, c.G), mix(navy.B, c.B), c.A})
		}
	}
	return out
}

// cut is the stack centered on a square of size pixels, a fifth of its size around it, on the
// navy or on nothing.
func cut(drawing image.Image, stack image.Rectangle, size int, clear bool, navy color.RGBA) (*image.RGBA, error) {
	side := float64(max(stack.Dx(), stack.Dy())) * margin
	scale := float64(size) / side
	if scale > 1 {
		return nil, fmt.Errorf("%d px would enlarge the drawing %.2f times: its stack is %d px", size, scale, max(stack.Dx(), stack.Dy()))
	}
	w, h := int(float64(stack.Dx())*scale+0.5), int(float64(stack.Dy())*scale+0.5)
	at := image.Pt((size-w)/2, (size-h)/2)
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	if !clear {
		draw.Draw(img, img.Bounds(), &image.Uniform{navy}, image.Point{}, draw.Src)
	}
	draw.CatmullRom.Scale(img, image.Rectangle{at, at.Add(image.Pt(w, h))}, drawing, stack, draw.Over, nil)
	return img, nil
}
