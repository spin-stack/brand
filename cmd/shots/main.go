// Command shots keeps the app's screenshots beside the palette that draws it, so the brand, the
// website and anything else show the dashboard as it is, from one place. spin's own dashboard
// suite takes them (dashboard/e2e/site.spec.ts, of its end-to-end stack - real workspaces on a
// real runner - each test first checking the screen shows what its title says); `task
// screenshots` copies them into shots/ with the release and the commit of the spin tree that
// took them, and the title of the test that took each. With -check it fails on a shot
// shots/shots.json lists and shots/ does not hold.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const spec = "dashboard/e2e/site.spec.ts"

type record struct {
	Release string `json:"release"`
	Commit  string `json:"commit"`
	Shots   []shot `json:"shots"`
}

type shot struct {
	Name string `json:"name"`
	Test string `json:"test"`
}

// specTest is a test of the spec and the name its shoot() call writes.
var specTest = regexp.MustCompile(`(?s)test\("([^"]+)",.*?shoot\(\w+, "([\w-]+)"\)`)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "shots:", err)
		os.Exit(1)
	}
}

func run() error {
	spin := flag.String("spin", "../spin", "the spin checkout whose dashboard suite took the shots")
	from := flag.String("from", "", "where the suite wrote them: SITE_SHOTS=<dir> task e2e:dashboard -- site.spec.ts")
	out := flag.String("out", "shots", "where they are kept")
	check := flag.Bool("check", false, "write nothing; fail on a shot shots.json lists and shots/ does not hold")
	flag.Parse()

	index := filepath.Join(*out, "shots.json")
	if *check {
		raw, err := os.ReadFile(index)
		if err != nil {
			return err
		}
		var r record
		if err := json.Unmarshal(raw, &r); err != nil {
			return fmt.Errorf("%s: %w", index, err)
		}
		for _, s := range r.Shots {
			if _, err := os.Stat(filepath.Join(*out, s.Name+".png")); err != nil {
				return fmt.Errorf("%s lists %s, and %s has no %s.png: run `task screenshots`", index, s.Name, *out, s.Name)
			}
		}
		fmt.Printf("shots: %d screenshots of spin %s\n", len(r.Shots), r.Release)
		return nil
	}
	if *from == "" {
		return fmt.Errorf("-from names where spin's suite wrote the screenshots")
	}
	raw, err := os.ReadFile(filepath.Join(*spin, spec))
	if err != nil {
		return err
	}
	r := record{}
	for _, m := range specTest.FindAllSubmatch(raw, -1) {
		r.Shots = append(r.Shots, shot{Name: string(m[2]), Test: string(m[1])})
	}
	if len(r.Shots) == 0 {
		return fmt.Errorf("%s: no test that shoots", spec)
	}
	if r.Release, err = git(*spin, "describe", "--tags", "--abbrev=0", "--match", "v[0-9]*"); err != nil {
		return err
	}
	if r.Commit, err = git(*spin, "rev-parse", "HEAD"); err != nil {
		return err
	}
	old, _ := filepath.Glob(filepath.Join(*out, "*.png"))
	for _, o := range old {
		if err := os.Remove(o); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		return err
	}
	for _, s := range r.Shots {
		png, err := os.ReadFile(filepath.Join(*from, s.Name+".png"))
		if err != nil {
			return fmt.Errorf("%w: run `SITE_SHOTS=%s task e2e:dashboard -- site.spec.ts` in spin first", err, *from)
		}
		if err := os.WriteFile(filepath.Join(*out, s.Name+".png"), png, 0o644); err != nil {
			return err
		}
	}
	b, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(index, append(b, '\n'), 0o644)
}

func git(dir string, args ...string) (string, error) {
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).Output()
	if err != nil {
		return "", fmt.Errorf("git %s in %s: %w", strings.Join(args, " "), dir, err)
	}
	return strings.TrimSpace(string(out)), nil
}
