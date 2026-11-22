package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

const root = "../.."

// shipped is the files fonts/ has to hold for a face to be one a consumer can serve: the latin
// subset at each weight the palette gives a rank, named as @fontsource names them, and the
// face's licence.
func shipped(p palette, face string) []string {
	slug := strings.ReplaceAll(strings.ToLower(face), " ", "-")
	files := []string{"LICENSE-" + strings.ReplaceAll(face, " ", "-") + ".txt"}
	for _, w := range []int{p.Type.WeightRegular, p.Type.WeightMedium, p.Type.WeightStrong, p.Type.HeadingWeight} {
		if f := fmt.Sprintf("%s-latin-%d-normal.woff2", slug, w); !slices.Contains(files, f) {
			files = append(files, f)
		}
	}
	return files
}

// A face an output names and fonts/ does not ship is drawn in the reader's fallback with nothing
// failing: the page loads, in another face. So every font stack dist/ writes leads with a face
// whose files are in fonts/, at every weight the palette ranks, and fonts/ holds no other face.
func TestEveryFaceAnOutputNamesIsShipped(t *testing.T) {
	var p palette
	if _, err := toml.DecodeFile(filepath.Join(root, "palette.toml"), &p); err != nil {
		t.Fatal(err)
	}
	results, err := holds(p)
	if err != nil {
		t.Fatal(err)
	}
	files, err := outputs(p, results)
	if err != nil {
		t.Fatal(err)
	}

	named := map[string][]string{}
	stack := regexp.MustCompile(`--[a-z-]*font[a-z-]*:\s*([^;]+);`)
	for _, name := range []string{"tokens.css", "starlight.css", "dashboard.css", "t3code.css"} {
		faces := 0
		for _, m := range stack.FindAllStringSubmatch(string(files[name]), -1) {
			if strings.HasPrefix(m[1], "var(") {
				continue
			}
			first, _, _ := strings.Cut(m[1], ",")
			face := strings.Trim(strings.TrimSpace(first), `'"`)
			named[face] = append(named[face], name)
			faces++
		}
		if faces == 0 {
			t.Fatalf("%s names no face: the test no longer reads how it writes them", name)
		}
	}
	var tokens struct {
		Type struct{ Sans, Mono string } `json:"type"`
	}
	if err := json.Unmarshal(files["tokens.json"], &tokens); err != nil {
		t.Fatal(err)
	}
	if tokens.Type.Sans == "" || tokens.Type.Mono == "" {
		t.Fatal("tokens.json's type names no sans or no mono")
	}
	named[tokens.Type.Sans] = append(named[tokens.Type.Sans], "tokens.json")
	named[tokens.Type.Mono] = append(named[tokens.Type.Mono], "tokens.json")

	want := map[string]bool{}
	for face, by := range named {
		for _, f := range shipped(p, face) {
			want[f] = true
			if _, err := os.Stat(filepath.Join(root, "fonts", f)); err != nil {
				t.Errorf("%v name the face %q, and fonts/ has no %s", by, face, f)
			}
		}
	}
	have, err := os.ReadDir(filepath.Join(root, "fonts"))
	if err != nil {
		t.Fatal(err)
	}
	for _, h := range have {
		if !want[h.Name()] {
			t.Errorf("fonts/%s is of no face at a weight the outputs name", h.Name())
		}
	}
}

// The pages this repository publishes draw the palette's faces from fonts/: an @font-face whose
// family is not the palette's, or whose file is not there, leaves that page in a fallback.
func TestThePagesLoadThePalettesFaces(t *testing.T) {
	var p palette
	if _, err := toml.DecodeFile(filepath.Join(root, "palette.toml"), &p); err != nil {
		t.Fatal(err)
	}
	rule := regexp.MustCompile(`@font-face \{ font-family: '([^']+)'; src: url\(([^)]+)\)[^}]*font-weight: (\d+); \}`)
	for _, page := range []string{"index.html", "screens.html", "five-seconds.html", "recognition.html", "ui-hill-climb/app.html"} {
		raw, err := os.ReadFile(filepath.Join(root, page))
		if err != nil {
			t.Fatal(err)
		}
		loaded := map[string]bool{}
		for _, m := range rule.FindAllStringSubmatch(string(raw), -1) {
			family, src, weight := m[1], m[2], m[3]
			loaded[family] = true
			if family != p.Type.Sans && family != p.Type.Mono {
				t.Errorf("%s loads %q, which is neither of the palette's faces", page, family)
			}
			if _, err := os.Stat(filepath.Join(root, filepath.Dir(page), src)); err != nil {
				t.Errorf("%s loads %s, which is not there", page, src)
			}
			if want := fmt.Sprintf("%s-latin-%s-normal.woff2", strings.ReplaceAll(strings.ToLower(family), " ", "-"), weight); filepath.Base(src) != want {
				t.Errorf("%s gives %q at %s the file %s, not %s", page, family, weight, filepath.Base(src), want)
			}
		}
		for _, face := range []string{p.Type.Sans, p.Type.Mono} {
			if !loaded[face] {
				t.Errorf("%s has no @font-face for %q", page, face)
			}
		}
	}
}
