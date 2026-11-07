package main

// The fills a button takes under the pointer and when pressed, derived from the roles so that a
// consumer does not invent them (the dashboard wrote primary/90: the accent at 90% over the page,
// lighter in dark and darker in light, and its white label under 4.5:1 in one of them).
//
// A solid fill - the accent, the destructive - darkens: its label is white, and a darker fill
// keeps it legible in both themes. A quiet fill - a secondary button's - steps one surface up.
type states struct {
	accentHover, accentActive           rgb
	destructiveHover, destructiveActive rgb
	control, controlHover               rgb
}

func controlStates(p palette, c map[string]rgb, isDark bool) states {
	black := rgb{0, 0, 0}
	darken := func(fill rgb, n float64) rgb { return black.over(fill, p.Tint.Fill*n) }
	s := states{
		accentHover: darken(c["accent"], 1), accentActive: darken(c["accent"], 2),
		destructiveHover: darken(c["destructive"], 1), destructiveActive: darken(c["destructive"], 2),
		// A secondary button is a step off the page: raised in dark, the surface in light, where
		// raised read as a disabled grey. Under the pointer, the next step.
		control: c["surface"], controlHover: c["raised"],
	}
	if isDark {
		s.control, s.controlHover = c["raised"], c["overlay"]
	}
	return s
}
