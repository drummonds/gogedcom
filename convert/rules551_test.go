package convert

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestRules551MatchTheDocument: the tag table in docs/convert-7-to-551.md
// and tagRules551 name the same rule for every 7 tag in every place, and
// every tag the registry has a structure term for has a row.
func TestRules551MatchTheDocument(t *testing.T) {
	src, err := os.ReadFile("../docs/convert-7-to-551.md")
	if err != nil {
		t.Fatal(err)
	}
	row := regexp.MustCompile("^\\| `([A-Z0-9_]+)` \\| (record|any|`[A-Z]+`) \\| ([a-z ]+) \\|")
	documented := map[string]string{}
	tags := map[string]bool{}
	for _, l := range strings.Split(string(src), "\n") {
		m := row.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		key := m[1]
		tags[key] = true
		switch place := strings.Trim(m[2], "`"); place {
		case "record":
		case "any":
			key = "*." + key
		default:
			key = place + "." + key
		}
		if _, dup := documented[key]; dup {
			t.Errorf("%s documented twice", key)
		}
		documented[key] = strings.TrimSpace(m[3])
	}
	for key, rule := range tagRules551 {
		if documented[key] != string(rule) {
			t.Errorf("%s: code says %q, document says %q", key, rule, documented[key])
		}
	}
	for key, rule := range documented {
		if _, ok := tagRules551[key]; !ok {
			t.Errorf("%s: documented as %q, not in tagRules551", key, rule)
		}
	}

	// Every structure term's standard tag.
	files, err := filepath.Glob("../gedcom7/registry/*")
	if err != nil || len(files) == 0 {
		t.Fatalf("registry: %v", err)
	}
	standardTag := regexp.MustCompile(`(?m)^standard tag: '?([A-Z0-9_]+)'?`)
	n := 0
	for _, f := range files {
		term, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(term), "\ntype: structure\n") {
			continue
		}
		m := standardTag.FindSubmatch(term)
		if m == nil {
			continue
		}
		n++
		if tag := string(m[1]); !tags[tag] {
			t.Errorf("%s: the registry's %s has no row", tag, filepath.Base(f))
		}
	}
	if n < 100 {
		t.Fatalf("%d structure terms read from the registry", n)
	}
}
