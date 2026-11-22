// Command release names this repository's next release as spin-stack names releases:
// v<YYYYMMDD>.<NN>, the day in UTC and which release of that day (spin's docs/releases.md,
// go-tools' internal/release, whose rule this is: go-tools exports no command for it).
//
//	go run ./cmd/release <commit>
//
// It prints the name and refuses a commit a release already names. `go tool task release` makes
// the release; the tags are the clone's, which the task fetches first.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// name has two digits of sequence, never one: v20261109.1 and v20261109.01 are not two
// spellings of one release. The screens' archives are tagged screens-v..., which this is not.
var name = regexp.MustCompile(`^v(\d{8})\.(\d{2})$`)

// next is the name of the next release on now's day, given every tag there is.
func next(tags []string, now time.Time) string {
	day := now.UTC().Format("20060102")
	seq := 1
	for _, t := range tags {
		m := name.FindStringSubmatch(t)
		if m == nil || m[1] != day {
			continue
		}
		if n, _ := strconv.Atoi(m[2]); n >= seq {
			seq = n + 1
		}
	}
	return fmt.Sprintf("v%s.%02d", day, seq)
}

// gitFunc is a git command's output, one line a word: the clone's, and a test's own.
type gitFunc func(args ...string) ([]string, error)

func run(args []string, w io.Writer, git gitFunc, now time.Time) error {
	if len(args) != 1 {
		return errors.New("release <commit>")
	}
	at, err := git("tag", "--points-at", args[0])
	if err != nil {
		return err
	}
	for _, t := range at {
		if name.MatchString(t) {
			return fmt.Errorf("%s is released already, as %s", args[0], t)
		}
	}
	all, err := git("tag", "--list")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, next(all, now))
	return err
}

func main() {
	git := func(args ...string) ([]string, error) {
		out, err := exec.Command("git", args...).Output()
		if err != nil {
			return nil, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
		}
		return strings.Fields(string(out)), nil
	}
	if err := run(os.Args[1:], os.Stdout, git, time.Now()); err != nil {
		fmt.Fprintln(os.Stderr, "release:", err)
		os.Exit(1)
	}
}
