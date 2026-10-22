package main

import (
	"strings"
	"testing"
)

func TestWholeRefusesACatalogWithAGap(t *testing.T) {
	workspaces := screen{Name: "workspaces", Kind: "page"}
	all := []string{
		"workspaces.dark.desktop.png", "workspaces.dark.mobile.png",
		"workspaces.light.desktop.png", "workspaces.light.mobile.png",
	}
	cases := []struct {
		name    string
		screens []screen
		images  []string
		audit   audit
		refused string
	}{
		{name: "every theme at every width", screens: []screen{workspaces}, images: all},
		{name: "a width not taken", screens: []screen{workspaces}, images: all[1:], refused: "workspaces.dark.desktop.png was not taken"},
		{name: "an image of no screen", screens: []screen{workspaces}, images: append(all, "gone.light.desktop.png"), refused: "gone.light.desktop.png is no screen"},
		{name: "a kind that is neither", screens: []screen{{Name: "workspaces", Kind: "drawer"}}, images: all, refused: `"drawer", not a page or a dialog`},
		{name: "one name twice", screens: []screen{workspaces, workspaces}, images: all, refused: "two screens are named workspaces"},
		{
			name: "a finding about no screen", screens: []screen{workspaces}, images: all,
			audit:   audit{Findings: []finding{{Screen: "gone", Theme: "dark", Width: "desktop"}}},
			refused: "a finding about gone.dark.desktop",
		},
		{
			name: "a finding in a theme there is not", screens: []screen{workspaces}, images: all,
			audit:   audit{Findings: []finding{{Screen: "workspaces", Theme: "sepia", Width: "desktop"}}},
			refused: "a finding about workspaces.sepia.desktop",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := whole(c.screens, c.images, c.audit)
			if c.refused == "" {
				if err != nil {
					t.Fatalf("refused a whole catalog: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), c.refused) {
				t.Fatalf("got %v, want a refusal saying %q", err, c.refused)
			}
		})
	}
}
