package main

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

// Every label the dashboard sets on a fill - text on a primary button, a success badge, a
// destructive one - reads at 4.5:1, in both themes, as dashboardCSS writes them from the palette.
// The rule it holds: on() once compared a fill with the light page, white, against white, and
// gave the light success and warning fills dark text at 2.9:1.
func TestTheDashboardsLabelsReadOnTheirFills(t *testing.T) {
	var p palette
	if _, err := toml.DecodeFile("../../palette.toml", &p); err != nil {
		t.Fatal(err)
	}
	dark, err := colors(p.Dark)
	if err != nil {
		t.Fatal(err)
	}
	light, err := colors(p.Light)
	if err != nil {
		t.Fatal(err)
	}
	css := dashboardCSS(p, dark, light)
	blocks := regexp.MustCompile(`(?s)(:root|\.dark) \{(.*?)\}`).FindAllStringSubmatch(css, -1)
	var themes int
	for _, blk := range blocks {
		vars := map[string]rgb{}
		for _, m := range regexp.MustCompile(`--([a-z-]+): ([\d.]+) ([\d.]+)% ([\d.]+)%;`).FindAllStringSubmatch(blk[2], -1) {
			h, _ := strconv.ParseFloat(m[2], 64)
			s, _ := strconv.ParseFloat(m[3], 64)
			l, _ := strconv.ParseFloat(m[4], 64)
			vars[m[1]] = fromHSL(h, s/100, l/100)
		}
		if len(vars) == 0 {
			continue
		}
		themes++
		for name, fg := range vars {
			fill, ok := strings.CutSuffix(name, "-foreground")
			if !ok || fill == "card" || fill == "muted" || fill == "accent" || fill == "secondary" {
				continue
			}
			if c := contrast(fg, vars[fill]); c < 4.5 {
				t.Errorf("%s: --%s on --%s is %.2f:1, under 4.5", blk[1], name, fill, c)
			}
		}
	}
	if themes != 2 {
		t.Fatalf("found %d themes with colors in dashboard.css, want light and dark", themes)
	}
}

func fromHSL(h, s, l float64) rgb {
	c := (1 - abs(2*l-1)) * s
	hp := h / 60
	x := c * (1 - abs(mod(hp, 2)-1))
	var r, g, b float64
	switch {
	case hp < 1:
		r, g = c, x
	case hp < 2:
		r, g = x, c
	case hp < 3:
		g, b = c, x
	case hp < 4:
		g, b = x, c
	case hp < 5:
		r, b = x, c
	default:
		r, b = c, x
	}
	m := l - c/2
	return rgb{r + m, g + m, b + m}
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func mod(a, b float64) float64 {
	return a - b*float64(int(a/b))
}
