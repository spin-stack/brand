// Command mark draws the spin-stack mark - the stack of disks - from assets/mark.svg, the
// organization's avatar traced as a vector, into logos/: the SVG itself, on its own ground (the
// tile) and with the ground taken out (clear), and PNGs at the sizes the places that cannot take
// an SVG show it at, each drawn at that size rather than scaled from a smaller one. The avatar is
// 460 px, its stack 240 of them: anything cut from it and shown larger, or on a 2x screen, was
// an enlargement.
//
// With -check it writes nothing and fails when logos/ is not what the SVG makes, or when the
// trace has moved further from the avatar than it is now.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	"golang.org/x/image/draw"
)

// logo is one file logos/ holds: the mark at a size, on its ground (a tile) or clear.
type logo struct {
	size  int
	clear bool
}

var sizes = map[string]logo{
	"mark-1024.png":        {1024, false}, // the organization's avatar, a print
	"mark-512.png":         {512, false},  // the link preview, a dark hero
	"mark-256.png":         {256, false},  // a dark header
	"mark-clear-1024.png":  {1024, true},
	"mark-clear-512.png":   {512, true}, // on any ground: a light page, a light hero
	"mark-clear-256.png":   {256, true}, // a light header
	"apple-touch-icon.png": {180, false},
	"favicon-64.png":       {64, false}, // a browser's tab, whose ground is the browser's
	"favicon-32.png":       {32, false},
}

// drift is how far the trace may be from the avatar: the mean difference per channel, in 8-bit
// levels, over the square the SVG draws, the wordmark under the stack left out. The trace is
// flatter than the avatar - no texture in the ground, no glow - and measures 10.3; without its
// middle disk it is 17.1. A change that takes it past this is a different mark.
const drift = 12.0

// ground is the tile's group: the navy and the shadow under the stack.
var ground = regexp.MustCompile(`(?s)\s*<g id="ground">.*?</g>`)

// origin is where the SVG's square is in the avatar: its outer group's translate, negated.
var origin = regexp.MustCompile(`<g transform="translate\(-(\d+) -(\d+)\)">`)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "mark:", err)
		os.Exit(1)
	}
}

func run() error {
	src := flag.String("svg", "assets/mark.svg", "the mark")
	avatar := flag.String("avatar", "assets/avatar.png", "the organization's avatar, which the mark traces")
	out := flag.String("out", "logos", "where the mark is written")
	check := flag.Bool("check", false, "write nothing; fail if logos/ is not what the mark makes or the trace left the avatar")
	flag.Parse()

	tile, err := os.ReadFile(*src)
	if err != nil {
		return err
	}
	if !ground.Match(tile) {
		return fmt.Errorf("%s has no <g id=\"ground\">: a clear mark is the tile without it", *src)
	}
	clear := ground.ReplaceAll(tile, nil)
	traced, err := square(tile, *avatar)
	if err != nil {
		return err
	}

	if *check {
		d, err := distance(tile, traced)
		if err != nil {
			return err
		}
		if d > drift {
			return fmt.Errorf("%s is %.1f from %s, past %.1f: it no longer traces the avatar", *src, d, *avatar, drift)
		}
	}

	side, err := sideOf(tile)
	if err != nil {
		return err
	}
	pair, err := trace(tile, traced, side)
	if err != nil {
		return err
	}
	files := map[string][]byte{"mark.svg": tile, "mark-clear.svg": clear, "trace.png": pair}
	for name, l := range sizes {
		svg := tile
		if l.clear {
			svg = clear
		}
		img, err := render(svg, l.size)
		if err != nil {
			return err
		}
		if files[name], err = encode(img); err != nil {
			return err
		}
	}

	var stale []string
	for name, b := range files {
		path := filepath.Join(*out, name)
		if have, err := os.ReadFile(path); err == nil && bytes.Equal(have, b) {
			continue
		}
		stale = append(stale, name)
		if *check {
			continue
		}
		if err := os.MkdirAll(*out, 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, b, 0o644); err != nil {
			return err
		}
	}
	if *check && len(stale) > 0 {
		return fmt.Errorf("logos/ is not what %s makes, run `task mark`: %v", *src, stale)
	}
	return nil
}

// render draws an SVG on a square of size pixels. Its viewBox starts at 0 0: the rasterizer
// shifts a drawing whose viewBox does not by an amount that changes with the size - the favicon
// came out empty at 32 - so the offset into the avatar is a translate instead.
func render(svg []byte, size int) (*image.RGBA, error) {
	icon, err := oksvg.ReadIconStream(bytes.NewReader(svg), oksvg.StrictErrorMode)
	if err != nil {
		return nil, err
	}
	if icon.ViewBox.X != 0 || icon.ViewBox.Y != 0 {
		return nil, fmt.Errorf("the mark's viewBox starts at %g %g, not 0 0: put the offset in a translate", icon.ViewBox.X, icon.ViewBox.Y)
	}
	icon.SetTarget(0, 0, float64(size), float64(size))
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	icon.Draw(rasterx.NewDasher(size, size, rasterx.NewScannerGV(size, size, img, img.Bounds())), 1)
	return img, nil
}

// square is the part of the avatar the mark traces: from its translate, negated, as wide as its
// viewBox. The SVG is in the avatar's coordinates, so drawn at one pixel a unit the two are the
// same pixels.
func square(tile []byte, avatarPath string) (image.Image, error) {
	f, err := os.Open(avatarPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	avatar, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	m := origin.FindSubmatch(tile)
	if m == nil {
		return nil, fmt.Errorf("the mark has no translate to the avatar's coordinates")
	}
	x0, _ := strconv.Atoi(string(m[1]))
	y0, _ := strconv.Atoi(string(m[2]))
	side, err := sideOf(tile)
	if err != nil {
		return nil, err
	}
	// The wordmark under the stack is not the mark: from the first rows of ground after the stack
	// down, the square is the ground, as the tile is.
	b := avatar.Bounds()
	ground := avatar.At(b.Min.X+4, b.Min.Y+4)
	gr, gg, gb, _ := ground.RGBA()
	plain := func(y int) bool {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := avatar.At(x, y).RGBA()
			if abs(r, gr)+abs(g, gg)+abs(bl, gb) > 40 {
				return false
			}
		}
		return true
	}
	end, gap, seen := b.Max.Y, 0, false
	for y := b.Min.Y; y < b.Max.Y; y++ {
		switch {
		case !plain(y):
			seen, gap = true, 0
		case seen:
			gap++
		}
		if gap > 6 {
			end = y
			break
		}
	}
	out := image.NewRGBA(image.Rect(x0, y0, x0+side, y0+side))
	draw.Draw(out, out.Bounds(), avatar, out.Bounds().Min, draw.Src)
	draw.Draw(out, image.Rect(x0, end, x0+side, y0+side), image.NewUniform(ground), image.Point{}, draw.Src)
	return out, nil
}

// sideOf is the side of the mark's square viewBox.
func sideOf(tile []byte) (int, error) {
	icon, err := oksvg.ReadIconStream(bytes.NewReader(tile), oksvg.StrictErrorMode)
	if err != nil {
		return 0, err
	}
	return int(icon.ViewBox.W), nil
}

// distance is the mean difference per channel, in 8-bit levels, between the tile drawn at one
// pixel a unit and the square of the avatar it traces.
func distance(tile []byte, traced image.Image) (float64, error) {
	b := traced.Bounds()
	drawn, err := render(tile, b.Dx())
	if err != nil {
		return 0, err
	}
	var sum float64
	for y := range b.Dy() {
		for x := range b.Dx() {
			r1, g1, b1, _ := drawn.At(x, y).RGBA()
			r2, g2, b2, _ := traced.At(b.Min.X+x, b.Min.Y+y).RGBA()
			sum += abs(r1, r2) + abs(g1, g2) + abs(b1, b2)
		}
	}
	return sum / float64(b.Dx()*b.Dy()*3), nil
}

// trace is the avatar's square and the mark beside it, each twice its size: the one enlarged, as
// anything cut from the avatar was, the other drawn at that size. What a change to the trace looks
// like next to what it traces, for whoever reviews it.
func trace(tile []byte, traced image.Image, side int) ([]byte, error) {
	const gap = 16
	s := 2 * side
	out := image.NewRGBA(image.Rect(0, 0, 2*s+gap, s))
	draw.Draw(out, out.Bounds(), image.White, image.Point{}, draw.Src)
	draw.CatmullRom.Scale(out, image.Rect(0, 0, s, s), traced, traced.Bounds(), draw.Src, nil)
	drawn, err := render(tile, s)
	if err != nil {
		return nil, err
	}
	draw.Draw(out, image.Rect(s+gap, 0, 2*s+gap, s), drawn, image.Point{}, draw.Src)
	return encode(out)
}

func encode(img image.Image) ([]byte, error) {
	var b bytes.Buffer
	err := (&png.Encoder{CompressionLevel: png.BestCompression}).Encode(&b, img)
	return b.Bytes(), err
}

func abs(a, b uint32) float64 {
	a, b = a>>8, b>>8
	if a > b {
		return float64(a - b)
	}
	return float64(b - a)
}
