package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

func readPalette(t *testing.T) palette {
	t.Helper()
	var p palette
	if _, err := toml.DecodeFile("../../palette.toml", &p); err != nil {
		t.Fatal(err)
	}
	return p
}

func written(t *testing.T, p palette) map[string][]byte {
	t.Helper()
	results, err := holds(p)
	if err != nil {
		t.Fatal(err)
	}
	files, err := outputs(p, results)
	if err != nil {
		t.Fatal(err)
	}
	return files
}

// blocks is a stylesheet's rules in order, each its selector and its variables by name.
type block struct {
	selector string
	vars     map[string]string
}

func blocks(css []byte) []block {
	var out []block
	for _, m := range regexp.MustCompile(`(?s)([^{}/]+?) \{(.*?)\}`).FindAllStringSubmatch(string(css), -1) {
		b := block{selector: strings.TrimSpace(m[1]), vars: map[string]string{}}
		for _, v := range regexp.MustCompile(`--([a-z0-9-]+): ([^;]+);`).FindAllStringSubmatch(m[2], -1) {
			b.vars[v[1]] = v[2]
		}
		out = append(out, b)
	}
	return out
}

// A consumer draws a control, a row, a corner and its navigation at the palette's sizes only if
// the file it copies says them: each output carries each size its consumer draws, as
// palette.toml has it.
func TestEveryOutputCarriesTheSizesItsConsumerDraws(t *testing.T) {
	p := readPalette(t)
	files := written(t, p)
	px := func(n int) string { return fmt.Sprintf("%dpx", n) }
	app := map[string]string{
		"spin-control-sm": px(p.Controls.Sm), "spin-control-md": px(p.Controls.Md), "spin-control-lg": px(p.Controls.Lg),
		"spin-row": px(p.Layout.Row), "spin-sidebar-app": px(p.Layout.SidebarApp),
	}
	for file, want := range map[string]map[string]string{
		"tokens.css":    merge(app, map[string]string{"spin-radius-sm": px(p.Layout.RadiusSm), "spin-sidebar": fmt.Sprintf("%drem", p.Layout.Sidebar)}),
		"dashboard.css": merge(app, map[string]string{"radius": px(p.Layout.RadiusSm)}),
		"starlight.css": {"sl-sidebar-width": fmt.Sprintf("%drem", p.Layout.Sidebar)},
	} {
		got := map[string]string{}
		for _, b := range blocks(files[file]) {
			for k, v := range b.vars {
				got[k] = v
			}
		}
		for name, v := range want {
			if got[name] != v {
				t.Errorf("%s: --%s is %q, want %q", file, name, got[name], v)
			}
		}
	}

	var tokens struct {
		Layout   map[string]int
		Controls map[string]int
	}
	if err := json.Unmarshal(files["tokens.json"], &tokens); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]int{"Row": p.Layout.Row, "SidebarApp": p.Layout.SidebarApp, "Sidebar": p.Layout.Sidebar, "RadiusSm": p.Layout.RadiusSm} {
		if tokens.Layout[name] != want {
			t.Errorf("tokens.json: layout.%s is %d, want %d", name, tokens.Layout[name], want)
		}
	}
	for name, want := range map[string]int{"Sm": p.Controls.Sm, "Md": p.Controls.Md, "Lg": p.Controls.Lg} {
		if tokens.Controls[name] != want {
			t.Errorf("tokens.json: controls.%s is %d, want %d", name, tokens.Controls[name], want)
		}
	}
}

func merge(a, b map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

// A button's outline and an input's edge are two variables in the outputs that draw controls, in
// both themes: the input's edge is what holds 3:1 on the page, and the outline is the quieter of
// the two. One variable for both is what drew every secondary button with the heaviest line on
// the screen.
func TestAButtonsOutlineIsQuieterThanAnInputsEdge(t *testing.T) {
	files := written(t, readPalette(t))
	for _, tc := range []struct {
		file, page, outline, input string
		parse                      func(string) (rgb, error)
	}{
		{"tokens.css", "spin-page", "spin-control-border", "spin-border-strong", parseHex},
		{"dashboard.css", "background", "control-border", "input", parseHSL},
	} {
		var themes int
		for _, b := range blocks(files[tc.file]) {
			if b.vars[tc.page] == "" {
				continue
			}
			themes++
			var c [3]rgb
			for i, name := range []string{tc.page, tc.outline, tc.input} {
				v, err := tc.parse(b.vars[name])
				if err != nil {
					t.Fatalf("%s %s: --%s: %v", tc.file, b.selector, name, err)
				}
				c[i] = v
			}
			outline, input := contrast(c[1], c[0]), contrast(c[2], c[0])
			if input < nonText {
				t.Errorf("%s %s: --%s on --%s is %.2f:1, under 3", tc.file, b.selector, tc.input, tc.page, input)
			}
			if outline >= input {
				t.Errorf("%s %s: --%s is %.2f:1 on the page and --%s %.2f:1: the outline is not the quieter", tc.file, b.selector, tc.outline, outline, tc.input, input)
			}
		}
		if themes != 2 {
			t.Fatalf("%s: %d themes carry --%s, want dark and light", tc.file, themes, tc.page)
		}
	}
}

func parseHSL(s string) (rgb, error) {
	var h, sat, l float64
	if _, err := fmt.Sscanf(s, "%f %f%% %f%%", &h, &sat, &l); err != nil {
		return rgb{}, fmt.Errorf("%q is not H S%% L%%: %w", s, err)
	}
	return fromHSL(h, sat/100, l/100), nil
}

// Each pair this palette holds a quiet line or a plane by fails when the palette breaks it, in
// the theme that broke it and by its own name: a check nothing can fail holds nothing.
func TestAQuietLineOrPlaneThatIsNotToldApartFails(t *testing.T) {
	for _, tc := range []struct {
		pair  string
		plant func(*mode)
	}{
		{"hairline on page", func(m *mode) { m.Hairline = m.Page }},
		{"a secondary button's outline on page", func(m *mode) { m.Border = m.Page }},
		{"a secondary button's outline on surface", func(m *mode) { m.Border = m.Surface }},
		{"a row under the pointer | page", func(m *mode) { m.Surface = m.Page }},
		{"a row's action under the pointer | its row", func(m *mode) { m.Raised = m.Surface }},
	} {
		for _, theme := range []string{"dark", "light"} {
			t.Run(theme+" "+tc.pair, func(t *testing.T) {
				p := readPalette(t)
				if got := pairs(t, p, tc.pair); !got["dark"] || !got["light"] {
					t.Fatalf("as written the pair holds in %v, want both themes", got)
				}
				other := "light"
				if theme == "dark" {
					tc.plant(&p.Dark)
				} else {
					tc.plant(&p.Light)
					other = "dark"
				}
				got := pairs(t, p, tc.pair)
				if got[theme] {
					t.Errorf("%s: the pair holds with the two colours the same", theme)
				}
				if !got[other] {
					t.Errorf("%s: the pair fails for a change to %s", other, theme)
				}
			})
		}
	}
}

// pairs is whether the named pair holds, by theme; a theme the checker has no such pair for is
// absent.
func pairs(t *testing.T, p palette, what string) map[string]bool {
	t.Helper()
	results, err := holds(p)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]bool{}
	for _, r := range results {
		if r.what == what {
			out[r.mode] = r.ok
		}
	}
	if len(out) != 2 {
		t.Fatalf("%q is held in %d themes, want 2", what, len(out))
	}
	return out
}

// The hairline is there to be quieter than the border: nearer the page, in both themes, as drawn.
func TestTheHairlineIsNearerThePageThanTheBorder(t *testing.T) {
	results, err := holds(readPalette(t))
	if err != nil {
		t.Fatal(err)
	}
	drawn := map[string]float64{}
	for _, r := range results {
		if r.what == "hairline on page" || r.what == "border on page" {
			drawn[r.mode+" "+r.what] = r.distance["drawn"]
		}
	}
	for _, theme := range []string{"dark", "light"} {
		h, b := drawn[theme+" hairline on page"], drawn[theme+" border on page"]
		if h == 0 || h >= b {
			t.Errorf("%s: hairline is ΔE00 %.1f from the page and border %.1f", theme, h, b)
		}
	}
}
