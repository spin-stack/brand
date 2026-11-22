package main

import (
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

// A solid fill is told from itself under the pointer and keeps its label at 4.5:1 there and
// pressed, whichever end of lightness it starts from: the mark's navy under a white label, where
// black over the fill is a step of ΔE 1, and the mark's cyan under a navy label. Each is set as
// light's accent in palette.toml and held by the checker's own pairs.
func TestASolidFillStepsAwayFromItsOwnLightness(t *testing.T) {
	for _, tc := range []struct {
		name, fill, label string
		lightens          bool
	}{
		{"ink", "navy", "#FFFFFF", true},
		{"cyan", "cyan", "#111827", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var p palette
			if _, err := toml.DecodeFile("../../palette.toml", &p); err != nil {
				t.Fatal(err)
			}
			p.Light.Accent, p.Light.OnAccent = p.Brand[tc.fill], tc.label
			results, err := holds(p)
			if err != nil {
				t.Fatal(err)
			}
			var held int
			for _, r := range results {
				if r.mode != "light" || !strings.Contains(r.what, "accent under the pointer") && r.what != "on_accent on accent pressed" {
					continue
				}
				held++
				if !r.ok {
					t.Error(r)
				}
			}
			if held != 3 {
				t.Fatalf("%d pairs hold the accent's states, want the label under the pointer, the label pressed, and the fill against itself", held)
			}

			c, err := colors(p.Light)
			if err != nil {
				t.Fatal(err)
			}
			st := controlStates(p, c, false)
			rest, hover, pressed := c["accent"].linear().luminance(), st.accentHover.linear().luminance(), st.accentActive.linear().luminance()
			if lighter := hover > rest && pressed > hover; lighter != tc.lightens {
				t.Errorf("luminance at rest %.3f, under the pointer %.3f, pressed %.3f: lightens = %v, want %v", rest, hover, pressed, lighter, tc.lightens)
			}
		})
	}
}
