package gedcom551

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"git.bytestone.uk/hum3/gogedcom/gedcom"
)

func header() *gedcom.Structure {
	return &gedcom.Structure{Tag: "HEAD", Substructures: []*gedcom.Structure{
		{Tag: "GEDC", Substructures: []*gedcom.Structure{{Tag: "VERS", Payload: "5.5.1"}}},
	}}
}

func write(t *testing.T, doc *gedcom.Document) string {
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
		record *gedcom.Structure
		want   string // the record's lines
	}{
		{"record and substructure",
			&gedcom.Structure{XRef: "@I1@", Tag: "INDI", Substructures: []*gedcom.Structure{{Tag: "NAME", Payload: "John /Smith/"}}},
			"0 @I1@ INDI\n1 NAME John /Smith/\n"},
		{"line breaks as CONT, an empty remainder as a bare CONT",
			&gedcom.Structure{XRef: "@N1@", Tag: "NOTE", Payload: "first\nsecond\n"},
			"0 @N1@ NOTE first\n1 CONT second\n1 CONT\n"},
		{"a leading line break",
			&gedcom.Structure{XRef: "@I1@", Tag: "INDI", Substructures: []*gedcom.Structure{{Tag: "NOTE", Payload: "\nx"}}},
			"0 @I1@ INDI\n1 NOTE\n2 CONT x\n"},
		{"at-signs are written as they are: leading, doubled, in a CONT",
			&gedcom.Structure{XRef: "@I1@", Tag: "INDI", Substructures: []*gedcom.Structure{{Tag: "NOTE", Payload: "@R1@  Ancestry\n@@home\nme@example.com"}}},
			"0 @I1@ INDI\n1 NOTE @R1@  Ancestry\n2 CONT @@home\n2 CONT me@example.com\n"},
		{"a trailing space and a tab are kept",
			&gedcom.Structure{XRef: "@I1@", Tag: "INDI", Substructures: []*gedcom.Structure{{Tag: "NOTE", Payload: "a\tb "}}},
			"0 @I1@ INDI\n1 NOTE a\tb \n"},
		{"a pointer is written as it is",
			&gedcom.Structure{XRef: "@F1@", Tag: "FAM", Substructures: []*gedcom.Structure{{Tag: "HUSB", Payload: "@I1@"}}},
			"0 @F1@ FAM\n1 HUSB @I1@\n"},
		{"a long payload is one line, not split",
			&gedcom.Structure{XRef: "@N1@", Tag: "NOTE", Payload: strings.Repeat("x", 300)},
			"0 @N1@ NOTE " + strings.Repeat("x", 300) + "\n"},
		{"5.5.1's tags and identifiers: lower case, a space in an identifier",
			&gedcom.Structure{XRef: "@I 1@", Tag: "INDI", Substructures: []*gedcom.Structure{{Tag: "Note", Payload: "x"}}},
			"0 @I 1@ INDI\n1 Note x\n"},
	}
	const head = "0 HEAD\n1 GEDC\n2 VERS 5.5.1\n"
	for _, c := range cases {
		got := write(t, &gedcom.Document{Header: header(), Records: []*gedcom.Structure{c.record}})
		if want := head + c.want + "0 TRLR\n"; got != want {
			t.Errorf("%s:\n got %q\nwant %q", c.name, got, want)
		}
	}
}

func TestWriteErrors(t *testing.T) {
	cases := []struct {
		name string
		doc  *gedcom.Document
		want string
	}{
		{"carriage return",
			&gedcom.Document{Header: header(), Records: []*gedcom.Structure{{XRef: "@N1@", Tag: "NOTE", Payload: "a\r\nb"}}},
			`gedcom551: @N1@ NOTE: carriage return in payload`},
		{"tag with a hyphen",
			&gedcom.Document{Header: header(), Records: []*gedcom.Structure{{XRef: "@I1@", Tag: "INDI", Substructures: []*gedcom.Structure{{Tag: "A-B"}}}}},
			`gedcom551: @I1@ INDI: "A-B" is not a tag`},
		{"empty tag",
			&gedcom.Document{Header: header(), Records: []*gedcom.Structure{{XRef: "@I1@", Tag: "INDI", Substructures: []*gedcom.Structure{{Tag: ""}}}}},
			`gedcom551: @I1@ INDI: "" is not a tag`},
		{"identifier beginning with a space",
			&gedcom.Document{Header: header(), Records: []*gedcom.Structure{{XRef: "@ 1@", Tag: "INDI"}}},
			`gedcom551: @ 1@ INDI: "@ 1@" is not a cross-reference identifier`},
		{"identifier without at-signs",
			&gedcom.Document{Header: header(), Records: []*gedcom.Structure{{XRef: "I1", Tag: "INDI"}}},
			`gedcom551: I1 INDI: "I1" is not a cross-reference identifier`},
		{"no header", &gedcom.Document{}, `gedcom551: no HEAD`},
	}
	for _, c := range cases {
		err := Write(&bytes.Buffer{}, c.doc)
		if err == nil || err.Error() != c.want {
			t.Errorf("%s: err = %v, want %s", c.name, err, c.want)
		}
	}
}

// TestWriteReadsBack: what Write writes, Read reads to the same tree in
// every field but Line.
func TestWriteReadsBack(t *testing.T) {
	doc := &gedcom.Document{Header: header(), Records: []*gedcom.Structure{
		{XRef: "@I1@", Tag: "INDI", Substructures: []*gedcom.Structure{
			{Tag: "NAME", Payload: "John /Smith/", Substructures: []*gedcom.Structure{{Tag: "_MARNM", Payload: "Jones"}}},
			{Tag: "NOTE", Payload: "@R1@  Ancestry\n@@home\nme@example.com\n\nends in space \n"},
			{Tag: "NOTE", Payload: "\n a tab\t"},
			{Tag: "BIRT", Substructures: []*gedcom.Structure{{Tag: "DATE", Payload: "5  1801"}}},
			{Tag: "FAMC", Payload: "@F1@"},
		}},
		{XRef: "@F1@", Tag: "FAM", Substructures: []*gedcom.Structure{{Tag: "CHIL", Payload: "@I1@"}}},
	}}
	src := write(t, doc)
	back, err := Read(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	clearLines(back.Header)
	for _, r := range back.Records {
		clearLines(r)
	}
	if !reflect.DeepEqual(doc, back) {
		t.Errorf("read back differs:\n%s", src)
	}
}

func clearLines(s *gedcom.Structure) {
	s.Line = 0
	for _, c := range s.Substructures {
		clearLines(c)
	}
}
