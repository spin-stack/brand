package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// rgb is a color in sRGB, each channel 0..1, as CSS writes it.
type rgb struct{ r, g, b float64 }

func parseHex(s string) (rgb, error) {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return rgb{}, fmt.Errorf("%q is not #RRGGBB", s)
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return rgb{}, fmt.Errorf("%q is not #RRGGBB: %w", s, err)
	}
	return rgb{float64(v>>16&0xff) / 255, float64(v>>8&0xff) / 255, float64(v&0xff) / 255}, nil
}

func (c rgb) hex() string {
	b := func(v float64) int { return int(math.Round(math.Max(0, math.Min(1, v)) * 255)) }
	return fmt.Sprintf("#%02X%02X%02X", b(c.r), b(c.g), b(c.b))
}

// over is c at alpha over bg, blended as a browser does: in sRGB, gamma and all.
func (c rgb) over(bg rgb, alpha float64) rgb {
	return rgb{bg.r + (c.r-bg.r)*alpha, bg.g + (c.g-bg.g)*alpha, bg.b + (c.b-bg.b)*alpha}
}

func toLinear(v float64) float64 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

func fromLinear(v float64) float64 {
	v = math.Max(0, math.Min(1, v))
	if v <= 0.0031308 {
		return v * 12.92
	}
	return 1.055*math.Pow(v, 1/2.4) - 0.055
}

// linear is a color in linear RGB, where light adds up.
type linear struct{ r, g, b float64 }

func (c rgb) linear() linear { return linear{toLinear(c.r), toLinear(c.g), toLinear(c.b)} }
func (l linear) srgb() rgb   { return rgb{fromLinear(l.r), fromLinear(l.g), fromLinear(l.b)} }

// luminance is WCAG 2's relative luminance.
func (l linear) luminance() float64 { return 0.2126*l.r + 0.7152*l.g + 0.0722*l.b }

// contrast is WCAG 2's contrast ratio, 1 to 21.
func contrast(a, b rgb) float64 {
	la, lb := a.linear().luminance(), b.linear().luminance()
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// lab is CIELAB under D65, from linear RGB through XYZ.
type lab struct{ l, a, b float64 }

func (l linear) lab() lab {
	x := 0.4124564*l.r + 0.3575761*l.g + 0.1804375*l.b
	y := 0.2126729*l.r + 0.7151522*l.g + 0.0721750*l.b
	z := 0.0193339*l.r + 0.1191920*l.g + 0.9503041*l.b
	f := func(t float64) float64 {
		if t > 216.0/24389 {
			return math.Cbrt(t)
		}
		return (24389.0/27*t + 16) / 116
	}
	fx, fy, fz := f(x/0.95047), f(y/1.0), f(z/1.08883)
	return lab{116*fy - 16, 500 * (fx - fy), 200 * (fy - fz)}
}

// deltaE00 is CIEDE2000 (Sharma, Wu and Dalal 2005), with kL = kC = kH = 1.
func deltaE00(p, q lab) float64 {
	deg := math.Pi / 180
	c1, c2 := math.Hypot(p.a, p.b), math.Hypot(q.a, q.b)
	cbar := (c1 + c2) / 2
	g := 0.5 * (1 - math.Sqrt(math.Pow(cbar, 7)/(math.Pow(cbar, 7)+math.Pow(25, 7))))
	a1, a2 := (1+g)*p.a, (1+g)*q.a
	c1p, c2p := math.Hypot(a1, p.b), math.Hypot(a2, q.b)
	hue := func(b, a float64) float64 {
		if a == 0 && b == 0 {
			return 0
		}
		h := math.Atan2(b, a) / deg
		if h < 0 {
			h += 360
		}
		return h
	}
	h1, h2 := hue(p.b, a1), hue(q.b, a2)
	dL, dC := q.l-p.l, c2p-c1p
	var dh float64
	switch {
	case c1p*c2p == 0:
		dh = 0
	case math.Abs(h2-h1) <= 180:
		dh = h2 - h1
	case h2-h1 > 180:
		dh = h2 - h1 - 360
	default:
		dh = h2 - h1 + 360
	}
	dH := 2 * math.Sqrt(c1p*c2p) * math.Sin(dh/2*deg)
	lbar, cbarp := (p.l+q.l)/2, (c1p+c2p)/2
	var hbar float64
	switch {
	case c1p*c2p == 0:
		hbar = h1 + h2
	case math.Abs(h1-h2) <= 180:
		hbar = (h1 + h2) / 2
	case h1+h2 < 360:
		hbar = (h1 + h2 + 360) / 2
	default:
		hbar = (h1 + h2 - 360) / 2
	}
	t := 1 - 0.17*math.Cos((hbar-30)*deg) + 0.24*math.Cos(2*hbar*deg) +
		0.32*math.Cos((3*hbar+6)*deg) - 0.20*math.Cos((4*hbar-63)*deg)
	dTheta := 30 * math.Exp(-math.Pow((hbar-275)/25, 2))
	rc := 2 * math.Sqrt(math.Pow(cbarp, 7)/(math.Pow(cbarp, 7)+math.Pow(25, 7)))
	sl := 1 + 0.015*math.Pow(lbar-50, 2)/math.Sqrt(20+math.Pow(lbar-50, 2))
	sc := 1 + 0.045*cbarp
	sh := 1 + 0.015*cbarp*t
	rt := -math.Sin(2*dTheta*deg) * rc
	return math.Sqrt(math.Pow(dL/sl, 2) + math.Pow(dC/sc, 2) + math.Pow(dH/sh, 2) + rt*(dC/sc)*(dH/sh))
}

// A view is how a color reaches an eye: through a deficiency of colour vision, a panel in a lit
// room, or both. Each maps linear RGB to linear RGB, and the eye is taken as adapted to the
// view's white, so a dim panel in a lit room does worse than the same panel in the dark.
type view struct {
	name string
	fn   func(linear) linear
}

// Machado, Oliveira and Fernandes 2009, severity 1.0: dichromacy, the worst case. In linear RGB.
var (
	protan = [3][3]float64{{0.152286, 1.052583, -0.204868}, {0.114503, 0.786281, 0.099216}, {-0.003882, -0.048116, 1.051998}}
	deutan = [3][3]float64{{0.367322, 0.860646, -0.227968}, {0.280085, 0.672501, 0.047413}, {-0.011820, 0.042940, 0.968881}}
	tritan = [3][3]float64{{1.255528, -0.076749, -0.178779}, {-0.078411, 0.930809, 0.147602}, {0.004733, 0.691367, 0.303900}}
)

func matrix(m [3][3]float64) func(linear) linear {
	clamp := func(v float64) float64 { return math.Max(0, math.Min(1, v)) }
	return func(c linear) linear {
		return linear{
			clamp(m[0][0]*c.r + m[0][1]*c.g + m[0][2]*c.b),
			clamp(m[1][0]*c.r + m[1][1]*c.g + m[1][2]*c.b),
			clamp(m[2][0]*c.r + m[2][1]*c.g + m[2][2]*c.b),
		}
	}
}

// panel is a display in a room: its white and black in cd/m2 and the room's light it reflects,
// as the eye adapted to its white sees it.
func panel(white, black, flare float64) func(linear) linear {
	ch := func(v float64) float64 { return (black + (white-black)*v + flare) / (white + flare) }
	return func(c linear) linear { return linear{ch(c.r), ch(c.g), ch(c.b)} }
}

func then(f, g func(linear) linear) func(linear) linear {
	return func(c linear) linear { return g(f(c)) }
}

// views are the ways a pair is held apart: as drawn, on an office's IPS panel (250 cd/m2 white,
// 0.3 black, 0.6 of flare from 200 lx at 1%), on the same panel at 30% backlight, and through
// each dichromacy, protan and deutan also on the office panel - 8% of men between them.
var views = func() []view {
	ips := panel(250, 0.3, 0.6)
	dim := panel(75, 0.09, 0.6)
	return []view{
		{"drawn", func(c linear) linear { return c }},
		{"ips", ips},
		{"dim", dim},
		{"protan", matrix(protan)},
		{"deutan", matrix(deutan)},
		{"tritan", matrix(tritan)},
		{"protan+ips", then(matrix(protan), ips)},
		{"deutan+ips", then(matrix(deutan), ips)},
	}
}()
