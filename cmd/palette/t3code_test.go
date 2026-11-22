package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

// agentClient is dist/t3code.css as the palette makes it, read back as a browser would: what
// each theme's rule sets, and what the rule for both sets.
type agentClient struct {
	p           palette
	light, dark map[string]string
	shared      map[string]string
	// The palette's own colours in each theme, by hex: every role, the states controls.go
	// derives, and in dark the terminal's.
	roles map[string]map[string]bool
}

func readAgentClient(t *testing.T) agentClient {
	t.Helper()
	a := agentClient{}
	if _, err := toml.DecodeFile("../../palette.toml", &a.p); err != nil {
		t.Fatal(err)
	}
	dark, err := colors(a.p.Dark)
	if err != nil {
		t.Fatal(err)
	}
	light, err := colors(a.p.Light)
	if err != nil {
		t.Fatal(err)
	}
	term, err := a.p.Terminal.colors()
	if err != nil {
		t.Fatal(err)
	}
	a.roles = map[string]map[string]bool{"light": {}, "dark": {}}
	for name, c := range map[string]map[string]rgb{"light": light, "dark": dark} {
		for _, v := range c {
			a.roles[name][v.hex()] = true
		}
		st := controlStates(a.p, c, name == "dark")
		for _, v := range []rgb{st.accentHover, st.accentActive, st.control, st.controlHover, st.outline, st.row, st.rowAction} {
			a.roles[name][v.hex()] = true
		}
	}
	for _, v := range term {
		a.roles["dark"][v.hex()] = true
	}

	rules := regexp.MustCompile(`(?s)(\S+) \{\n(.*?)\}\n`).FindAllStringSubmatch(t3codeCSS(a.p, dark, light, term), -1)
	for _, r := range rules {
		set := map[string]string{}
		for _, d := range regexp.MustCompile(`\t--([a-z0-9-]+): ([^;]+);\n`).FindAllStringSubmatch(r[2], -1) {
			set[d[1]] = d[2]
		}
		switch {
		case r[1] == "html.dark:root" && a.dark == nil:
			a.dark = set
		case r[1] == "html:root" && a.light == nil:
			a.light = set
		case r[1] == "html:root" && a.shared == nil:
			a.shared = set
		default:
			t.Fatalf("t3code.css has a rule for %s the test does not know", r[1])
		}
	}
	if len(a.light) == 0 || len(a.dark) == 0 || len(a.shared) == 0 {
		t.Fatalf("t3code.css sets %d variables in light, %d in dark and %d for both: the test no longer reads how it is written", len(a.light), len(a.dark), len(a.shared))
	}
	return a
}

// A variable the fork does not declare is one nothing in the client reads: the stylesheet loads,
// sets it, and the client draws its own colour. testdata/t3code.vars is the fork's names at one
// commit (`task agent-client` writes it from a checkout), so a name the fork renamed fails here
// when the fork is read again, and not as a zinc panel in the dashboard.
func TestTheAgentClientsVariablesAreTheForks(t *testing.T) {
	raw, err := os.ReadFile("testdata/t3code.vars")
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`(?m)^# commit [0-9a-f]{40}$`).Match(raw) {
		t.Fatal("testdata/t3code.vars does not say which commit of the fork it was read at")
	}
	fork := map[string]bool{}
	for _, line := range strings.Split(string(raw), "\n") {
		if name, ok := strings.CutPrefix(line, "--"); ok {
			fork[name] = true
		}
	}
	a := readAgentClient(t)
	for rule, set := range map[string]map[string]string{"light": a.light, "dark": a.dark, "both themes": a.shared} {
		for name := range set {
			if !fork[name] {
				t.Errorf("%s sets --%s, which the fork does not declare", rule, name)
			}
		}
	}
	// The fork's dark rule outranks light's here, so a name only light sets is, in dark, the
	// fork's own colour where it has one for dark and light's where it has none.
	for name := range a.light {
		if _, ok := a.dark[name]; !ok {
			t.Errorf("light sets --%s and dark does not", name)
		}
	}
}

// Every colour the client is given is one the palette has, in that theme: a role, or a state
// controls.go derives from the roles. A colour written here by hand is one no check holds.
func TestTheAgentClientsColoursAreThePalettes(t *testing.T) {
	a := readAgentClient(t)
	tint := regexp.MustCompile(`^color-mix\(in srgb, var\(--([a-z-]+)\) ([\d.]+)%, transparent\)$`)
	for theme, set := range map[string]map[string]string{"light": a.light, "dark": a.dark} {
		for name, value := range set {
			if m := tint.FindStringSubmatch(value); m != nil {
				if _, ok := set[m[1]]; !ok || m[2] != fmt.Sprint(a.p.Tint.Badge*100) {
					t.Errorf("%s: --%s is %s, which is not a colour this theme sets at tint.badge", theme, name, value)
				}
				continue
			}
			if !a.roles[theme][value] {
				t.Errorf("%s: --%s is %s, which is no role of the palette in that theme", theme, name, value)
			}
		}
	}
	radius := fmt.Sprintf("%dpx", a.p.Layout.RadiusSm)
	for name, value := range a.shared {
		ref, isRef := strings.CutPrefix(value, "var(--")
		ref = strings.TrimSuffix(ref, ")")
		switch {
		case name == "radius" || name == "control-radius":
			if value != radius {
				t.Errorf("--%s is %s, and the dashboard's corner (layout.radius_sm) is %s", name, value, radius)
			}
		case name == "font-sans" || name == "font-mono":
			// TestEveryFaceAnOutputNamesIsShipped holds the face each leads with.
		case name == "surface-grain":
			if value != "none" {
				t.Errorf("--surface-grain is %s: the page is its colour and nothing over it", value)
			}
		case isRef:
			if a.light[ref] == "" || a.dark[ref] == "" {
				t.Errorf("--%s is %s, which the themes do not both set", name, value)
			}
		default:
			t.Errorf("--%s is %s, and the test does not know what of the palette that is", name, value)
		}
	}
}

// What the conversation writes reads on what it is written on, as the client pairs its
// variables: the product's text at 7:1 on each plane a message, a card, a menu or the composer
// is, and at 4.5:1 a label on the action, secondary text, a link, and a state's words.
func TestTheAgentClientsTextReadsOnItsGrounds(t *testing.T) {
	a := readAgentClient(t)
	planes := []string{"background", "card", "surface-raised", "message-surface"}
	pairs := []struct {
		text    string
		grounds []string
		floor   float64
	}{
		{"foreground", append([]string{"popover", "secondary", "muted", "accent"}, planes...), 7},
		{"card-foreground", []string{"card"}, 7},
		{"popover-foreground", []string{"popover"}, 7},
		{"secondary-foreground", []string{"secondary"}, 7},
		{"accent-foreground", []string{"accent"}, 7},
		{"primary-foreground", []string{"primary", "message-action-hover"}, 4.5},
		{"muted-foreground", planes, 4.5},
		{"info-foreground", planes, 4.5},
		{"update-foreground", planes, 4.5},
		{"error-foreground", planes, 4.5},
		{"warning-foreground", planes, 4.5},
		{"success-foreground", planes, 4.5},
	}
	for theme, set := range map[string]map[string]string{"light": a.light, "dark": a.dark} {
		for _, pair := range pairs {
			fg, err := parseHex(set[pair.text])
			if err != nil {
				t.Fatalf("%s: --%s: %v", theme, pair.text, err)
			}
			for _, ground := range pair.grounds {
				bg, err := parseHex(set[ground])
				if err != nil {
					t.Fatalf("%s: --%s: %v", theme, ground, err)
				}
				if c := contrast(fg, bg); c < pair.floor {
					t.Errorf("%s: --%s on --%s is %.2f:1, under %g", theme, pair.text, ground, c, pair.floor)
				}
			}
		}
	}
}
