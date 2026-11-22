package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func mark(t *testing.T) drawing {
	t.Helper()
	f, err := os.Open("../../assets/mark.svg")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d, err := readDrawing(f)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// The mark at 16 px, where the illustration was a smudge: down its middle, three disks with the
// ground between them.
func TestTheMarkIsThreeDisksAt16Pixels(t *testing.T) {
	img, err := mark(t).raster(form{one: "#111827"}, 16)
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range []int{7, 8} {
		var column []string
		disks, in := 0, false
		for y := range 16 {
			a := img.RGBAAt(x, y).A
			switch {
			case a >= 0xC0:
				column = append(column, "#")
				if !in {
					disks++
				}
				in = true
			case a <= 0x20:
				column = append(column, ".")
				in = false
			default:
				column = append(column, "+")
			}
		}
		if disks != 3 {
			t.Errorf("column %d is %s: %d disks, not 3 with the ground between them", x, strings.Join(column, ""), disks)
		}
	}
}

func TestAWrittenSVGIsTheDrawing(t *testing.T) {
	d := mark(t)
	back, err := readDrawing(bytes.NewReader(d.svg(form{})))
	if err != nil {
		t.Fatal(err)
	}
	want, _ := d.raster(form{}, 64)
	got, err := back.raster(form{}, 64)
	if err != nil {
		t.Fatal(err)
	}
	for i := range want.Pix {
		// The file's coordinates are rounded to a thousandth of a unit.
		if diff := int(want.Pix[i]) - int(got.Pix[i]); diff < -2 || diff > 2 {
			t.Fatalf("byte %d of the 64 px raster is %d read back from the SVG, %d from the source", i, got.Pix[i], want.Pix[i])
		}
	}
}

func TestATileIsTheMarkOnItsGround(t *testing.T) {
	d := mark(t)
	img, err := d.raster(form{tile: "#111827"}, 32)
	if err != nil {
		t.Fatal(err)
	}
	if c := img.RGBAAt(0, 0); c.R != 0x11 || c.G != 0x18 || c.B != 0x27 || c.A != 0xFF {
		t.Errorf("the tile's corner is %v, not the ground", c)
	}
	if c := img.RGBAAt(16, 9); c.R != 0xEA || c.G != 0xF0 || c.B != 0xF6 {
		t.Errorf("the tile's top disk is %v at 16,9, not frost", c)
	}
	if svg := string(d.svg(form{tile: "#111827"})); !strings.Contains(svg, `<rect width="64" height="64" fill="#111827"/>`) {
		t.Errorf("the tile's SVG has no ground:\n%s", svg)
	}
	clear, _ := d.raster(form{}, 32)
	if c := clear.RGBAAt(0, 0); c.A != 0 {
		t.Errorf("the clear mark's corner is %v, not nothing", c)
	}
}

func TestReadDrawingRefusesWhatTheLogosWouldNotCarry(t *testing.T) {
	svg := func(body string) string {
		return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">` + body + `</svg>`
	}
	cases := []struct{ name, src, refused string }{
		{"a stroke", svg(`<ellipse cx="1" cy="1" rx="1" ry="1" fill="#000000" stroke="#FFFFFF"/>`), "stroke"},
		{"a group", svg(`<g><ellipse cx="1" cy="1" rx="1" ry="1" fill="#000000"/></g>`), "<g>"},
		{"a gradient", svg(`<linearGradient/>`), "<linearGradient>"},
		{"a relative segment", svg(`<path fill="#000000" d="M0 0 l1 1 Z"/>`), `"l1"`},
		{"an arc", svg(`<path fill="#000000" d="M0 0 A1 1 0 0 0 2 2 Z"/>`), `"A1"`},
		{"an open outline", svg(`<path fill="#000000" d="M0 0 L1 1"/>`), "no Z"},
		{"a box that is no square", `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 48"/>`, "not a square"},
		{"nothing drawn", svg(``), "draws nothing"},
	}
	for _, c := range cases {
		_, err := readDrawing(strings.NewReader(c.src))
		if err == nil || !strings.Contains(err.Error(), c.refused) {
			t.Errorf("%s: %v, not refused for %s", c.name, err, c.refused)
		}
	}
}

func TestAMarkOutsideTheBrandIsRefused(t *testing.T) {
	b := brand{Navy: "#111827", Deep: "#0C3C84", Blue: "#2F6BE3", Cyan: "#27BFE8", Frost: "#EAF0F6"}
	if err := inBrand(mark(t), b); err != nil {
		t.Errorf("the mark: %v", err)
	}
	mint := drawing{side: 64, shapes: []shape{{fill: "#5FE3D6"}}}
	if err := inBrand(mint, b); err == nil {
		t.Error("a shape in mint was not refused")
	}
}
