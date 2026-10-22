package main

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
)

// outputs are dist/'s files: one per place that draws Spin, each in the shape that place reads.
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
		"tint": p.Tint, "type": p.Type, "layout": p.Layout,
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

// tokensCSS is every role as --spin-<role>: dark on :root, light under data-theme="light", and
// the type and layout tokens beside them.
func tokensCSS(p palette, dark, light map[string]rgb) string {
	var b strings.Builder
	b.WriteString(header)
	block := func(sel string, c map[string]rgb) {
		fmt.Fprintf(&b, "%s {\n", sel)
		for _, k := range sortedKeys(c) {
			fmt.Fprintf(&b, "\t--spin-%s: %s;\n", strings.ReplaceAll(k, "_", "-"), c[k].hex())
		}
		fmt.Fprintf(&b, "\t--spin-accent-hover: %s;\n", c["accent"].over(c["page"], p.Tint.Hover).hex())
		fmt.Fprintf(&b, "\t--spin-accent-selected: %s;\n", c["accent"].over(c["page"], p.Tint.Selected).hex())
		b.WriteString("}\n")
	}
	block(":root", dark)
	block(":root[data-theme='light']", light)
	b.WriteString(":root {\n")
	fmt.Fprintf(&b, "\t--spin-font-sans: '%s', %s;\n", p.Type.Sans, p.Type.FallbackSans)
	fmt.Fprintf(&b, "\t--spin-font-mono: '%s', %s;\n", p.Type.Mono, p.Type.FallbackMono)
	fmt.Fprintf(&b, "\t--spin-text-reading: %dpx;\n\t--spin-text-dense: %dpx;\n", p.Type.ReadingSize, p.Type.DenseSize)
	fmt.Fprintf(&b, "\t--spin-line-height: %g;\n", p.Type.LineHeight)
	fmt.Fprintf(&b, "\t--spin-heading-weight: %d;\n\t--spin-heading-tracking: %gem;\n", p.Type.HeadingWeight, p.Type.HeadingTracking)
	fmt.Fprintf(&b, "\t--spin-measure: %dch;\n\t--spin-sidebar: %drem;\n", p.Layout.Measure, p.Layout.Sidebar)
	fmt.Fprintf(&b, "\t--spin-radius-sm: %dpx;\n\t--spin-radius: %dpx;\n\t--spin-radius-lg: %dpx;\n", p.Layout.RadiusSm, p.Layout.Radius, p.Layout.RadiusLg)
	fmt.Fprintf(&b, "\t--spin-target: %dpx;\n\t--spin-target-touch: %dpx;\n", p.Layout.Target, p.Layout.TargetTouch)
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
			{"gray-6", c["surface"].hex()},
			{"black", c["page"].hex()},
			{"banner-bg", c["accent"].over(c["surface"], 0.22).hex()},
			{"banner-text", c["text"].hex()},
		}
		if !isDark {
			vars = append(vars, [2]string{"gray-7", c["raised"].hex()})
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
		on := func(fill rgb) rgb {
			if contrast(c["page"], fill) >= contrast(rgb{1, 1, 1}, fill) {
				if isDark {
					return c["page"]
				}
				return c["text"]
			}
			return rgb{1, 1, 1}
		}
		vars := [][2]string{
			{"background", hsl(c["page"])}, {"foreground", hsl(c["text"])},
			{"card", hsl(card)}, {"card-foreground", hsl(c["text"])},
			{"muted", hsl(c["raised"])}, {"muted-foreground", hsl(c["muted"])},
			{"border", hsl(c["border"])}, {"input", hsl(c["border_strong"])}, {"ring", hsl(c["accent"])},
			{"primary", hsl(c["accent"])}, {"primary-foreground", hsl(c["on_accent"])},
			{"secondary", hsl(c["raised"])}, {"secondary-foreground", hsl(c["text"])},
			{"accent", hsl(c["accent"].over(c["page"], p.Tint.Selected))}, {"accent-foreground", hsl(c["accent_text"])},
			{"destructive", hsl(c["destructive"])}, {"destructive-foreground", hsl(rgb{1, 1, 1})},
			{"success", hsl(c["success"])}, {"success-foreground", hsl(on(c["success"]))},
			{"warning", hsl(c["warning"])}, {"warning-foreground", hsl(on(c["warning"]))},
		}
		fmt.Fprintf(&b, "%s {\n", sel)
		for _, v := range vars {
			fmt.Fprintf(&b, "\t--%s: %s;\n", v[0], v[1])
		}
		if !isDark {
			fmt.Fprintf(&b, "\t--radius: %grem;\n", float64(p.Layout.Radius)/16)
		}
		b.WriteString("}\n")
	}
	block(":root", light, false)
	block(".dark", dark, true)
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
