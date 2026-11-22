// Command consumers says which brand each place that draws Spinbox is on. For every consumer in
// consumers.toml it reads that repository's pin and its copies at its default branch, and prints
// the commit and the release it is on, how far behind main that is, whether each copy is what this
// repository made at that commit, and whether the faces it installs are the ones versions.yaml
// pinned there.
//
//	go tool task consumers
//	go tool task consumers -- -checkout spin-stack/spin=../spin -checkout spin-stack/t3code=../t3code@fork/spin
//
// A consumer is read through `gh api`, or from a checkout given by -checkout (at origin/HEAD, or
// the ref after the @). This repository is read from the clone it runs in, which has to hold the
// commits the consumers pin: a shallow clone does not.
//
// It is a report and exits 0 whatever it finds: a consumer that is behind is that repository's to
// fix, and this one's check must not fail for it. A repository that cannot be read is said so,
// and the others are still reported.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
	"go.yaml.in/yaml/v3"
)

// Consumer is one row of consumers.toml, which documents each field.
type Consumer struct {
	Name     string            `toml:"name"`
	Repo     string            `toml:"repo"`
	Pin      string            `toml:"pin"`
	Packages string            `toml:"packages"`
	Copies   map[string]string `toml:"copies"`
}

func load(path string) ([]Consumer, error) {
	var table struct {
		Consumer []Consumer `toml:"consumer"`
	}
	md, err := toml.DecodeFile(path, &table)
	if err != nil {
		return nil, err
	}
	if extra := md.Undecoded(); len(extra) > 0 {
		return nil, fmt.Errorf("%s: %s is no field of a consumer", path, extra[0])
	}
	seen := map[string]bool{}
	for _, c := range table.Consumer {
		switch {
		case c.Name == "" || seen[c.Name]:
			return nil, fmt.Errorf("%s: a consumer's name is its own: %q", path, c.Name)
		case !repoName.MatchString(c.Repo):
			return nil, fmt.Errorf("%s: %s: repo %q is not owner/name", path, c.Name, c.Repo)
		case c.Pin == "":
			return nil, fmt.Errorf("%s: %s names no versions.yaml for its pin", path, c.Name)
		}
		seen[c.Name] = true
	}
	return table.Consumer, nil
}

var repoName = regexp.MustCompile(`^[\w.-]+/[\w.-]+$`)

// source is a consumer's repository at its default branch.
type source interface {
	// name is what answered - GitHub's own name for the repository, a checkout and its ref - or
	// why nothing of it can be read.
	name() (string, error)
	// read is a file's bytes, and false when the repository has no such file.
	read(path string) ([]byte, bool, error)
}

type github struct{ repo string }

func (g github) name() (string, error) {
	out, err := output("gh", "api", "repos/"+g.repo, "--jq", ".full_name")
	if err != nil {
		if strings.Contains(err.Error(), "HTTP 404") {
			return "", errors.New("private or missing: GitHub answers 404 to this login")
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (g github) read(path string) ([]byte, bool, error) {
	out, err := output("gh", "api", "-H", "Accept: application/vnd.github.raw", "repos/"+g.repo+"/contents/"+path)
	if err != nil {
		if strings.Contains(err.Error(), "HTTP 404") {
			return nil, false, nil
		}
		return nil, false, err
	}
	return out, true, nil
}

type checkout struct{ dir, ref string }

func (c checkout) name() (string, error) {
	if _, err := output("git", "-C", c.dir, "rev-parse", "--verify", "--quiet", c.ref+"^{commit}"); err != nil {
		return "", fmt.Errorf("%s has no %s", c.dir, c.ref)
	}
	return c.dir + " at " + c.ref, nil
}

// read answers false for any path git cannot show: name has already said the ref is a commit.
func (c checkout) read(path string) ([]byte, bool, error) {
	out, err := output("git", "-C", c.dir, "show", c.ref+":"+path)
	return out, err == nil, nil
}

// brand is this repository: a clone, and the ref a consumer is behind.
type brand struct{ dir, main string }

func (b brand) git(args ...string) ([]byte, error) {
	return output("git", append([]string{"-C", b.dir}, args...)...)
}

// release is a release's name as spin-stack names them, v<YYYYMMDD>.<NN> (cmd/release). The
// screens' archives are releases of this repository too, under names this is not.
var release = regexp.MustCompile(`^v\d{8}\.\d{2}$`)

func (b brand) releases(args ...string) []string {
	out, _ := b.git(append([]string{"tag"}, args...)...)
	var names []string
	for _, t := range strings.Fields(string(out)) {
		if release.MatchString(t) {
			names = append(names, t)
		}
	}
	slices.Sort(names)
	return names
}

// pinned is the `brand` entry of a consumer's versions.yaml.
type pinned struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Pin     string `yaml:"pin"`
	Track   string `yaml:"track"`
}

// report is what one consumer is on, a line a fact.
func report(c Consumer, s source, b brand) []string {
	as, err := s.name()
	if err != nil {
		return []string{"unread   " + err.Error()}
	}
	var lines []string
	if as != c.Repo {
		lines = append(lines, "read as  "+as)
	}
	pin, why, err := readPin(c, s)
	if err != nil {
		return append(lines, "unread   "+err.Error())
	}
	at := b.main
	switch {
	case why != "":
		lines = append(lines, fmt.Sprintf("pin      none: %s. Held to nothing, and compared with %s", why, b.main))
	default:
		if _, err := b.git("cat-file", "-e", pin.Pin+"^{commit}"); err != nil {
			return append(lines, fmt.Sprintf("pin      %s (%s), which is no commit of brand in this clone", pin.Pin, pin.Version))
		}
		at = pin.Pin
		lines = append(lines, "pin      "+b.where(pin))
	}
	lines = append(lines, copies(c, s, b, at)...)
	if c.Packages != "" {
		faces, err := fonts(c, s, b, at)
		if err != nil {
			return append(lines, "unread   "+err.Error())
		}
		lines = append(lines, faces...)
	}
	return lines
}

// readPin is the consumer's brand entry, or why it has none.
func readPin(c Consumer, s source) (pinned, string, error) {
	raw, found, err := s.read(c.Pin)
	if err != nil {
		return pinned{}, "", err
	}
	if !found {
		return pinned{}, "it has no " + c.Pin, nil
	}
	var entries []pinned
	if err := yaml.Unmarshal(raw, &entries); err != nil {
		return pinned{}, "", fmt.Errorf("%s: %w", c.Pin, err)
	}
	for _, e := range entries {
		if e.Name == "brand" && e.Pin != "" {
			return e, "", nil
		}
	}
	return pinned{}, "its " + c.Pin + " has no brand entry", nil
}

// where is a pin as a place in this repository's history: the release that names it, and how far
// main and the newest release have moved past it.
func (b brand) where(p pinned) string {
	said := fmt.Sprintf("%s (%s, track %s)", short(p.Pin), p.Version, p.Track)
	all := b.releases("--merged", b.main)
	switch on := b.releases("--points-at", p.Pin); {
	case len(all) == 0:
		said += ", and brand has no release"
	case len(on) == 0:
		said += ", which no release names; the newest is " + all[len(all)-1]
	case on[len(on)-1] == all[len(all)-1]:
		said += ", the newest release " + on[len(on)-1]
	default:
		said += fmt.Sprintf(", release %s; the newest is %s", on[len(on)-1], all[len(all)-1])
	}
	behind, err := b.git("rev-list", "--count", p.Pin+".."+b.main)
	if err != nil {
		return said + "; " + err.Error()
	}
	return fmt.Sprintf("%s; %s commits behind %s", said, strings.TrimSpace(string(behind)), b.main)
}

// copies holds each of the consumer's copies to what this repository had at commit at.
func copies(c Consumer, s source, b brand, at string) []string {
	if len(c.Copies) == 0 {
		return []string{"copies   none"}
	}
	var off []string
	for _, to := range slices.Sorted(maps.Keys(c.Copies)) {
		from := c.Copies[to]
		want, err := b.git("show", at+":"+from)
		if err != nil {
			off = append(off, fmt.Sprintf("  %s: brand had no %s then", to, from))
			continue
		}
		switch got, found, err := s.read(to); {
		case err != nil:
			off = append(off, fmt.Sprintf("  %s: %v", to, err))
		case !found:
			off = append(off, fmt.Sprintf("  %s is not there (%s)", to, from))
		case !bytes.Equal(got, want):
			off = append(off, fmt.Sprintf("  %s is not %s", to, from))
		}
	}
	if len(off) == 0 {
		return []string{fmt.Sprintf("copies   all %d are brand's at %s", len(c.Copies), short(at))}
	}
	head := fmt.Sprintf("copies   %d of %d are not brand's at %s:", len(off), len(c.Copies), short(at))
	return append([]string{head}, off...)
}

// npm is where versions.yaml says a face's package comes from; the package's name is the path up
// to /-/.
const npm = "https://registry.npmjs.org/"

// fonts holds the @fontsource packages the consumer installs to the ones this repository's
// versions.yaml pinned at commit at: the faces the outputs of that commit name.
func fonts(c Consumer, s source, b brand, at string) ([]string, error) {
	want := map[string]string{}
	if raw, err := b.git("show", at+":versions.yaml"); err == nil {
		var entries []struct {
			Source  string `yaml:"source"`
			Version string `yaml:"version"`
		}
		if err := yaml.Unmarshal(raw, &entries); err != nil {
			return nil, fmt.Errorf("brand's versions.yaml at %s: %w", short(at), err)
		}
		for _, e := range entries {
			if rest, ok := strings.CutPrefix(e.Source, npm); ok {
				pkg, _, _ := strings.Cut(rest, "/-/")
				want[pkg] = e.Version
			}
		}
	}
	raw, found, err := s.read(c.Packages)
	if err != nil {
		return nil, err
	}
	if !found {
		return []string{"fonts    it has no " + c.Packages}, nil
	}
	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(raw, &pkg); err != nil {
		return nil, fmt.Errorf("%s: %w", c.Packages, err)
	}
	has := map[string]string{}
	maps.Copy(has, pkg.DevDependencies)
	maps.Copy(has, pkg.Dependencies)

	var off []string
	for _, name := range slices.Sorted(maps.Keys(want)) {
		switch got, ok := has[name]; {
		case !ok:
			off = append(off, fmt.Sprintf("  %s %s is not installed", name, want[name]))
		case got != want[name]:
			off = append(off, fmt.Sprintf("  %s is %s, and brand's is %s", name, got, want[name]))
		}
	}
	for _, name := range slices.Sorted(maps.Keys(has)) {
		if _, ok := want[name]; !ok && strings.HasPrefix(name, "@fontsource") {
			off = append(off, fmt.Sprintf("  %s %s is no face brand pins", name, has[name]))
		}
	}
	if len(off) == 0 {
		return []string{fmt.Sprintf("fonts    brand's at %s: %s", short(at), strings.Join(slices.Sorted(maps.Keys(want)), ", "))}, nil
	}
	head := fmt.Sprintf("fonts    not brand's at %s, which pins %d:", short(at), len(want))
	if len(want) == 0 {
		head = fmt.Sprintf("fonts    brand pinned no face at %s, and it installs:", short(at))
	}
	return append([]string{head}, off...), nil
}

func short(commit string) string {
	if len(commit) == 40 {
		return commit[:12]
	}
	return commit
}

func output(name string, args ...string) ([]byte, error) {
	var stderr bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w: %s", name, args[0], err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

func run(args []string, w io.Writer) error {
	fs := flag.NewFlagSet("consumers", flag.ContinueOnError)
	table := fs.String("table", "consumers.toml", "the consumers")
	dir := fs.String("brand", ".", "a clone of this repository that holds the commits the consumers pin")
	ref := fs.String("main", "origin/main", "the ref a consumer is behind")
	local := map[string]checkout{}
	fs.Func("checkout", "owner/name=dir[@ref]: read that consumer from a checkout, at origin/HEAD or ref, and not through gh", func(v string) error {
		repo, at, ok := strings.Cut(v, "=")
		if !ok {
			return errors.New("want owner/name=dir[@ref]")
		}
		d, ref, ok := strings.Cut(at, "@")
		if !ok {
			ref = "origin/HEAD"
		}
		local[repo] = checkout{d, ref}
		return nil
	})
	if err := fs.Parse(args); err != nil {
		return err
	}
	consumers, err := load(*table)
	if err != nil {
		return err
	}
	b := brand{*dir, *ref}
	head, err := b.git("rev-parse", "--verify", b.main+"^{commit}")
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "brand %s is %s\n", b.main, short(strings.TrimSpace(string(head))))
	for _, c := range consumers {
		var s source = github{c.Repo}
		if co, ok := local[c.Repo]; ok {
			s = co
		}
		fmt.Fprintf(w, "\n%s  %s\n", c.Name, c.Repo)
		for _, line := range report(c, s, b) {
			fmt.Fprintln(w, "  "+line)
		}
	}
	return nil
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			fmt.Fprintln(os.Stderr, "consumers:", err)
		}
		os.Exit(1)
	}
}
