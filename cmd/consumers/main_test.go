package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// repo is a git repository of a test's own: a brand, or a consumer's checkout.
type repo struct {
	t   *testing.T
	dir string
}

func newRepo(t *testing.T) repo {
	t.Helper()
	// A test run from a git hook inherits the hook's repository: these are asked only of theirs.
	for _, v := range []string{"GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE"} {
		if old, ok := os.LookupEnv(v); ok {
			os.Unsetenv(v)
			t.Cleanup(func() { os.Setenv(v, old) })
		}
	}
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
	r := repo{t, t.TempDir()}
	r.git("init", "-q", "-b", "main")
	return r
}

func (r repo) git(args ...string) string {
	r.t.Helper()
	args = append([]string{"-C", r.dir, "-c", "user.name=test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false", "-c", "tag.gpgsign=false"}, args...)
	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// commit writes files and commits them, and answers the commit.
func (r repo) commit(files map[string]string) string {
	r.t.Helper()
	for name, content := range files {
		path := filepath.Join(r.dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			r.t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			r.t.Fatal(err)
		}
	}
	r.git("add", ".")
	r.git("commit", "-q", "--allow-empty", "-m", "a change")
	return r.git("rev-parse", "HEAD")
}

const faces = `- name: inter
  source: https://registry.npmjs.org/@fontsource/inter/-/inter-{version}.tgz
  version: 5.3.0
- name: jetbrains-mono
  source: https://registry.npmjs.org/@fontsource/jetbrains-mono/-/jetbrains-mono-{version}.tgz
  version: 5.3.0
- name: task
  source: https://github.com/go-task/task/releases/download/v{version}/task.tar.gz
  version: 3.54.0
`

const brandsFaces = `{"dependencies": {"@fontsource/inter": "5.3.0", "astro": "6.0.0"}, "devDependencies": {"@fontsource/jetbrains-mono": "5.3.0"}}`

func pin(commit, version, track string) string {
	return "- name: task\n  version: 3.54.0\n- name: brand\n  kind: git\n  version: " + version + "\n  pin: " + commit + "\n  track: " + track + "\n"
}

// TestEachConsumerIsReportedAsWhatItIs reads consumers from checkouts against a brand with two
// releases, the first two commits behind main.
func TestEachConsumerIsReportedAsWhatItIs(t *testing.T) {
	b := newRepo(t)
	first := b.commit(map[string]string{"dist/tokens.css": "old\n", "versions.yaml": faces})
	b.git("tag", "v20261101.01")
	b.git("tag", "screens-v20261019.02-c2f0602021cc")
	b.commit(map[string]string{"dist/tokens.css": "new\n"})
	head := b.commit(nil)
	b.git("tag", "v20261102.01")

	consumers := map[string]map[string]string{
		"current": {"versions.yaml": pin(head, "v20261102.01", "github-release"), "brand.css": "new\n", "package.json": brandsFaces},
		"behind":  {"versions.yaml": pin(first, "v20261101.01", "github-release"), "brand.css": "old\n", "package.json": brandsFaces},
		"edited":  {"versions.yaml": pin(head, "main", "branch"), "brand.css": "new\nand a line by hand\n", "package.json": brandsFaces},
		"faces": {"versions.yaml": pin(head, "main", "branch"), "brand.css": "new\n",
			"package.json": `{"dependencies": {"@fontsource/inter": "^5.3.0", "@fontsource/ibm-plex-mono": "5.3.0"}}`},
		"unpinned": {"brand.css": "a palette of its own\n", "package.json": brandsFaces},
		"copyless": {"versions.yaml": pin(head, "main", "branch"), "package.json": brandsFaces},
	}
	table := ""
	args := []string{"-brand", b.dir, "-main", "main"}
	for _, name := range []string{"current", "behind", "edited", "faces", "unpinned", "copyless", "gone"} {
		table += "[[consumer]]\nname = \"" + name + "\"\nrepo = \"example/" + name + "\"\npin = \"versions.yaml\"\npackages = \"package.json\"\n" +
			"[consumer.copies]\n\"brand.css\" = \"dist/tokens.css\"\n\n"
		dir := filepath.Join(t.TempDir(), "absent")
		if files, ok := consumers[name]; ok {
			c := newRepo(t)
			c.commit(files)
			dir = c.dir
		}
		args = append(args, "-checkout", "example/"+name+"="+dir+"@HEAD")
	}
	path := filepath.Join(t.TempDir(), "consumers.toml")
	if err := os.WriteFile(path, []byte(table), 0o644); err != nil {
		t.Fatal(err)
	}

	var out strings.Builder
	if err := run(append(args, "-table", path), &out); err != nil {
		t.Fatalf("a report failed on what it found: %v", err)
	}
	said := map[string]string{}
	for _, section := range strings.Split(out.String(), "\n\n")[1:] {
		name, _, _ := strings.Cut(section, " ")
		said[name] = section
	}

	cases := []struct {
		consumer string
		says     []string
		not      []string
	}{
		{"current", []string{
			head[:12] + " (v20261102.01, track github-release), the newest release v20261102.01; 0 commits behind main",
			"copies   all 1 are brand's at " + head[:12],
			"fonts    brand's at " + head[:12] + ": @fontsource/inter, @fontsource/jetbrains-mono",
		}, []string{"not brand's"}},
		{"behind", []string{
			"release v20261101.01; the newest is v20261102.01; 2 commits behind main",
			"copies   all 1 are brand's at " + first[:12],
		}, nil},
		{"edited", []string{"copies   1 of 1 are not brand's at " + head[:12], "brand.css is not dist/tokens.css", "fonts    brand's"}, nil},
		{"faces", []string{
			"copies   all 1 are brand's",
			"fonts    not brand's at " + head[:12] + ", which pins 2:",
			"@fontsource/inter is ^5.3.0, and brand's is 5.3.0",
			"@fontsource/jetbrains-mono 5.3.0 is not installed",
			"@fontsource/ibm-plex-mono 5.3.0 is no face brand pins",
		}, nil},
		{"unpinned", []string{"pin      none: it has no versions.yaml", "compared with main", "brand.css is not dist/tokens.css"}, nil},
		{"copyless", []string{"brand.css is not there (dist/tokens.css)"}, nil},
		{"gone", []string{"unread   "}, []string{"pin", "copies", "fonts"}},
	}
	for _, c := range cases {
		for _, want := range c.says {
			if !strings.Contains(said[c.consumer], want) {
				t.Errorf("%s: no %q in:\n%s", c.consumer, want, said[c.consumer])
			}
		}
		for _, not := range c.not {
			if strings.Contains(said[c.consumer], "  "+not) {
				t.Errorf("%s: says %q:\n%s", c.consumer, not, said[c.consumer])
			}
		}
	}
}

// A consumer pinned before versions.yaml named the faces is on a brand that pinned none: what it
// installs is said, and not as a difference from pins made later.
func TestFacesAreHeldToThePinnedCommitNotToMain(t *testing.T) {
	b := newRepo(t)
	before := b.commit(map[string]string{"dist/tokens.css": "old\n", "versions.yaml": "- name: task\n  version: 3.54.0\n"})
	b.commit(map[string]string{"versions.yaml": faces})
	c := newRepo(t)
	c.commit(map[string]string{"versions.yaml": pin(before, "main", "branch"), "package.json": `{"dependencies": {"@fontsource/ibm-plex-sans": "5.3.0"}}`})
	got := strings.Join(report(Consumer{Repo: "example/c", Pin: "versions.yaml", Packages: "package.json"}, checkout{c.dir, "HEAD"}, brand{b.dir, "main"}), "\n")
	if !strings.Contains(got, "brand pinned no face at "+before[:12]) || !strings.Contains(got, "@fontsource/ibm-plex-sans 5.3.0") || strings.Contains(got, "is not installed") {
		t.Fatalf("got:\n%s", got)
	}
}

func TestAPinBrandNeverHadIsSaid(t *testing.T) {
	b := newRepo(t)
	b.commit(map[string]string{"dist/tokens.css": "new\n"})
	c := newRepo(t)
	stray := strings.Repeat("ab", 20)
	c.commit(map[string]string{"versions.yaml": pin(stray, "main", "branch")})
	got := strings.Join(report(Consumer{Repo: "example/c", Pin: "versions.yaml"}, checkout{c.dir, "HEAD"}, brand{b.dir, "main"}), "\n")
	if !strings.Contains(got, stray+" (main), which is no commit of brand") {
		t.Fatalf("got:\n%s", got)
	}
}

// laboratory is what dist/ holds for this repository's own pages, which no consumer copies.
var laboratory = []string{"dist/perception.js", "dist/simulate.svg"}

// TestTheTableIsOfThisRepository holds consumers.toml to the tree: a copy of a file this
// repository does not have is a row nobody can satisfy, and an output nobody copies is either a
// consumer missing from the table or a file made for no one.
func TestTheTableIsOfThisRepository(t *testing.T) {
	consumers, err := load("../../consumers.toml")
	if err != nil {
		t.Fatal(err)
	}
	copied := map[string]bool{}
	for _, c := range consumers {
		for to, from := range c.Copies {
			copied[from] = true
			if _, err := os.Stat(filepath.Join("../..", from)); err != nil {
				t.Errorf("%s copies %s to %s: %v", c.Name, from, to, err)
			}
		}
	}
	for _, own := range laboratory {
		copied[own] = true
	}
	outputs, err := filepath.Glob("../../dist/*")
	if err != nil || len(outputs) == 0 {
		t.Fatalf("dist/: %v, %d files", err, len(outputs))
	}
	for _, out := range outputs {
		if name := "dist/" + filepath.Base(out); !copied[name] {
			t.Errorf("%s is copied by no consumer in consumers.toml", name)
		}
	}
}

func TestATableWithAFieldItDoesNotKnowIsRefused(t *testing.T) {
	cases := map[string]string{
		"a misspelt field":  "[[consumer]]\nname = \"a\"\nrepo = \"o/a\"\npin = \"versions.yaml\"\npackage = \"package.json\"\n",
		"one name twice":    "[[consumer]]\nname = \"a\"\nrepo = \"o/a\"\npin = \"v\"\n[[consumer]]\nname = \"a\"\nrepo = \"o/b\"\npin = \"v\"\n",
		"a repo with a URL": "[[consumer]]\nname = \"a\"\nrepo = \"https://github.com/o/a\"\npin = \"v\"\n",
		"no pin's place":    "[[consumer]]\nname = \"a\"\nrepo = \"o/a\"\n",
	}
	for name, table := range cases {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "consumers.toml")
			if err := os.WriteFile(path, []byte(table), 0o644); err != nil {
				t.Fatal(err)
			}
			if got, err := load(path); err == nil {
				t.Fatalf("read %v", got)
			}
		})
	}
}
