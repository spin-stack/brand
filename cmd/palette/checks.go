package main

import (
	"fmt"
	"math"
	"strings"
)

// A result is one pair held to its floor: WCAG contrast where one of the two is read, CIEDE2000
// in every view where the two must be told apart.
type result struct {
	mode, kind, what string
	contrast         float64            // 0 when the pair is not held by contrast
	distance         map[string]float64 // ΔE00 by view; nil when the pair is not held by distance
	floor            float64
	worst            float64 // the floor in the worst view
	tritan           float64 // a separate floor through tritanopia, 0 for none
	ok               bool
}

func (r result) String() string {
	mark := "ok  "
	if !r.ok {
		mark = "FAIL"
	}
	if r.distance == nil {
		return fmt.Sprintf("%s %-5s %-9s %-44s %5.2f:1 (floor %.1f)", mark, r.mode, r.kind, r.what, r.contrast, r.floor)
	}
	var b strings.Builder
	for _, v := range views {
		fmt.Fprintf(&b, " %s %.1f", v.name, r.distance[v.name])
	}
	return fmt.Sprintf("%s %-5s %-9s %-44s ΔE00 floor %.0f/%.0f:%s", mark, r.mode, r.kind, r.what, r.floor, r.worst, b.String())
}

// Contrast floors, WCAG 2.2: body text at AAA where it is the product's own text, secondary text
// and links at AA, and what a person must see to operate a control - an input's edge, the focus
// ring, an indicator - at 3:1 (1.4.11).
const (
	bodyText = 7.0
	text     = 4.5
	nonText  = 3.0
)

// Distance tiers, after CONSTRUCT's (spin-stack/brand's THEMING.md says why each): as drawn, and
// in the worst view.
var tiers = map[string][2]float64{
	"edge":    {3, 2},   // two surfaces that meet with nothing between them
	"line":    {6, 4},   // a 1px border against the surface outside it
	"state":   {10, 6},  // selected against not selected
	"hover":   {3, 2},   // a step under the pointer
	"meaning": {20, 10}, // hue that carries meaning: status against status and against the accent,
	// a chart's series against the one drawn next to it, and a chart's statuses against each other
}

// meaningTritan is the floor for meaning through tritanopia, where blue and green are one: a
// status shown by hue alone is not allowed anyway (THEMING.md), so this keeps them apart, not
// readable by hue alone.
const meaningTritan = 6

func holds(p palette) ([]result, error) {
	var out []result
	for _, m := range []struct {
		name string
		m    mode
	}{{"dark", p.Dark}, {"light", p.Light}} {
		c, err := colors(m.m)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", m.name, err)
		}
		surfaces := []string{"page", "surface", "raised", "overlay"}
		white := rgb{1, 1, 1}
		hold := func(kind, what string, fg, bg rgb, floor float64) {
			r := result{mode: m.name, kind: kind, what: what, contrast: contrast(fg, bg), floor: floor}
			r.ok = r.contrast >= floor
			out = append(out, r)
		}
		apart := func(tier, what string, a, b rgb, tritan float64) {
			f := tiers[tier]
			r := result{mode: m.name, kind: tier, what: what, distance: map[string]float64{}, floor: f[0], worst: f[1], tritan: tritan, ok: true}
			for _, v := range views {
				d := deltaE00(v.fn(a.linear()).lab(), v.fn(b.linear()).lab())
				r.distance[v.name] = d
				floor := f[1]
				switch {
				case v.name == "drawn":
					floor = f[0]
				case tritan > 0 && v.name == "tritan":
					floor = tritan
				}
				if d < floor {
					r.ok = false
				}
			}
			out = append(out, r)
		}

		for _, s := range surfaces {
			hold("text", "text on "+s, c["text"], c[s], bodyText)
			hold("text", "muted on "+s, c["muted"], c[s], text)
			hold("text", "accent_text on "+s, c["accent_text"], c[s], text)
			for _, st := range []string{"success", "warning", "error"} {
				hold("text", st+" on "+s, c[st], c[s], text)
			}
		}
		// A state's badge: its colour as text on its colour as a tint, wherever a badge sits.
		for _, s := range []string{"page", "surface", "raised"} {
			for _, st := range []string{"success", "warning", "error", "accent_text"} {
				hold("text", st+" on its badge over "+s, c[st], c[st].over(c[s], p.Tint.Badge), text)
			}
		}
		for _, s := range []string{"page", "surface"} {
			hold("text", "label on "+s, c["label"], c[s], text)
			sel := c["accent"].over(c[s], p.Tint.Selected)
			hold("text", "accent_text on the selected tint over "+s, c["accent_text"], sel, text)
			hold("text", "text on the selected tint over "+s, c["text"], sel, text)
		}
		hold("text", "on_accent on accent", c["on_accent"], c["accent"], text)
		hold("text", "white on destructive", white, c["destructive"], text)
		st := controlStates(p, c, m.name == "dark")
		hold("text", "on_accent on accent under the pointer", c["on_accent"], st.accentHover, text)
		hold("text", "on_accent on accent pressed", c["on_accent"], st.accentActive, text)
		hold("text", "white on destructive under the pointer", white, st.destructiveHover, text)
		hold("text", "white on destructive pressed", white, st.destructiveActive, text)
		hold("text", "text on a secondary button", c["text"], st.control, bodyText)
		hold("text", "text on a secondary button under the pointer", c["text"], st.controlHover, bodyText)
		for _, s := range []string{"page", "raised"} {
			hold("non-text", "border_strong on "+s, c["border_strong"], c[s], nonText)
			hold("non-text", "accent (the focus ring) on "+s, c["accent"], c[s], nonText)
		}

		apart("edge", "page | surface", c["page"], c["surface"], 0)
		apart("edge", "surface | raised", c["surface"], c["raised"], 0)
		apart("edge", "page | raised", c["page"], c["raised"], 0)
		apart("line", "border on page", c["border"], c["page"], 0)
		apart("line", "border on raised", c["border"], c["raised"], 0)
		for _, s := range []string{"page", "surface"} {
			apart("state", "selected tint over "+s+" | "+s, c["accent"].over(c[s], p.Tint.Selected), c[s], 0)
			apart("hover", "hover tint over "+s+" | "+s, c["accent"].over(c[s], p.Tint.Hover), c[s], 0)
		}
		apart("hover", "accent under the pointer | accent", st.accentHover, c["accent"], 0)
		apart("hover", "destructive under the pointer | destructive", st.destructiveHover, c["destructive"], 0)
		apart("hover", "secondary under the pointer | secondary", st.controlHover, st.control, 0)
		apart("edge", "secondary button | page", st.control, c["page"], 0)
		for _, pair := range [][2]string{
			{"success", "error"}, {"success", "warning"}, {"warning", "error"},
			{"accent", "success"}, {"accent", "error"},
		} {
			apart("meaning", pair[0]+" | "+pair[1], c[pair[0]], c[pair[1]], meaningTritan)
		}
	}
	term, err := terminalHolds(p)
	if err != nil {
		return nil, err
	}
	out = append(out, term...)
	ch, err := chartHolds(p)
	if err != nil {
		return nil, err
	}
	out = append(out, ch...)
	for _, r := range out {
		if math.IsNaN(r.contrast) {
			return nil, fmt.Errorf("%s %s: not a number", r.mode, r.what)
		}
	}
	return out, nil
}

// terminalHolds is the terminal's pairs, the same in either theme: its text on its background,
// each ANSI colour on it, and the foreground on what a program or a person paints under a cell -
// ANSI black and the selection. The cursor is a block the character under it reads on.
func terminalHolds(p palette) ([]result, error) {
	c, err := p.Terminal.colors()
	if err != nil {
		return nil, fmt.Errorf("terminal: %w", err)
	}
	var out []result
	hold := func(what string, fg, bg rgb, floor float64) {
		r := result{mode: "term", kind: "text", what: what, contrast: contrast(fg, bg), floor: floor}
		r.ok = r.contrast >= floor
		out = append(out, r)
	}
	bg := c["background"]
	hold("foreground on background", c["foreground"], bg, bodyText)
	for _, k := range ansi {
		switch k {
		case "black":
			hold("foreground on black", c["foreground"], c[k], bodyText)
		case "white", "bright_white":
			hold(k+" on background", c[k], bg, bodyText)
		default:
			hold(k+" on background", c[k], bg, text)
		}
	}
	hold("foreground on selection", c["foreground"], c["selection"], bodyText)
	hold("background on cursor", bg, c["cursor"], text)
	r := result{mode: "term", kind: "non-text", what: "cursor on background", contrast: contrast(c["cursor"], bg), floor: nonText}
	r.ok = r.contrast >= nonText
	out = append(out, r)
	return out, nil
}

// chartHolds is the chart's colours, one set for both modes: each a line or a fill that holds
// 3:1 on either mode's page, each series apart from the one drawn next to it, and each status
// apart from the others.
func chartHolds(p palette) ([]result, error) {
	series, statuses, err := p.Chart.colors()
	if err != nil {
		return nil, fmt.Errorf("chart: %w", err)
	}
	dark, err := colors(p.Dark)
	if err != nil {
		return nil, err
	}
	light, err := colors(p.Light)
	if err != nil {
		return nil, err
	}
	var out []result
	hold := func(what string, fg rgb) {
		for _, page := range []struct {
			mode string
			c    rgb
		}{{"dark", dark["page"]}, {"light", light["page"]}} {
			r := result{mode: page.mode, kind: "non-text", what: what + " on page", contrast: contrast(fg, page.c), floor: nonText}
			r.ok = r.contrast >= nonText
			out = append(out, r)
		}
	}
	apart := func(tier, what string, a, b rgb) {
		f := tiers[tier]
		r := result{mode: "chart", kind: tier, what: what, distance: map[string]float64{}, floor: f[0], worst: f[1], ok: true}
		for _, v := range views {
			d := deltaE00(v.fn(a.linear()).lab(), v.fn(b.linear()).lab())
			r.distance[v.name] = d
			floor := f[1]
			if v.name == "drawn" {
				floor = f[0]
			}
			if d < floor {
				r.ok = false
			}
		}
		out = append(out, r)
	}
	for i, c := range series {
		hold(fmt.Sprintf("series %d", i+1), c)
		if i > 0 {
			apart("meaning", fmt.Sprintf("series %d | series %d", i, i+1), series[i-1], c)
		}
	}
	for _, s := range chartStatuses {
		hold(s, statuses[s])
	}
	for i, a := range chartStatuses {
		for _, b := range chartStatuses[i+1:] {
			apart("meaning", a+" | "+b, statuses[a], statuses[b])
		}
	}
	return out, nil
}

// colors reads a mode's roles by name.
func colors(m mode) (map[string]rgb, error) {
	named := map[string]string{
		"page": m.Page, "surface": m.Surface, "raised": m.Raised, "overlay": m.Overlay,
		"border": m.Border, "border_strong": m.BorderStrong, "text": m.Text, "muted": m.Muted,
		"accent": m.Accent, "accent_text": m.AccentText, "on_accent": m.OnAccent, "label": m.Label,
		"success": m.Success, "warning": m.Warning, "error": m.Error, "destructive": m.Destructive,
	}
	return parseNamed(named)
}

// parseNamed parses every color by its name, failing on one that is missing.
func parseNamed(named map[string]string) (map[string]rgb, error) {
	out := map[string]rgb{}
	for name, hex := range named {
		if hex == "" {
			return nil, fmt.Errorf("%s has no color", name)
		}
		c, err := parseHex(hex)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		out[name] = c
	}
	return out, nil
}
