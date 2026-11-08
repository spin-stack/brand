package main

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
)

// outputs are dist/'s files: one per place that draws Spinbox, each in the shape that place reads.
func outputs(p palette, results []result) (map[string][]byte, error) {
	dark, err := colors(p.Dark)
	if err != nil {
		return nil, err
	}
	light, err := colors(p.Light)
	if err != nil {
		return nil, err
	}
	files := map[string][]byte{}

	j, err := json.MarshalIndent(map[string]any{
		"brand": p.Brand, "dark": hexes(dark), "light": hexes(light),
		"tint": p.Tint, "type": p.Type, "layout": p.Layout, "controls": p.Controls, "elevation": p.Elevation, "motion": p.Motion,
	}, "", "  ")
	if err != nil {
		return nil, err
	}
	files["tokens.json"] = append(j, '\n')
	files["tokens.css"] = []byte(tokensCSS(p, dark, light))
	files["starlight.css"] = []byte(starlightCSS(p, dark, light))
	files["dashboard.css"] = []byte(dashboardCSS(p, dark, light))
	files["perception.js"] = []byte(perceptionJS(results))
	files["simulate.svg"] = []byte(simulateSVG())
	return files, nil
}

const header = "/* Written by `task palette` from palette.toml (spin-stack/brand). Edit palette.toml, not this file. */\n"

func hexes(c map[string]rgb) map[string]string {
	out := map[string]string{}
	for k, v := range c {
		out[k] = v.hex()
	}
	return out
}

func sortedKeys(m map[string]rgb) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// tokensCSS is every role as --spin-<role>: dark on :root and .spin-dark, light under
// data-theme="light", and the type, layout and motion tokens beside them.
func tokensCSS(p palette, dark, light map[string]rgb) string {
	var b strings.Builder
	b.WriteString(header)
	block := func(sel string, c map[string]rgb, shadow string, isDark bool) {
		fmt.Fprintf(&b, "%s {\n", sel)
		for _, k := range sortedKeys(c) {
			fmt.Fprintf(&b, "\t--spin-%s: %s;\n", strings.ReplaceAll(k, "_", "-"), c[k].hex())
		}
		st := controlStates(p, c, isDark)
		for _, v := range [][2]string{
			{"accent-fill-hover", st.accentHover.hex()}, {"accent-fill-active", st.accentActive.hex()},
			{"destructive-hover", st.destructiveHover.hex()}, {"destructive-active", st.destructiveActive.hex()},
			{"control", st.control.hex()}, {"control-hover", st.controlHover.hex()},
		} {
			fmt.Fprintf(&b, "\t--spin-%s: %s;\n", v[0], v[1])
		}
		fmt.Fprintf(&b, "\t--spin-shadow-overlay: %s;\n", shadow)
		// A state's badge: its colour at tint.badge over whatever is under it.
		for _, k := range []string{"success", "warning", "error", "accent-text"} {
			fmt.Fprintf(&b, "\t--spin-%s-badge: color-mix(in srgb, var(--spin-%s) %g%%, transparent);\n", k, k, p.Tint.Badge*100)
		}
		fmt.Fprintf(&b, "\t--spin-accent-hover: %s;\n", c["accent"].over(c["page"], p.Tint.Hover).hex())
		fmt.Fprintf(&b, "\t--spin-accent-selected: %s;\n", c["accent"].over(c["page"], p.Tint.Selected).hex())
		b.WriteString("}\n")
	}
	// .spin-dark is a part that is dark in either theme - a terminal, a drawing on the navy - and
	// takes the dark roles in place of the page's.
	block(":root, .spin-dark", dark, p.Elevation.Dark, true)
	block(":root[data-theme='light']", light, p.Elevation.Light, false)
	b.WriteString(":root {\n")
	names := make([]string, 0, len(p.Brand))
	for k := range p.Brand {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		fmt.Fprintf(&b, "\t--spin-brand-%s: %s;\n", k, strings.ToUpper(p.Brand[k]))
	}
	fmt.Fprintf(&b, "\t--spin-wordmark: %q;\n", p.Type.Wordmark)
	fmt.Fprintf(&b, "\t--spin-font-sans: '%s', %s;\n", p.Type.Sans, p.Type.FallbackSans)
	fmt.Fprintf(&b, "\t--spin-font-mono: '%s', %s;\n", p.Type.Mono, p.Type.FallbackMono)
	fmt.Fprintf(&b, "\t--spin-text-reading: %dpx;\n\t--spin-text-dense: %dpx;\n", p.Type.ReadingSize, p.Type.DenseSize)
	fmt.Fprintf(&b, "\t--spin-line-height: %g;\n", p.Type.LineHeight)
	fmt.Fprintf(&b, "\t--spin-heading-weight: %d;\n\t--spin-heading-tracking: %gem;\n", p.Type.HeadingWeight, p.Type.HeadingTracking)
	fmt.Fprintf(&b, "\t--spin-font-heading: var(--spin-font-%s);\n", p.Type.Heading)
	fmt.Fprintf(&b, "\t--spin-numeric: %s;\n", p.Type.Numeric)
	for _, k := range []string{"display", "h1", "h2", "h3"} {
		fmt.Fprintf(&b, "\t--spin-text-%s: %s;\n", k, fluid(p, k))
	}
	fmt.Fprintf(&b, "\t--spin-measure: %dch;\n\t--spin-sidebar: %drem;\n", p.Layout.Measure, p.Layout.Sidebar)
	fmt.Fprintf(&b, "\t--spin-radius-sm: %dpx;\n\t--spin-radius: %dpx;\n\t--spin-radius-lg: %dpx;\n", p.Layout.RadiusSm, p.Layout.Radius, p.Layout.RadiusLg)
	fmt.Fprintf(&b, "\t--spin-target: %dpx;\n\t--spin-target-touch: %dpx;\n", p.Layout.Target, p.Layout.TargetTouch)
	fmt.Fprintf(&b, "\t--spin-control-sm: %dpx;\n\t--spin-control-md: %dpx;\n\t--spin-control-lg: %dpx;\n", p.Controls.Sm, p.Controls.Md, p.Controls.Lg)
	fmt.Fprintf(&b, "\t--spin-control-icon: %dpx;\n\t--spin-control-gap: %dpx;\n", p.Controls.Icon, p.Controls.Gap)
	fmt.Fprintf(&b, "\t--spin-width-page: %dpx;\n", p.Layout.WidthPage)
	fmt.Fprintf(&b, "\t--spin-duration-fast: %dms;\n\t--spin-duration: %dms;\n", p.Motion.Fast, p.Motion.Base)
	fmt.Fprintf(&b, "\t--spin-duration-slow: %dms;\n\t--spin-duration-draw: %dms;\n", p.Motion.Slow, p.Motion.Draw)
	fmt.Fprintf(&b, "\t--spin-ease: %s;\n", p.Motion.Ease)
	b.WriteString("}\n")
	return b.String()
}

// starlightCSS is the palette in Starlight's own variables, for the website: Starlight draws
// body text in gray-2, headings in white, the sidebar on gray-6, the page on black, and the
// current page's link in accent-high on accent-low.
func starlightCSS(p palette, dark, light map[string]rgb) string {
	var b strings.Builder
	b.WriteString(header)
	white := rgb{1, 1, 1}
	black := rgb{0, 0, 0}
	block := func(sel string, c map[string]rgb, isDark bool) {
		// Dark's sidebar, header and inline code are gray-6, which is the surface; light's
		// sidebar is gray-6 too, a step under its header (gray-7).
		sidebar := c["surface"]
		if !isDark {
			sidebar = c["raised"]
		}
		strong := white.over(c["text"], 0.6)
		if !isDark {
			strong = black.over(c["text"], 0.4)
		}
		vars := [][2]string{
			{"accent-low", c["accent"].over(c["page"], p.Tint.Selected).hex()},
			{"accent", c["accent"].hex()},
			{"accent-high", c["accent_text"].over(c["text"], 0.5).hex()},
			{"white", strong.hex()},
			{"gray-1", c["text"].over(strong, 0.5).hex()},
			{"gray-2", c["text"].hex()},
			{"gray-3", c["muted"].hex()},
			{"gray-4", c["border_strong"].hex()},
			{"gray-5", c["border"].hex()},
			{"gray-6", sidebar.hex()},
			{"black", c["page"].hex()},
			{"banner-bg", c["accent"].over(c["surface"], 0.22).hex()},
			{"banner-text", c["text"].hex()},
		}
		if !isDark {
			// Light's header is gray-7, and its sidebar and its hairlines gray-6: the sidebar a
			// step under the header, and a hairline that shows on white.
			vars = append(vars, [2]string{"gray-7", c["surface"].hex()})
		}
		fmt.Fprintf(&b, "%s {\n", sel)
		for _, v := range vars {
			fmt.Fprintf(&b, "\t--sl-color-%s: %s;\n", v[0], v[1])
		}
		b.WriteString("}\n")
	}
	block(":root", dark, true)
	block(":root[data-theme='light']", light, false)
	fmt.Fprintf(&b, ":root {\n\t--sl-font: '%s', %s;\n\t--sl-font-mono: '%s', %s;\n}\n",
		p.Type.Sans, p.Type.FallbackSans, p.Type.Mono, p.Type.FallbackMono)
	// The headings' sizes, fluid from a phone's width to a desktop's: Starlight's own scale is a
	// step for each breakpoint, and between them a heading jumped.
	b.WriteString(":root {\n")
	for _, h := range [][2]string{{"h1", "display"}, {"h2", "h1"}, {"h3", "h2"}, {"h4", "h3"}} {
		fmt.Fprintf(&b, "\t--sl-text-%s: %s;\n", h[0], fluid(p, h[1]))
	}
	b.WriteString("}\n")
	return b.String()
}

// dashboardCSS is the palette in the dashboard's variables (shadcn's, as Tailwind's theme reads
// them in dashboard/src/styles/globals.css): HSL triplets, light on :root and dark under .dark.
func dashboardCSS(p palette, dark, light map[string]rgb) string {
	var b strings.Builder
	b.WriteString(header)
	block := func(sel string, c map[string]rgb, isDark bool) {
		card := c["page"]
		if isDark {
			card = c["surface"]
		}
		// The label on a fill: white, or the theme's darkest ink - the dark page, the light text -
		// whichever reads on it.
		ink := c["text"]
		if isDark {
			ink = c["page"]
		}
		on := func(fill rgb) rgb {
			if contrast(ink, fill) >= contrast(rgb{1, 1, 1}, fill) {
				return ink
			}
			return rgb{1, 1, 1}
		}
		// Hover and the navigation's selection are neutral: the accent is the primary action's
		// and the focus ring's alone. Raised in both themes: light's surface was a step off white
		// that a selected item in the sidebar did not show.
		neutral := c["raised"]
		st := controlStates(p, c, isDark)
		// A step off the page - a code block, a secondary button's fill: raised in dark; in light
		// the surface, since raised there read as a grey slab and a button on it as disabled.
		step := c["surface"]
		if isDark {
			step = c["raised"]
		}
		vars := [][2]string{
			{"background", hsl(c["page"])}, {"foreground", hsl(c["text"])},
			{"card", hsl(card)}, {"card-foreground", hsl(c["text"])},
			// A menu's or a dialog's plane: overlay, the lightest step in dark, so that it is lifted
			// by its own lightness rather than by a shadow; white in light.
			{"popover", hsl(c["overlay"])}, {"popover-foreground", hsl(c["text"])},
			{"muted", hsl(step)}, {"muted-foreground", hsl(c["muted"])},
			{"border", hsl(c["border"])}, {"input", hsl(c["border_strong"])}, {"ring", hsl(c["accent"])},
			{"primary", hsl(c["accent"])}, {"primary-foreground", hsl(c["on_accent"])},
			{"primary-hover", hsl(st.accentHover)}, {"primary-active", hsl(st.accentActive)},
			{"destructive-hover", hsl(st.destructiveHover)}, {"destructive-active", hsl(st.destructiveActive)},
			{"secondary-hover", hsl(st.controlHover)},
			{"secondary", hsl(st.control)}, {"secondary-foreground", hsl(c["text"])},
			{"accent", hsl(neutral)}, {"accent-foreground", hsl(c["text"])},
			{"destructive", hsl(c["destructive"])}, {"destructive-foreground", hsl(on(c["destructive"]))},
			{"error", hsl(c["error"])}, {"label", hsl(c["label"])},
			// A badge's fill: the state at tint.badge over the card it sits on.
			{"success-subtle", hsl(c["success"].over(card, p.Tint.Badge))},
			{"warning-subtle", hsl(c["warning"].over(card, p.Tint.Badge))},
			{"error-subtle", hsl(c["error"].over(card, p.Tint.Badge))},
			{"success", hsl(c["success"])}, {"success-foreground", hsl(on(c["success"]))},
			{"warning", hsl(c["warning"])}, {"warning-foreground", hsl(on(c["warning"]))},
		}
		fmt.Fprintf(&b, "%s {\n", sel)
		for _, v := range vars {
			fmt.Fprintf(&b, "\t--%s: %s;\n", v[0], v[1])
		}
		shadow := p.Elevation.Light
		if isDark {
			shadow = p.Elevation.Dark
		}
		fmt.Fprintf(&b, "\t--shadow-overlay: %s;\n", shadow)
		b.WriteString("}\n")
	}
	block(":root", light, false)
	block(".dark", dark, true)

	// What does not change with the theme. The dashboard is dense - tables of runners and
	// workspaces - and square enough that a table reads as one: every corner is the small radius.
	// Its theme (dashboard/src/styles/globals.css) maps Tailwind's names onto these.
	b.WriteString(":root {\n")
	fmt.Fprintf(&b, "\t--radius: %gpx;\n", float64(p.Layout.RadiusSm))
	fmt.Fprintf(&b, "\t--spin-control-sm: %dpx;\n\t--spin-control-md: %dpx;\n\t--spin-control-lg: %dpx;\n", p.Controls.Sm, p.Controls.Md, p.Controls.Lg)
	fmt.Fprintf(&b, "\t--spin-control-icon: %dpx;\n\t--spin-control-gap: %dpx;\n", p.Controls.Icon, p.Controls.Gap)
	// How wide a page may grow (the dashboard's page-width.ts reads it).
	fmt.Fprintf(&b, "\t--spin-width-page: %dpx;\n", p.Layout.WidthPage)
	// The dim behind a dialog: the dark page, in either theme, at the palette's scrim - a neutral
	// dim, not the mark's navy, which tinted everything behind a dialog blue.
	dim := dark["page"]
	b255 := func(v float64) int { return int(math.Round(v * 255)) }
	fmt.Fprintf(&b, "\t--scrim: rgb(%d %d %d / %g);\n", b255(dim.r), b255(dim.g), b255(dim.b), p.Tint.Scrim)
	// The word the sidebar writes beside the mark: the app's catalog audit holds its text to this.
	fmt.Fprintf(&b, "\t--spin-wordmark: %q;\n", p.Type.Wordmark)
	fmt.Fprintf(&b, "\t--spin-font-sans: '%s', %s;\n\t--spin-font-mono: '%s', %s;\n", p.Type.Sans, p.Type.FallbackSans, p.Type.Mono, p.Type.FallbackMono)
	for _, k := range []string{"xs", "sm", "base", "lg", "xl", "2xl", "3xl"} {
		step, ok := p.Type.DenseScale[k]
		if !ok {
			panic("palette.toml's dense_scale has no " + k)
		}
		fmt.Fprintf(&b, "\t--spin-text-%s: %dpx;\n\t--spin-text-%s-lh: %dpx;\n", k, step[0], k, step[1])
	}
	fmt.Fprintf(&b, "\t--spin-heading-tracking: %gem;\n", p.Type.HeadingTracking)
	fmt.Fprintf(&b, "\t--spin-numeric: %s;\n", p.Type.Numeric)
	fmt.Fprintf(&b, "\t--spin-duration-fast: %dms;\n\t--spin-duration: %dms;\n", p.Motion.Fast, p.Motion.Base)
	fmt.Fprintf(&b, "\t--spin-duration-slow: %dms;\n\t--spin-ease: %s;\n", p.Motion.Slow, p.Motion.Ease)
	b.WriteString("}\n")
	return b.String()
}

// hsl writes a color as shadcn's variables hold it: "H S% L%", for hsl(var(--x)).
func hsl(c rgb) string {
	mx, mn := math.Max(c.r, math.Max(c.g, c.b)), math.Min(c.r, math.Min(c.g, c.b))
	l := (mx + mn) / 2
	var h, s float64
	if d := mx - mn; d > 0 {
		s = d / (1 - math.Abs(2*l-1))
		switch mx {
		case c.r:
			h = math.Mod((c.g-c.b)/d, 6)
		case c.g:
			h = (c.b-c.r)/d + 2
		default:
			h = (c.r-c.g)/d + 4
		}
		h *= 60
		if h < 0 {
			h += 360
		}
	}
	return fmt.Sprintf("%.1f %.1f%% %.1f%%", h, s*100, l*100)
}

// perceptionJS is every pair and how it held, for preview.html's table.
func perceptionJS(results []result) string {
	type row struct {
		Mode, Kind, What string
		Contrast         float64            `json:",omitempty"`
		Distance         map[string]float64 `json:",omitempty"`
		Floor, Worst     float64
		OK               bool
	}
	rows := make([]row, 0, len(results))
	for _, r := range results {
		d := map[string]float64{}
		for k, v := range r.distance {
			d[k] = math.Round(v*10) / 10
		}
		if r.distance == nil {
			d = nil
		}
		rows = append(rows, row{r.mode, r.kind, r.what, math.Round(r.contrast*100) / 100, d, r.floor, r.worst, r.ok})
	}
	b, _ := json.Marshal(rows)
	return "// Written by `task palette`: every pair palette.toml is held to, and how it held.\nexport const perception = " + string(b) + ";\n"
}

// simulateSVG is each deficiency's matrix as an SVG filter, for preview.html's "See as": the
// same matrices the checks use, applied in linear RGB as the model is.
func simulateSVG() string {
	var b strings.Builder
	b.WriteString("<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"0\" height=\"0\" style=\"position:absolute\">\n")
	b.WriteString("<!-- Written by `task palette`: Machado, Oliveira and Fernandes 2009, severity 1.0, in linear RGB. -->\n")
	for _, f := range []struct {
		id string
		m  [3][3]float64
	}{{"protan", protan}, {"deutan", deutan}, {"tritan", tritan}} {
		var v []string
		for _, row := range f.m {
			v = append(v, fmt.Sprintf("%g %g %g 0 0", row[0], row[1], row[2]))
		}
		v = append(v, "0 0 0 1 0")
		fmt.Fprintf(&b, "<filter id=\"%s\" color-interpolation-filters=\"linearRGB\"><feColorMatrix type=\"matrix\" values=\"%s\"/></filter>\n", f.id, strings.Join(v, " "))
	}
	b.WriteString("</svg>\n")
	return b.String()
}

// fluid is a size that grows with the viewport from the phone's width to the desktop's, as a
// clamp() in rem, so that a reader's zoom still scales it: [min, max] in px at those widths.
func fluid(p palette, name string) string {
	s, ok := p.Type.Fluid[name]
	if !ok {
		panic("palette.toml's type.fluid has no " + name)
	}
	lo, hi := float64(p.Layout.Phone), float64(p.Layout.Desktop)
	slope := float64(s[1]-s[0]) / (hi - lo)
	base := float64(s[0]) - slope*lo
	return fmt.Sprintf("clamp(%.4grem, %.4grem + %.4gvw, %.4grem)", float64(s[0])/16, base/16, slope*100, float64(s[1])/16)
}
