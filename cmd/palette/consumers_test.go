package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spin-stack/go-tools/consumers"
)

// laboratory is what dist/ holds for this repository's own pages, which no consumer copies.
var laboratory = []string{"dist/perception.js", "dist/simulate.svg"}

// TestTheTableIsOfThisRepository holds consumers.toml to the tree: a copy of a file this
// repository does not have is a row nobody can satisfy, and an output nobody copies is either a
// consumer missing from the table or a file made for no one.
func TestTheTableIsOfThisRepository(t *testing.T) {
	table, err := consumers.Load(filepath.Join(root, consumers.File))
	if err != nil {
		t.Fatal(err)
	}
	copied := map[string]bool{}
	for _, c := range table.Consumers {
		for to, from := range c.Copies {
			copied[from] = true
			if _, err := os.Stat(filepath.Join(root, from)); err != nil {
				t.Errorf("%s copies %s to %s: %v", c.Name, from, to, err)
			}
		}
	}
	for _, own := range laboratory {
		copied[own] = true
	}
	outputs, err := filepath.Glob(filepath.Join(root, "dist/*"))
	if err != nil || len(outputs) == 0 {
		t.Fatalf("dist/: %v, %d files", err, len(outputs))
	}
	for _, out := range outputs {
		if name := "dist/" + filepath.Base(out); !copied[name] {
			t.Errorf("%s is copied by no consumer in consumers.toml", name)
		}
	}
}
