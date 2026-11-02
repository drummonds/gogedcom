package gedcom7

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func header() *Structure {
	return &Structure{Tag: "HEAD", Substructures: []*Structure{
		{Tag: "GEDC", Substructures: []*Structure{{Tag: "VERS", Payload: "7.0"}}},
	}}
}

func write(t *testing.T, doc *Document) string {
	t.Helper()
	var buf bytes.Buffer
	if err := Write(&buf, doc); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestWrite(t *testing.T) {
	cases := []struct {
		name   string
		record *Structure
		want   string // the record's lines
	}{
		{"record and substructure",
			&Structure{XRef: "@I1@", Tag: "INDI", Substructures: []*Structure{{Tag: "NAME", Payload: "John /Smith/"}}},
			"0 @I1@ INDI\n1 NAME John /Smith/\n"},
		{"line breaks as CONT, an empty remainder as a bare CONT",
			&Structure{XRef: "@N1@", Tag: "SNOTE", Payload: "first\nsecond\n"},
			"0 @N1@ SNOTE first\n1 CONT second\n1 CONT\n"},
		{"a leading line break",
			&Structure{XRef: "@I1@", Tag: "INDI", Substructures: []*Structure{{Tag: "NOTE", Payload: "\nx"}}},
			"0 @I1@ INDI\n1 NOTE\n2 CONT x\n"},
		{"a leading at-sign is doubled, on the line and on a CONT",
			&Structure{XRef: "@I1@", Tag: "INDI", Substructures: []*Structure{{Tag: "NOTE", Payload: "@x\n@y"}}},
			"0 @I1@ INDI\n1 NOTE @@x\n2 CONT @@y\n"},
		{"an internal at-sign is not",
			&Structure{XRef: "@I1@", Tag: "INDI", Substructures: []*Structure{{Tag: "NOTE", Payload: "me@example.com"}}},
			"0 @I1@ INDI\n1 NOTE me@example.com\n"},
		{"a pointer is written as it is",
			&Structure{XRef: "@F1@", Tag: "FAM", Substructures: []*Structure{{Tag: "HUSB", Payload: "@I1@"}, {Tag: "WIFE", Payload: "@VOID@"}}},
			"0 @F1@ FAM\n1 HUSB @I1@\n1 WIFE @VOID@\n"},
	}
	const head = "0 HEAD\n1 GEDC\n2 VERS 7.0\n"
	for _, c := range cases {
		got := write(t, &Document{Header: header(), Records: []*Structure{c.record}})
		if want := head + c.want + "0 TRLR\n"; got != want {
			t.Errorf("%s:\n got %q\nwant %q", c.name, got, want)
		}
	}
}

func TestWriteErrors(t *testing.T) {
	cases := []struct {
		name string
		doc  *Document
		want string
	}{
		{"carriage return",
			&Document{Header: header(), Records: []*Structure{{XRef: "@N1@", Tag: "SNOTE", Payload: "a\r\nb"}}},
			`gedcom7: @N1@ SNOTE: carriage return in payload`},
		{"tag with a space",
			&Document{Header: header(), Records: []*Structure{{XRef: "@I1@", Tag: "INDI", Substructures: []*Structure{{Tag: "bad tag"}}}}},
			`gedcom7: @I1@ INDI: "bad tag" is not a tag`},
		{"empty tag",
			&Document{Header: header(), Records: []*Structure{{XRef: "@I1@", Tag: "INDI", Substructures: []*Structure{{Tag: ""}}}}},
			`gedcom7: @I1@ INDI: "" is not a tag`},
		{"lowercase xref",
			&Document{Header: header(), Records: []*Structure{{XRef: "@a@", Tag: "INDI"}}},
			`gedcom7: @a@ INDI: "@a@" is not a cross-reference identifier`},
		{"void xref",
			&Document{Header: header(), Records: []*Structure{{XRef: "@VOID@", Tag: "INDI"}}},
			`gedcom7: @VOID@ INDI: "@VOID@" is not a cross-reference identifier`},
		{"xref without at-signs",
			&Document{Header: header(), Records: []*Structure{{XRef: "I1", Tag: "INDI"}}},
			`gedcom7: I1 INDI: "I1" is not a cross-reference identifier`},
		{"no header", &Document{}, `gedcom7: no HEAD`},
	}
	for _, c := range cases {
		err := Write(&bytes.Buffer{}, c.doc)
		if err == nil || err.Error() != c.want {
			t.Errorf("%s: err = %v, want %s", c.name, err, c.want)
		}
	}
}

func TestReadStripsAtSignEscape(t *testing.T) {
	src := "0 HEAD\n0 @I1@ INDI\n1 NOTE @@x\n1 NOTE x\n2 CONT @@y\n1 ALIA @I1@\n0 TRLR\n"
	doc, err := Read(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	notes := doc.Records[0].All("NOTE")
	if got := []string{notes[0].Payload, notes[1].Payload, doc.Records[0].Find("ALIA").Payload}; !reflect.DeepEqual(got, []string{"@x", "x\n@y", "@I1@"}) {
		t.Errorf("payloads = %q", got)
	}
}

// knownRejected are the specification's test files Read does not accept
// yet, with the issue. A file leaving the list must be removed from it.
var knownRejected = map[string]string{
	"date.ged":       "#19",
	"extensions.ged": "#19",
}

func TestRoundTrip(t *testing.T) {
	files, err := filepath.Glob("testdata/gedcom70/*.ged")
	if err != nil || len(files) == 0 {
		t.Fatalf("testdata: %v", err)
	}
	for _, f := range files {
		name := filepath.Base(f)
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		doc, err := Read(bytes.NewReader(src))
		if issue, known := knownRejected[name]; known {
			if err == nil {
				t.Errorf("%s: now read; remove it from knownRejected (%s)", name, issue)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		var buf bytes.Buffer
		if err := Write(&buf, doc); err != nil {
			t.Errorf("%s: write: %v", name, err)
			continue
		}
		again, err := Read(bytes.NewReader(buf.Bytes()))
		if err != nil {
			t.Errorf("%s: read back: %v", name, err)
			continue
		}
		if !sameDocument(doc, again) {
			t.Errorf("%s: read back differs", name)
		}
	}
}

// sameDocument is equality in every field but Line.
func sameDocument(a, b *Document) bool {
	if len(a.Records) != len(b.Records) || !sameStructure(a.Header, b.Header) {
		return false
	}
	for i := range a.Records {
		if !sameStructure(a.Records[i], b.Records[i]) {
			return false
		}
	}
	return true
}

func sameStructure(a, b *Structure) bool {
	if a.XRef != b.XRef || a.Tag != b.Tag || a.Payload != b.Payload || len(a.Substructures) != len(b.Substructures) {
		return false
	}
	for i := range a.Substructures {
		if !sameStructure(a.Substructures[i], b.Substructures[i]) {
			return false
		}
	}
	return true
}
