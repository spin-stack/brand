package main

import (
	"encoding/json"
	"fmt"
	"testing"
)

// A trace's service the palette cannot stand behind fails by its own pair, in the theme that
// broke it: two slots a reader cannot tell apart, a slot that is a failure's colour or red, one
// too pale to carry the count on the page, one that is lost on a selected row.
func TestAServiceThatIsNotToldApartFails(t *testing.T) {
	for _, tc := range []struct {
		pair  string
		plant func(services []string, m mode)
	}{
		{"service 1 | service 2", func(s []string, _ mode) { s[1] = s[0] }},
		{"service 3 | service 8", func(s []string, _ mode) { s[7] = s[2] }},
		{"service 4 | error", func(s []string, m mode) { s[3] = m.Error }},
		{"service 7 | destructive", func(s []string, m mode) { s[6] = m.Destructive }},
		{"service 5 on page", func(s []string, m mode) { s[4] = m.Border }},
		{"service 6 on raised", func(s []string, m mode) { s[5] = m.Raised }},
	} {
		for _, theme := range []string{"dark", "light"} {
			t.Run(theme+" "+tc.pair, func(t *testing.T) {
				p := readPalette(t)
				if got := pairs(t, p, tc.pair); !got["dark"] || !got["light"] {
					t.Fatalf("as written the pair holds in %v, want both themes", got)
				}
				// The plant writes into a copy: the two themes' sets are separate lists.
				p.Trace.Dark = append([]string{}, p.Trace.Dark...)
				p.Trace.Light = append([]string{}, p.Trace.Light...)
				other := "light"
				if theme == "dark" {
					tc.plant(p.Trace.Dark, p.Dark)
				} else {
					tc.plant(p.Trace.Light, p.Light)
					other = "dark"
				}
				got := pairs(t, p, tc.pair)
				if got[theme] {
					t.Errorf("%s: the pair holds with the plant", theme)
				}
				if !got[other] {
					t.Errorf("%s: the pair fails for a change to %s", other, theme)
				}
			})
		}
	}
}

// The dashboard picks a slot by a name's hash in either theme, so a theme with a slot the other
// has not is a service with no colour there.
func TestTheThemesHaveTheSameServices(t *testing.T) {
	p := readPalette(t)
	p.Trace.Light = p.Trace.Light[:len(p.Trace.Light)-1]
	if _, err := holds(p); err == nil {
		t.Error("a palette with a service fewer in light than in dark holds")
	}
}

// The dashboard and the laboratory draw a service in the theme's own colour only if the file each
// copies says it under the theme's selector: every slot palette.toml has, and no other.
func TestEveryOutputCarriesEachThemesServices(t *testing.T) {
	p := readPalette(t)
	files := written(t, p)
	for _, tc := range []struct {
		file, prefix, dark, light string
	}{
		{"dashboard.css", "trace-service-", ".dark", ":root"},
		{"tokens.css", "spin-trace-service-", ":root, .spin-dark", ":root[data-theme='light']"},
	} {
		for selector, want := range map[string][]string{tc.dark: p.Trace.Dark, tc.light: p.Trace.Light} {
			got := map[string]string{}
			for _, b := range blocks(files[tc.file]) {
				if b.selector != selector {
					continue
				}
				for k, v := range b.vars {
					if len(k) > len(tc.prefix) && k[:len(tc.prefix)] == tc.prefix {
						got[k] = v
					}
				}
			}
			if len(got) != len(want) {
				t.Errorf("%s %s: %d services, want %d", tc.file, selector, len(got), len(want))
			}
			for i, hex := range want {
				if v := got[fmt.Sprintf("%s%d", tc.prefix, i+1)]; v != hex {
					t.Errorf("%s %s: service %d is %q, want %s", tc.file, selector, i+1, v, hex)
				}
			}
		}
	}
	var tokens struct {
		Trace struct{ Dark, Light []string }
	}
	if err := json.Unmarshal(files["tokens.json"], &tokens); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(tokens.Trace.Dark) != fmt.Sprint(p.Trace.Dark) || fmt.Sprint(tokens.Trace.Light) != fmt.Sprint(p.Trace.Light) {
		t.Errorf("tokens.json's trace is %v, want palette.toml's", tokens.Trace)
	}
}
