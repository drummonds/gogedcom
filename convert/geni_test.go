package convert

import (
	"bytes"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"git.bytestone.uk/hum3/gogedcom/gedcom"
	"git.bytestone.uk/hum3/gogedcom/gedcom551"
	"git.bytestone.uk/hum3/gogedcom/gedcom7"
)

// TestTo7Geni converts the author's Geni.com export when GOGEDCOM_GENI
// names it (story 006: the file is not in the repo).
func TestTo7Geni(t *testing.T) {
	path := os.Getenv("GOGEDCOM_GENI")
	if path == "" {
		t.Skip("GOGEDCOM_GENI not set")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	in, err := gedcom551.Read(f)
	if err != nil {
		t.Fatal(err)
	}
	out, err := To7(in)
	if err != nil {
		t.Fatal(err)
	}

	if v := gedcom7.Standard().Validate(out); len(v) > 0 {
		t.Errorf("%d violations, the first %v", len(v), v[0])
	}

	counts := map[string]int{}
	for _, r := range out.Records {
		counts[r.Tag]++
	}
	for tag, want := range map[string]int{"INDI": 1976, "FAM": 722, "SUBM": 187, "SNOTE": 164, "OBJE": 1448, "NOTE": 0, "SOUR": 0} {
		if counts[tag] != want {
			t.Errorf("%s records = %d, want %d", tag, counts[tag], want)
		}
	}

	var tags []string
	if schma := out.Header.Find("SCHMA"); schma != nil {
		for _, s := range schma.All("TAG") {
			tag, _, _ := strings.Cut(s.Payload, " ")
			tags = append(tags, tag)
		}
	}
	if want := []string{"_ADDR", "_EMAIL", "_PHON", "_PHRASE"}; !slices.Equal(tags, want) {
		t.Errorf("SCHMA tags = %q, want %q", tags, want)
	}

	// Every payload of the input is in the output verbatim, but for the
	// rows that re-spell: RFN (EXID drops the prefix), FORM (a media type),
	// VERS and CHAR (the header's constants).
	payloads := map[string]bool{}
	datePhrases := 0
	var collect func(s *gedcom.Structure)
	collect = func(s *gedcom.Structure) {
		payloads[s.Payload] = true
		if s.Tag == "DATE" && s.Find("PHRASE") != nil {
			datePhrases++
		}
		for _, c := range s.Substructures {
			collect(c)
		}
	}
	collect(out.Header)
	for _, r := range out.Records {
		collect(r)
	}
	if datePhrases != 1 {
		t.Errorf("DATE with PHRASE = %d, want 1", datePhrases)
	}
	missing := 0
	var check func(s *gedcom.Structure)
	check = func(s *gedcom.Structure) {
		switch s.Tag {
		case "RFN", "FORM", "VERS", "CHAR":
		default:
			if s.Payload != "" && !payloads[s.Payload] {
				if missing++; missing <= 5 {
					t.Errorf("line %d: %s %q not in the output", s.Line, s.Tag, s.Payload)
				}
			}
		}
		for _, c := range s.Substructures {
			check(c)
		}
	}
	check(in.Header)
	for _, r := range in.Records {
		check(r)
	}

	var buf bytes.Buffer
	if err := gedcom7.Write(&buf, out); err != nil {
		t.Fatal(err)
	}
	back, err := gedcom7.Read(&buf)
	if err != nil {
		t.Fatal(err)
	}
	clearLines(out.Header)
	clearLines(back.Header)
	for _, d := range [][]*gedcom.Structure{out.Records, back.Records} {
		for _, r := range d {
			clearLines(r)
		}
	}
	if !reflect.DeepEqual(out, back) {
		t.Error("the written file reads back differently")
	}
}

// TestTo551Geni: the Geni export through To7 and To551 comes back equal to
// what was read, and the 5.5.1 writer's file reads back the same.
func TestTo551Geni(t *testing.T) {
	path := os.Getenv("GOGEDCOM_GENI")
	if path == "" {
		t.Skip("GOGEDCOM_GENI not set")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	in, err := gedcom551.Read(f)
	if err != nil {
		t.Fatal(err)
	}
	seven, err := To7(in)
	if err != nil {
		t.Fatal(err)
	}
	back := To551(seven)

	counts := map[string]int{}
	for _, r := range back.Records {
		counts[r.Tag]++
	}
	for tag, want := range map[string]int{"INDI": 1976, "FAM": 722, "SUBM": 187, "NOTE": 164, "OBJE": 0, "SNOTE": 0} {
		if counts[tag] != want {
			t.Errorf("%s records = %d, want %d", tag, counts[tag], want)
		}
	}

	var buf bytes.Buffer
	if err := gedcom551.Write(&buf, back); err != nil {
		t.Fatal(err)
	}
	if n := bytes.Count(buf.Bytes(), []byte("\n")); n != 58627 {
		t.Errorf("written lines = %d, want 58627", n)
	}
	again, err := gedcom551.Read(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range []*gedcom.Document{in, back, again} {
		clearLines(d.Header)
		for _, r := range d.Records {
			clearLines(r)
		}
	}
	if !reflect.DeepEqual(in, back) {
		t.Error("the trip changed the tree")
		reportFirstDifference(t, in, back)
	}
	if !reflect.DeepEqual(back, again) {
		t.Error("the written file reads back differently")
	}
}

// reportFirstDifference names the first record and structure where the
// two trees differ.
func reportFirstDifference(t *testing.T, a, b *gedcom.Document) {
	t.Helper()
	var walk func(path string, x, y *gedcom.Structure) bool
	walk = func(path string, x, y *gedcom.Structure) bool {
		if x.Tag != y.Tag || x.Payload != y.Payload || x.XRef != y.XRef {
			t.Errorf("%s: %s %s %q vs %s %s %q", path, x.XRef, x.Tag, x.Payload, y.XRef, y.Tag, y.Payload)
			return true
		}
		for i := range min(len(x.Substructures), len(y.Substructures)) {
			if walk(path+"/"+x.Tag, x.Substructures[i], y.Substructures[i]) {
				return true
			}
		}
		if len(x.Substructures) != len(y.Substructures) {
			t.Errorf("%s/%s %s: %d vs %d substructures", path, x.Tag, x.Payload, len(x.Substructures), len(y.Substructures))
			return true
		}
		return false
	}
	if walk("HEAD", a.Header, b.Header) {
		return
	}
	for i := range min(len(a.Records), len(b.Records)) {
		if walk("", a.Records[i], b.Records[i]) {
			return
		}
	}
	if len(a.Records) != len(b.Records) {
		t.Errorf("%d vs %d records", len(a.Records), len(b.Records))
	}
}
