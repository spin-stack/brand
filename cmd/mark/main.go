// Command mark cuts the spin-stack mark - the stack of disks - out of the organization's avatar
// (assets/avatar.png), without the wordmark under it, and writes it at the sizes the places that
// show it use, into logos/: on its own ground, square, for each to round as it rounds a tile.
// With -check it writes nothing and fails when logos/ is not what the avatar makes.
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

	"golang.org/x/image/draw"
)

// sizes are the files logos/ holds: the tile at each size something shows it at.
var sizes = map[string]int{
	"mark-512.png":   512, // the hero, the link preview, the preview page
	"mark-256.png":   256, // a site's header
	"favicon-64.png": 64,
	"favicon-32.png": 32,
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "mark:", err)
		os.Exit(1)
	}
}

func run() error {
	src := flag.String("avatar", "assets/avatar.png", "the organization's avatar")
	out := flag.String("out", "logos", "where the mark is written")
	check := flag.Bool("check", false, "write nothing; fail if logos/ is not what the avatar makes")
	flag.Parse()

	f, err := os.Open(*src)
	if err != nil {
		return err
	}
	defer f.Close()
	avatar, err := png.Decode(f)
	if err != nil {
		return err
	}
	tile, err := cut(avatar)
	if err != nil {
		return err
	}
	var stale []string
	for name, size := range sizes {
		dst := image.NewRGBA(image.Rect(0, 0, size, size))
		draw.CatmullRom.Scale(dst, dst.Bounds(), tile, tile.Bounds(), draw.Src, nil)
		var b bytes.Buffer
		if err := (&png.Encoder{CompressionLevel: png.BestCompression}).Encode(&b, dst); err != nil {
			return err
		}
		path := filepath.Join(*out, name)
		have, err := os.ReadFile(path)
		if err == nil && bytes.Equal(have, b.Bytes()) {
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

// cut finds the stack in the avatar - the first band of rows with anything but the ground in
// them, which ends at the gap above the wordmark - and centers it on a square of the ground, with
// a fifth of its size as margin.
func cut(avatar image.Image) (image.Image, error) {
	b := avatar.Bounds()
	ground := avatar.At(b.Min.X+4, b.Min.Y+4)
	away := func(x, y int) bool {
		r1, g1, b1, _ := avatar.At(x, y).RGBA()
		r2, g2, b2, _ := ground.RGBA()
		d := func(a, b uint32) int {
			if a > b {
				return int(a-b) >> 8
			}
			return int(b-a) >> 8
		}
		return d(r1, r2)+d(g1, g2)+d(b1, b2) > 40
	}
	rowHas := func(y int) bool {
		for x := b.Min.X; x < b.Max.X; x++ {
			if away(x, y) {
				return true
			}
		}
		return false
	}
	top := -1
	for y := b.Min.Y; y < b.Max.Y; y++ {
		if rowHas(y) {
			top = y
			break
		}
	}
	if top < 0 {
		return nil, fmt.Errorf("the avatar is only its ground")
	}
	bottom, gap := top, 0
	for y := top; y < b.Max.Y && gap <= 6; y++ {
		if rowHas(y) {
			bottom, gap = y, 0
		} else {
			gap++
		}
	}
	left, right := b.Max.X, b.Min.X
	for y := top; y <= bottom; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if away(x, y) {
				left, right = min(left, x), max(right, x)
			}
		}
	}
	stack := image.Rect(left-4, top-4, right+5, bottom+5)
	side := int(float64(max(stack.Dx(), stack.Dy())) * 1.22)
	tile := image.NewRGBA(image.Rect(0, 0, side, side))
	draw.Draw(tile, tile.Bounds(), &image.Uniform{color.RGBAModel.Convert(ground)}, image.Point{}, draw.Src)
	at := image.Pt((side-stack.Dx())/2, (side-stack.Dy())/2)
	draw.Draw(tile, stack.Sub(stack.Min).Add(at), avatar, stack.Min, draw.Src)
	return tile, nil
}
