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
	"math"
	"os"
	"path/filepath"

	"golang.org/x/image/draw"
)

// logo is one file logos/ holds: the mark at a size something shows it at, on its own ground (a
// tile) or with the ground taken out (clear).
type logo struct {
	size  int
	clear bool
}

var sizes = map[string]logo{
	"mark-512.png":       {512, false}, // the link preview, a dark hero
	"mark-256.png":       {256, false}, // a dark header
	"mark-clear-512.png": {512, true},  // on any ground: a light page, a light hero
	"mark-clear-256.png": {256, true},  // a light header
	"favicon-64.png":     {64, false},  // a browser's tab, whose ground is the browser's
	"favicon-32.png":     {32, false},
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
	clear := unground(tile, avatar.At(avatar.Bounds().Min.X+4, avatar.Bounds().Min.Y+4))
	var stale []string
	for name, l := range sizes {
		src := tile
		if l.clear {
			src = clear
		}
		dst := image.NewRGBA(image.Rect(0, 0, l.size, l.size))
		draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Src, nil)
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

// unground takes the avatar's ground out of the tile: the ground is what a fill from the
// tile's edges reaches through pixels within a few levels of it, and everything else - the
// disks, the ring, the dark seams between the disks, which are darker than the ground - stays
// opaque in its own color. The ground is textured, so a pixel near its edge fades by how far it
// is from the ground rather than being cut, which keeps the silhouette's antialiasing.
func unground(tile image.Image, ground color.Color) image.Image {
	const near, far = 5.0, 11.0 // distance to the ground, in 8-bit levels: ground under near, opaque past far
	gr, gg, gb, _ := ground.RGBA()
	dist := func(c color.Color) float64 {
		r, g, b, _ := c.RGBA()
		dr, dg, db := float64(r>>8)-float64(gr>>8), float64(g>>8)-float64(gg>>8), float64(b>>8)-float64(gb>>8)
		return math.Sqrt(dr*dr + dg*dg + db*db)
	}
	b := tile.Bounds()
	w, h := b.Dx(), b.Dy()
	isGround := make([]bool, w*h)
	var queue []image.Point
	push := func(x, y int) {
		if x < 0 || y < 0 || x >= w || y >= h || isGround[y*w+x] {
			return
		}
		if dist(tile.At(b.Min.X+x, b.Min.Y+y)) >= far {
			return
		}
		isGround[y*w+x] = true
		queue = append(queue, image.Pt(x, y))
	}
	for x := range w {
		push(x, 0)
		push(x, h-1)
	}
	for y := range h {
		push(0, y)
		push(w-1, y)
	}
	for len(queue) > 0 {
		p := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		push(p.X+1, p.Y)
		push(p.X-1, p.Y)
		push(p.X, p.Y+1)
		push(p.X, p.Y-1)
	}
	out := image.NewNRGBA(b)
	for y := range h {
		for x := range w {
			c := color.NRGBAModel.Convert(tile.At(b.Min.X+x, b.Min.Y+y)).(color.NRGBA)
			if isGround[y*w+x] {
				d := dist(c)
				if d <= near {
					continue
				}
				c.A = uint8(255 * (d - near) / (far - near))
			}
			out.SetNRGBA(b.Min.X+x, b.Min.Y+y, c)
		}
	}
	return out
}
