package main

import (
	"bytes"
	"os"
	"regexp"
	"testing"
)

func TestTheMarkTracesTheAvatar(t *testing.T) {
	tile, err := os.ReadFile("../../assets/mark.svg")
	if err != nil {
		t.Fatal(err)
	}
	middle := regexp.MustCompile(`(?s)<!-- The middle disk.*?</g>`)
	if !middle.Match(tile) {
		t.Fatal("assets/mark.svg has no middle disk to take out")
	}
	for _, c := range []struct {
		name   string
		svg    []byte
		traces bool
	}{
		{"the trace", tile, true},
		{"the trace without its middle disk", middle.ReplaceAll(tile, nil), false},
	} {
		t.Run(c.name, func(t *testing.T) {
			traced, err := square(c.svg, "../../assets/avatar.png")
			if err != nil {
				t.Fatal(err)
			}
			d, err := distance(c.svg, traced)
			if err != nil {
				t.Fatal(err)
			}
			if (d <= drift) != c.traces {
				t.Errorf("distance %.1f, drift %.1f: traces the avatar = %v, want %v", d, drift, d <= drift, c.traces)
			}
		})
	}
}

func TestEverySizeIsDrawn(t *testing.T) {
	tile, err := os.ReadFile("../../assets/mark.svg")
	if err != nil {
		t.Fatal(err)
	}
	clear := ground.ReplaceAll(tile, nil)
	if bytes.Equal(clear, tile) {
		t.Fatal("the clear mark is the tile: no ground was taken out")
	}
	for _, size := range []int{32, 64, 180, 256, 1024} {
		// The middle disk's face is under the centre at every size; the clear mark's corner is
		// the page behind it.
		img, err := render(tile, size)
		if err != nil {
			t.Fatal(err)
		}
		if c := img.RGBAAt(size/2, size*48/100); c.A != 255 || c.G < 100 {
			t.Errorf("%d px: the centre is %v, not the middle disk", size, c)
		}
		img, err = render(clear, size)
		if err != nil {
			t.Fatal(err)
		}
		if c := img.RGBAAt(0, 0); c.A != 0 {
			t.Errorf("%d px clear: the corner is %v, not transparent", size, c)
		}
	}
}
