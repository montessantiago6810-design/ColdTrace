package store

import (
	"fmt"
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

var migrationName = regexp.MustCompile(`^(\d{5})_[a-z0-9_]+\.sql$`)

func TestMigrationsAreSequentialAndReversible(t *testing.T) {
	entries, err := fs.ReadDir(Migrations(), ".")
	if err != nil {
		t.Fatalf("read migrations: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("no migrations embedded")
	}
	for i, e := range entries {
		m := migrationName.FindStringSubmatch(e.Name())
		if m == nil {
			t.Fatalf("%s: name must be NNNNN_snake_case.sql", e.Name())
		}
		if want := i + 1; m[1] != leftPad(want) {
			t.Errorf("%s: expected version %s (no gaps)", e.Name(), leftPad(want))
		}
		body, err := fs.ReadFile(Migrations(), e.Name())
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		for _, marker := range []string{"-- +goose Up", "-- +goose Down"} {
			if !strings.Contains(string(body), marker) {
				t.Errorf("%s: missing %q", e.Name(), marker)
			}
		}
	}
}

func leftPad(n int) string {
	return fmt.Sprintf("%05d", n)
}
