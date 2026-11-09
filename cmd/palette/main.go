// Command palette checks palette.toml - every pair of colors that carries meaning, for contrast
// and for perceptual distance as drawn, on real panels and through each deficiency of colour
// vision - and writes it out in dist/ for each place that draws Spinbox. With -check it writes
// nothing and fails when dist/ is not what palette.toml makes, or when a pair is below its floor.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/BurntSushi/toml"
)

// mode is one theme's roles, each a color.
type mode struct {
	Page, Surface, Raised, Overlay string
	Border                         string
	BorderStrong                   string `toml:"border_strong"`
	Text, Muted                    string
	Accent                         string
	AccentText                     string `toml:"accent_text"`
	OnAccent                       string `toml:"on_accent"`
	Label                          string
	Success, Warning, Error        string
	Destructive                    string
}

type terminal struct {
	Background, Foreground, Cursor, Selection             string
	Black, Red, Green, Yellow, Blue, Magenta, Cyan, White string
	BrightBlack                                           string `toml:"bright_black"`
	BrightRed                                             string `toml:"bright_red"`
	BrightGreen                                           string `toml:"bright_green"`
	BrightYellow                                          string `toml:"bright_yellow"`
	BrightBlue                                            string `toml:"bright_blue"`
	BrightMagenta                                         string `toml:"bright_magenta"`
	BrightCyan                                            string `toml:"bright_cyan"`
	BrightWhite                                           string `toml:"bright_white"`
}

// ansi is the sixteen colours by palette.toml's names, in the order a program numbers them.
var ansi = []string{
	"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white",
	"bright_black", "bright_red", "bright_green", "bright_yellow", "bright_blue", "bright_magenta", "bright_cyan", "bright_white",
}

// terminalRoles is the terminal's roles before the sixteen colours, in the order they are written.
var terminalRoles = []string{"background", "foreground", "cursor", "selection"}

// colors reads the terminal's roles and colours by palette.toml's names.
func (t terminal) colors() (map[string]rgb, error) {
	named := map[string]string{
		"background": t.Background, "foreground": t.Foreground, "cursor": t.Cursor, "selection": t.Selection,
		"black": t.Black, "red": t.Red, "green": t.Green, "yellow": t.Yellow,
		"blue": t.Blue, "magenta": t.Magenta, "cyan": t.Cyan, "white": t.White,
		"bright_black": t.BrightBlack, "bright_red": t.BrightRed, "bright_green": t.BrightGreen,
		"bright_yellow": t.BrightYellow, "bright_blue": t.BrightBlue, "bright_magenta": t.BrightMagenta,
		"bright_cyan": t.BrightCyan, "bright_white": t.BrightWhite,
	}
	return parseNamed(named)
}

type palette struct {
	Brand map[string]string
	Dark  mode
	Light mode
	// A terminal's roles and its sixteen ANSI colours, the same in either theme.
	Terminal terminal
	Tint     struct{ Hover, Selected, Scrim, Fill, Badge float64 }
	// What lifts a menu or a dialog off what is under it, per theme: a CSS box-shadow.
	Elevation struct{ Dark, Light string }
	Type      struct {
		Sans, Mono      string
		FallbackSans    string            `toml:"fallback_sans"`
		FallbackMono    string            `toml:"fallback_mono"`
		ReadingSize     int               `toml:"reading_size"`
		DenseSize       int               `toml:"dense_size"`
		DenseScale      map[string][2]int `toml:"dense_scale"`
		LineHeight      float64           `toml:"line_height"`
		HeadingWeight   int               `toml:"heading_weight"`
		HeadingTracking float64           `toml:"heading_tracking"`
		Heading         string
		Display         string
		DisplayTracking float64 `toml:"display_tracking"`
		Fluid           map[string][2]int
		Numeric         string
		Wordmark        string
	}
	Layout struct {
		Measure, Sidebar       int
		RadiusSm               int `toml:"radius_sm"`
		Radius                 int
		RadiusLg               int `toml:"radius_lg"`
		Target                 int
		TargetTouch            int `toml:"target_touch"`
		Phone, Tablet, Desktop int
		WidthPage              int `toml:"width_page"`
		WidthPanelMin          int `toml:"width_panel_min"`
	}
	// A control's geometry: the heights of a button or an input, the icon in it, and the gap
	// between an icon and its word.
	Controls struct {
		Sm, Md, Lg, Icon, Gap int
	}
	Motion struct {
		Fast, Base, Slow, Draw int
		Ease                   string
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "palette:", err)
		os.Exit(1)
	}
}

func run() error {
	src := flag.String("palette", "palette.toml", "the palette")
	out := flag.String("out", "dist", "where the palette is written")
	check := flag.Bool("check", false, "write nothing; fail if dist/ is not what the palette makes")
	report := flag.Bool("report", false, "print every pair with its contrast and distance in each view")
	flag.Parse()

	var p palette
	md, err := toml.DecodeFile(*src, &p)
	if err != nil {
		return err
	}
	if undecoded := md.Undecoded(); len(undecoded) > 0 {
		return fmt.Errorf("%s: keys nothing reads: %v", *src, undecoded)
	}

	results, err := holds(p)
	if err != nil {
		return err
	}
	failed := 0
	for _, r := range results {
		if !r.ok || *report {
			fmt.Println(r)
		}
		if !r.ok {
			failed++
		}
	}

	files, err := outputs(p, results)
	if err != nil {
		return err
	}
	stale, err := sync(*out, files, *check)
	if err != nil {
		return err
	}
	switch {
	case failed > 0:
		return fmt.Errorf("%d pairs below their floor", failed)
	case *check && len(stale) > 0:
		return fmt.Errorf("dist/ is not what %s makes, run `task palette`: %v", *src, stale)
	}
	if *check {
		fmt.Printf("palette: %d pairs hold, dist/ is what %s makes\n", len(results), *src)
		return nil
	}
	fmt.Printf("palette: %d pairs hold, dist/ written (%d files)\n", len(results), len(files))
	return nil
}

// sync makes out hold exactly files, and says which differed.
func sync(out string, files map[string][]byte, check bool) ([]string, error) {
	var stale []string
	have, _ := filepath.Glob(filepath.Join(out, "*"))
	for _, h := range have {
		if _, ok := files[filepath.Base(h)]; !ok {
			stale = append(stale, filepath.Base(h))
			if !check {
				if err := os.Remove(h); err != nil {
					return nil, err
				}
			}
		}
	}
	names := make([]string, 0, len(files))
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		path := filepath.Join(out, n)
		old, err := os.ReadFile(path)
		if err == nil && bytes.Equal(old, files[n]) {
			continue
		}
		stale = append(stale, n)
		if check {
			continue
		}
		if err := os.MkdirAll(out, 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, files[n], 0o644); err != nil {
			return nil, err
		}
	}
	return stale, nil
}
