// Command screens keeps every screen and dialog of the app, in each theme and at each width, beside
// the palette that draws them, with what spin's suite found off the palette on each. spin's
// dashboard suite takes them (dashboard/e2e/catalog.spec.ts walks dashboard/e2e/catalog.ts, which a
// test holds to every route and dialog the app has) and writes, into one directory:
//
//   - catalog.json: each screen - its name, title, area, route and whether it is a page or a dialog -
//     and each route or dialog it leaves out, with the reason;
//   - <name>.audit.json: each value a screen draws that brand does not have, and each rule it
//     breaks;
//   - <name>.<theme>.<width>.png for every screen, theme and width.
//
// The two JSON files are committed in screens/, so a change in what the app is shows as a diff. The
// images are not: there are hundreds a release, and git would keep every one forever. They go up as
// one archive on a release of this repository, which screens.json names and Pages unpacks.
// With -check it fails on a catalog that is not whole: a screen without each of its images in the
// archive's list, or an audit finding about a screen the catalog does not have.
package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

var (
	themes = []string{"light", "dark"}
	widths = []string{"desktop", "mobile"}
)

// Archive is the asset's name on each release.
const archive = "screens.tar.gz"

type screen struct {
	Name  string `json:"name"`
	Title string `json:"title"`
	Area  string `json:"area"`
	Route string `json:"route"`
	Kind  string `json:"kind"`
}

type exemption struct {
	What   string `json:"what"`
	Reason string `json:"reason"`
}

type catalog struct {
	Screens []screen    `json:"screens"`
	Exempt  []exemption `json:"exempt"`
}

type finding struct {
	Screen   string `json:"screen"`
	Theme    string `json:"theme"`
	Width    string `json:"width"`
	Kind     string `json:"kind"`
	Value    string `json:"value"`
	Selector string `json:"selector"`
	Count    int    `json:"count"`
}

type audit struct {
	Findings []finding `json:"findings"`
}

// index is screens/screens.json: the catalog with what took it and where its images are.
type index struct {
	Release string      `json:"release"`
	Commit  string      `json:"commit"`
	Tag     string      `json:"tag"`
	Images  []string    `json:"images"`
	Screens []screen    `json:"screens"`
	Exempt  []exemption `json:"exempt"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "screens:", err)
		os.Exit(1)
	}
}

func run() error {
	spin := flag.String("spin", "../spin", "the spin checkout whose dashboard suite took the screens")
	from := flag.String("from", "", "where the suite wrote them: CATALOG_SHOTS=<dir> task e2e:dashboard -- catalog.spec.ts")
	out := flag.String("out", "screens", "where the catalog and the audit are kept")
	repo := flag.String("repo", "spin-stack/brand", "the repository whose release holds the images")
	check := flag.Bool("check", false, "write nothing; fail on a catalog that is not whole")
	flag.Parse()

	if *check {
		ix, a, err := readKept(*out)
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("screens: none taken yet")
			return nil
		}
		if err != nil {
			return err
		}
		if err := whole(ix.Screens, ix.Images, a); err != nil {
			return err
		}
		fmt.Printf("screens: %d screens of spin %s, %d findings, images on %s\n", len(ix.Screens), ix.Release, len(a.Findings), ix.Tag)
		return nil
	}
	if *from == "" {
		return errors.New("-from names where spin's suite wrote the screens")
	}

	var c catalog
	if err := readJSON(filepath.Join(*from, "catalog.json"), &c); err != nil {
		return err
	}
	// One file a screen: a test that is retried starts a new worker, which would lose what an
	// earlier test had gathered in memory. An audit with nothing found is an empty list, not null:
	// screens.html reads it as an array.
	a := audit{Findings: []finding{}}
	audits, err := filepath.Glob(filepath.Join(*from, "*.audit.json"))
	if err != nil {
		return err
	}
	for _, f := range audits {
		var found []finding
		if err := readJSON(f, &found); err != nil {
			return err
		}
		a.Findings = append(a.Findings, found...)
	}
	taken, err := filepath.Glob(filepath.Join(*from, "*.png"))
	if err != nil {
		return err
	}
	images := make([]string, 0, len(taken))
	for _, t := range taken {
		images = append(images, filepath.Base(t))
	}
	slices.Sort(images)
	if err := whole(c.Screens, images, a); err != nil {
		return err
	}

	ix := index{Images: images, Screens: c.Screens, Exempt: c.Exempt}
	if ix.Release, err = git(*spin, "describe", "--tags", "--abbrev=0", "--match", "v[0-9]*"); err != nil {
		return err
	}
	if ix.Commit, err = git(*spin, "rev-parse", "HEAD"); err != nil {
		return err
	}
	ix.Tag = fmt.Sprintf("screens-%s-%s", ix.Release, ix.Commit[:12])

	tmp, err := os.MkdirTemp("", "screens")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	tgz := filepath.Join(tmp, archive)
	if err := pack(tgz, *from, images); err != nil {
		return err
	}
	notes := fmt.Sprintf("Every screen of spin %s (%s), in each theme and at each width: %d images. screens/screens.json names this release.", ix.Release, ix.Commit, len(images))
	gh := exec.Command("gh", "release", "create", ix.Tag, tgz, "--repo", *repo, "--title", ix.Tag, "--notes", notes, "--latest=false")
	gh.Stdout, gh.Stderr = os.Stdout, os.Stderr
	if err := gh.Run(); err != nil {
		return fmt.Errorf("gh release create %s: %w", ix.Tag, err)
	}

	if err := os.MkdirAll(*out, 0o755); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(*out, "screens.json"), ix); err != nil {
		return err
	}
	return writeJSON(filepath.Join(*out, "audit.json"), a)
}

// whole is the catalog's promise: each screen in each theme at each width, nothing else, and
// every finding about one of them.
func whole(screens []screen, images []string, a audit) error {
	var errs []error
	want := map[string]bool{}
	names := map[string]bool{}
	for _, s := range screens {
		if names[s.Name] {
			errs = append(errs, fmt.Errorf("two screens are named %s", s.Name))
		}
		names[s.Name] = true
		if s.Kind != "page" && s.Kind != "dialog" {
			errs = append(errs, fmt.Errorf("%s is a %q, not a page or a dialog", s.Name, s.Kind))
		}
		for _, t := range themes {
			for _, w := range widths {
				want[fmt.Sprintf("%s.%s.%s.png", s.Name, t, w)] = true
			}
		}
	}
	have := map[string]bool{}
	for _, i := range images {
		have[i] = true
		if !want[i] {
			errs = append(errs, fmt.Errorf("%s is no screen of the catalog in a theme and a width", i))
		}
	}
	for w := range want {
		if !have[w] {
			errs = append(errs, fmt.Errorf("%s was not taken", w))
		}
	}
	if a.Findings == nil {
		errs = append(errs, errors.New("the audit's findings are null, not a list: screens.html reads an array"))
	}
	for _, f := range a.Findings {
		if !names[f.Screen] || !slices.Contains(themes, f.Theme) || !slices.Contains(widths, f.Width) {
			errs = append(errs, fmt.Errorf("a finding about %s.%s.%s, which the catalog does not have", f.Screen, f.Theme, f.Width))
		}
	}
	slices.SortFunc(errs, func(a, b error) int { return strings.Compare(a.Error(), b.Error()) })
	return errors.Join(errs...)
}

func readKept(dir string) (index, audit, error) {
	var ix index
	var a audit
	if err := readJSON(filepath.Join(dir, "screens.json"), &ix); err != nil {
		return ix, a, err
	}
	return ix, a, readJSON(filepath.Join(dir, "audit.json"), &a)
}

func pack(path, dir string, images []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for _, name := range images {
		src, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		st, err := src.Stat()
		if err != nil {
			src.Close()
			return err
		}
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: st.Size()}); err != nil {
			src.Close()
			return err
		}
		_, err = io.Copy(tw, src)
		src.Close()
		if err != nil {
			return err
		}
	}
	if err := tw.Close(); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}
	return f.Close()
}

func readJSON(path string, v any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

func git(dir string, args ...string) (string, error) {
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).Output()
	if err != nil {
		return "", fmt.Errorf("git %s in %s: %w", strings.Join(args, " "), dir, err)
	}
	return strings.TrimSpace(string(out)), nil
}
