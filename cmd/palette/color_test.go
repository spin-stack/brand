package main

import (
	"math"
	"testing"
)

// Sharma, Wu and Dalal's test data: a CIEDE2000 that gets these right gets the hue term, the
// rotation term and the wrap at 360 right.
func TestDeltaE00IsSharmas(t *testing.T) {
	for _, c := range []struct {
		p, q lab
		want float64
	}{
		{lab{50, 2.6772, -79.7751}, lab{50, 0, -82.7485}, 2.0425},
		{lab{50, -1, 2}, lab{50, 0, 0}, 2.3669},
		{lab{50, 2.5, 0}, lab{73, 25, -18}, 27.1492},
		{lab{2.0776, 0.0795, -1.1350}, lab{0.9033, -0.0636, -0.5514}, 0.9082},
		{lab{90.8027, -2.0831, 1.4410}, lab{91.1528, -1.6435, 0.0447}, 1.4441},
	} {
		if got := deltaE00(c.p, c.q); math.Abs(got-c.want) > 1e-4 {
			t.Errorf("deltaE00(%v, %v) = %.4f, want %.4f", c.p, c.q, got, c.want)
		}
	}
}

// WCAG's own extremes: black on white is 21, a color on itself is 1.
func TestContrastIsWCAGs(t *testing.T) {
	black, white := rgb{0, 0, 0}, rgb{1, 1, 1}
	if got := contrast(black, white); math.Abs(got-21) > 1e-9 {
		t.Errorf("black on white = %v, want 21", got)
	}
	if got := contrast(white, white); got != 1 {
		t.Errorf("white on white = %v, want 1", got)
	}
	// #767676 on white is the classic 4.54:1.
	grey, _ := parseHex("#767676")
	if got := contrast(grey, white); math.Abs(got-4.54) > 0.01 {
		t.Errorf("#767676 on white = %.3f, want 4.54", got)
	}
}
