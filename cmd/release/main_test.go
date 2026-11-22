package main

import (
	"strings"
	"testing"
	"time"
)

func TestAReleaseIsNamedByItsDayAndItsTurn(t *testing.T) {
	// 22 November in UTC, and still the 21st in Buenos Aires.
	now := time.Date(2026, 11, 22, 1, 30, 0, 0, time.FixedZone("ART", -3*3600))
	cases := []struct {
		name   string
		at     []string
		all    []string
		want   string
		refuse string
	}{
		{name: "the first there is", want: "v20261122.01"},
		{name: "the first of a day", all: []string{"v20261121.01", "v20261121.02"}, want: "v20261122.01"},
		{name: "the third of a day", all: []string{"v20261122.02", "v20261122.01", "v20261121.07"}, want: "v20261122.03"},
		{name: "past the ninth", all: []string{"v20261122.09"}, want: "v20261122.10"},
		{
			name: "the screens' archives and a one-digit tag are no releases",
			at:   []string{"screens-v20261122.01-c2f0602021cc"},
			all:  []string{"screens-v20261122.01-c2f0602021cc", "v20261122.4", "v0.20261122.5"},
			want: "v20261122.01",
		},
		{name: "a commit released already", at: []string{"v20261121.02"}, all: []string{"v20261121.02"}, refuse: "released already, as v20261121.02"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			git := func(args ...string) ([]string, error) {
				if args[1] == "--points-at" {
					return c.at, nil
				}
				return c.all, nil
			}
			var out strings.Builder
			err := run([]string{"abc123"}, &out, git, now)
			if c.refuse != "" {
				if err == nil || !strings.Contains(err.Error(), c.refuse) {
					t.Fatalf("got %v and %q, want a refusal saying %q", err, out.String(), c.refuse)
				}
				return
			}
			if err != nil || out.String() != c.want+"\n" {
				t.Fatalf("got %q, %v; want %s", out.String(), err, c.want)
			}
		})
	}
}
