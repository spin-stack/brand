package main

// The fills a button takes under the pointer and when pressed, derived from the roles so that a
// consumer does not invent them (the dashboard wrote primary/90: the accent at 90% over the page,
// lighter in dark and darker in light, and its white label under 4.5:1 in one of them).
//
// A solid fill - the accent, the destructive - moves away from its own lightness, by tint.fill
// under the pointer and twice it when pressed: black over it, or white over a fill that is already
// ink. The fill decides and not the theme: a theme's action can be the mark's navy, where black
// over it is a step nobody sees, and the same theme's destructive a mid red, where white over it
// takes its white label under 4.5:1. A quiet fill - a secondary button's - steps one surface up.
//
// A list's row under the pointer or the keyboard is a plane, and an action in it under the
// pointer the next one, so that neither is drawn with a line.
type states struct {
	accentHover, accentActive           rgb
	destructiveHover, destructiveActive rgb
	control, controlHover               rgb
	outline                             rgb
	row, rowAction                      rgb
}

// inkFill is the CIELAB lightness under which a solid fill lightens. Measured at tint.fill 0.12, in
// grey, navy, red and cyan: from L* 32 up, black over a fill is told from it in every view (the
// hover tier), and under L* 32 a white label holds 4.5:1 on white over the fill, pressed. A mid
// fill is far above it (the reds and blues a white label reads on are L* 44 to 50) and the mark's
// navy far below (L* 8).
const inkFill = 32

func controlStates(p palette, c map[string]rgb, isDark bool) states {
	step := func(fill rgb, n float64) rgb {
		toward := rgb{0, 0, 0}
		if fill.linear().lab().l < inkFill {
			toward = rgb{1, 1, 1}
		}
		return toward.over(fill, p.Tint.Fill*n)
	}
	s := states{
		accentHover: step(c["accent"], 1), accentActive: step(c["accent"], 2),
		destructiveHover: step(c["destructive"], 1), destructiveActive: step(c["destructive"], 2),
		// A secondary button is a step off the page: raised in dark, the surface in light, where
		// raised read as a disabled grey. Under the pointer, the next step.
		control: c["surface"], controlHover: c["raised"],
		// A secondary button's outline is the border, never the input's edge: a button has its
		// word to say what it is, and an empty input only its edge (palette.toml's [controls]).
		outline: c["border"],
		// The first step off the page in either theme, the quietest plane the hover tier holds.
		row: c["surface"], rowAction: c["raised"],
	}
	if isDark {
		s.control, s.controlHover = c["raised"], c["overlay"]
	}
	return s
}
