// Command mark writes logos/ from assets/mark.svg, the mark as it is drawn: a stack of disks in
// the brand's solid colours. Every file is that one drawing: in its colours on nothing (clear), in
// one colour on nothing (ink for a light ground, frost for a dark one), and in its colours on the
// brand's navy, square, for a platform that asks for a tile and rounds it itself (a favicon, a
// touch icon, a link preview). Each form is written as SVG, and as PNG at the sizes that need one.
//
// With -check it writes nothing and fails when logos/ is not what the drawing makes.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"slices"

	"github.com/BurntSushi/toml"
)

// brand is the palette's [brand]: the colours the mark may be drawn in, its tile and its one-colour
// forms.
type brand struct{ Navy, Deep, Blue, Cyan, Frost string }

// logo is one file logos/ holds: a form of the mark, as SVG (size 0) or as PNG at a size
// something shows it at.
type logo struct {
	name string
	form form
	size int
}

func logos(b brand) []logo {
	tile, clear := form{tile: b.Navy}, form{}
	ink, frost := form{one: b.Navy}, form{one: b.Frost}
	return []logo{
		{"mark.svg", tile, 0},
		{"mark-512.png", tile, 512}, // the link preview
		{"mark-256.png", tile, 256},
		{"apple-touch-icon.png", tile, 180}, // a phone's home screen
		{"favicon-64.png", tile, 64},        // a browser's tab, whose ground is the browser's
		{"favicon-32.png", tile, 32},

		{"mark-clear.svg", clear, 0}, // the mark in the product, on either theme
		{"mark-clear-512.png", clear, 512},
		{"mark-clear-256.png", clear, 256},
		// The same drawing as mark-clear-*: the illustration needed its shadow redrawn for a dark
		// page, and the places that show the mark there read these names.
		{"mark-dark-512.png", clear, 512},
		{"mark-dark-256.png", clear, 256},

		{"mark-ink.svg", ink, 0}, // where colour cannot go, on a light ground
		{"mark-ink-512.png", ink, 512},
		{"mark-ink-256.png", ink, 256},
		{"mark-frost.svg", frost, 0}, // and on a dark one
		{"mark-frost-512.png", frost, 512},
		{"mark-frost-256.png", frost, 256},
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "mark:", err)
		os.Exit(1)
	}
}

func run() error {
	src := flag.String("mark", "assets/mark.svg", "the mark as drawn")
	palette := flag.String("palette", "palette.toml", "the palette, whose [brand] is the mark's colours")
	out := flag.String("out", "logos", "where the mark is written")
	check := flag.Bool("check", false, "write nothing; fail if logos/ is not what the drawing makes")
	flag.Parse()

	var p struct{ Brand brand }
	if _, err := toml.DecodeFile(*palette, &p); err != nil {
		return err
	}
	f, err := os.Open(*src)
	if err != nil {
		return err
	}
	defer f.Close()
	d, err := readDrawing(f)
	if err != nil {
		return fmt.Errorf("%s: %w", *src, err)
	}
	if err := inBrand(d, p.Brand); err != nil {
		return fmt.Errorf("%s: %w", *src, err)
	}

	var stale []string
	for _, l := range logos(p.Brand) {
		want, err := encode(d, l)
		if err != nil {
			return fmt.Errorf("%s: %w", l.name, err)
		}
		path := filepath.Join(*out, l.name)
		if have, err := os.ReadFile(path); err == nil && bytes.Equal(have, want) {
			continue
		}
		stale = append(stale, l.name)
		if *check {
			continue
		}
		if err := os.MkdirAll(*out, 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, want, 0o644); err != nil {
			return err
		}
	}
	if *check && len(stale) > 0 {
		return fmt.Errorf("logos/ is not what %s makes, run `task mark`: %v", *src, stale)
	}
	return nil
}

// inBrand fails on a shape filled with anything but the four colours the mark is drawn in. Navy is
// its ground, never a shape; mint is for drawings, and the mark has no ring.
func inBrand(d drawing, b brand) error {
	own := []string{b.Deep, b.Blue, b.Cyan, b.Frost}
	for _, s := range d.shapes {
		if !slices.Contains(own, s.fill) {
			return fmt.Errorf("fill %q is not [brand]'s deep, blue, cyan or frost %v", s.fill, own)
		}
	}
	return nil
}

func encode(d drawing, l logo) ([]byte, error) {
	if l.size == 0 {
		return d.svg(l.form), nil
	}
	img, err := d.raster(l.form, l.size)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	err = (&png.Encoder{CompressionLevel: png.BestCompression}).Encode(&b, img)
	return b.Bytes(), err
}
