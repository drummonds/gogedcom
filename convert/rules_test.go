package convert

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// TestRulesMatchTheDocument: the tag table in docs/convert-551-to-7.md and
// tagRules name the same rule for every 5.5.1 tag in every place.
func TestRulesMatchTheDocument(t *testing.T) {
	src, err := os.ReadFile("../docs/convert-551-to-7.md")
	if err != nil {
		t.Fatal(err)
	}
	// A row of the tag table: | `TAG` | place | rule | note |, where place
	// is "record", "any" or a superstructure tag in backticks.
	row := regexp.MustCompile("^\\| `([A-Z0-9_]+)` \\| (record|any|`[A-Z]+`) \\| ([a-z ]+) \\|")
	documented := map[string]string{}
	for _, l := range strings.Split(string(src), "\n") {
		m := row.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		key := m[1]
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
	if len(documented) < 100 {
		t.Fatalf("%d rows read from the tag table; the 5.5.1 specification has over 130 tags", len(documented))
	}
	for key, rule := range tagRules {
		if documented[key] != string(rule) {
			t.Errorf("%s: code says %q, document says %q", key, rule, documented[key])
		}
	}
	for key, rule := range documented {
		if _, ok := tagRules[key]; !ok {
			t.Errorf("%s: documented as %q, not in tagRules", key, rule)
		}
	}
}
