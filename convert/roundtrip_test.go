package convert

import (
	"bytes"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"

	"git.bytestone.uk/hum3/gogedcom/gedcom"
	"git.bytestone.uk/hum3/gogedcom/gedcom551"
	"git.bytestone.uk/hum3/gogedcom/gedcom7"
)

// read551 reads 5.5.1 text to a tree.
func read551(t *testing.T, src string) *gedcom.Document {
	t.Helper()
	doc, err := gedcom551.Read(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

// TestDifferences pins story 006's comparison table: what counts as a
// difference between the first tree of the trip and the last.
func TestDifferences(t *testing.T) {
	head := "0 HEAD\n1 GEDC\n2 VERS 5.5.1\n"
	for _, c := range []struct {
		name string
		a, b string
		want []string
	}{
		{"provenance",
			head + "1 SOUR Geni.com\n2 VERS 1.0\n1 DATE 24 OCT 2026\n2 TIME 10:00:00\n0 TRLR\n",
			head + "1 SOUR gogedcom\n1 DATE 5 NOV 2026\n2 TIME 11:11:11\n0 TRLR\n",
			nil},
		{"GEDC.VERS",
			head + "0 TRLR\n",
			"0 HEAD\n1 GEDC\n2 VERS 5.5.5\n0 TRLR\n",
			nil},
		{"CHAR",
			head + "1 CHAR UTF-8\n0 TRLR\n",
			head + "1 CHAR ANSEL\n0 TRLR\n",
			[]string{`HEAD/CHAR: "UTF-8" vs "ANSEL"`}},
		{"payload",
			head + "0 @I1@ INDI\n1 BIRT\n2 DATE 1 JAN 1900\n0 TRLR\n",
			head + "0 @I1@ INDI\n1 BIRT\n2 DATE 1 JAN 1901\n0 TRLR\n",
			[]string{`@I1@ INDI/BIRT/DATE: "1 JAN 1900" vs "1 JAN 1901"`}},
		{"xref",
			head + "0 @I1@ INDI\n1 NAME A\n0 TRLR\n",
			head + "0 @I2@ INDI\n1 NAME A\n0 TRLR\n",
			[]string{`@I1@ INDI: @I1@ vs @I2@`}},
		{"substructure count",
			head + "0 @I1@ INDI\n1 BIRT\n2 DATE 1900\n2 PLAC X\n0 TRLR\n",
			head + "0 @I1@ INDI\n1 BIRT\n2 DATE 1900\n0 TRLR\n",
			[]string{`@I1@ INDI/BIRT: 2 vs 1 substructures`}},
		{"records swapped",
			head + "0 @I1@ INDI\n0 @I2@ INDI\n0 TRLR\n",
			head + "0 @I2@ INDI\n0 @I1@ INDI\n0 TRLR\n",
			[]string{`@I1@ INDI: @I1@ vs @I2@`, `@I2@ INDI: @I2@ vs @I1@`}},
		{"record count",
			head + "0 @I1@ INDI\n0 @I2@ INDI\n0 TRLR\n",
			head + "0 @I1@ INDI\n0 TRLR\n",
			[]string{`2 vs 1 records`}},
		{"line numbers",
			head + "0 @I1@ INDI\n1 NAME A\n0 TRLR\n",
			head + "\n\n0 @I1@ INDI\n1 NAME A\n0 TRLR\n",
			nil},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := differences(read551(t, c.a), read551(t, c.b))
			if !slices.Equal(got, c.want) {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

// TestRoundTripGeni is roadmap #7: the Geni export read as 5.5.1,
// converted to 7, written, read, converted back, written as 5.5.1 and
// read, the first tree equal to the last by story 006's rules.
func TestRoundTripGeni(t *testing.T) {
	path := os.Getenv("GOGEDCOM_GENI")
	if path == "" {
		t.Skip("GOGEDCOM_GENI not set")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	first, err := gedcom551.Read(f)
	if err != nil {
		t.Fatal(err)
	}
	seven, err := To7(first)
	if err != nil {
		t.Fatal(err)
	}
	var file7 bytes.Buffer
	if err := gedcom7.Write(&file7, seven); err != nil {
		t.Fatal(err)
	}
	seven, err = gedcom7.Read(&file7)
	if err != nil {
		t.Fatal(err)
	}
	var file551 bytes.Buffer
	if err := gedcom551.Write(&file551, To551(seven)); err != nil {
		t.Fatal(err)
	}
	last, err := gedcom551.Read(&file551)
	if err != nil {
		t.Fatal(err)
	}
	if d := differences(first, last); len(d) > 0 {
		t.Errorf("%d records differ, the first:\n%s", len(d), strings.Join(d[:min(len(d), 20)], "\n"))
	}
}

// differences lists where two trees differ by story 006's rules: the
// header's provenance (SOUR, DATE, GEDC.VERS) aside, every tag,
// identifier, payload, placement and order counts, Line does not. One line
// per record, the first difference in it, so nothing hides what follows.
func differences(a, b *gedcom.Document) []string {
	var out []string
	var walk func(path string, x, y *gedcom.Structure) bool
	walk = func(path string, x, y *gedcom.Structure) bool {
		if x.XRef != y.XRef {
			out = append(out, path+": "+x.XRef+" vs "+y.XRef)
			return true
		}
		if x.Tag != y.Tag {
			out = append(out, path+": "+x.Tag+" vs "+y.Tag)
			return true
		}
		if x.Payload != y.Payload {
			out = append(out, path+": "+strconv.Quote(x.Payload)+" vs "+strconv.Quote(y.Payload))
			return true
		}
		for i := range min(len(x.Substructures), len(y.Substructures)) {
			if walk(path+"/"+x.Substructures[i].Tag, x.Substructures[i], y.Substructures[i]) {
				return true
			}
		}
		if len(x.Substructures) != len(y.Substructures) {
			out = append(out, path+": "+strconv.Itoa(len(x.Substructures))+" vs "+strconv.Itoa(len(y.Substructures))+" substructures")
			return true
		}
		return false
	}
	walk("HEAD", withoutProvenance(a.Header), withoutProvenance(b.Header))
	for i := range min(len(a.Records), len(b.Records)) {
		r := a.Records[i]
		walk(r.XRef+" "+r.Tag, r, b.Records[i])
	}
	if len(a.Records) != len(b.Records) {
		out = append(out, strconv.Itoa(len(a.Records))+" vs "+strconv.Itoa(len(b.Records))+" records")
	}
	return out
}

// withoutProvenance is the header less the fields that name the writer.
func withoutProvenance(h *gedcom.Structure) *gedcom.Structure {
	out := &gedcom.Structure{XRef: h.XRef, Tag: h.Tag, Payload: h.Payload}
	for _, c := range h.Substructures {
		switch c.Tag {
		case "SOUR", "DATE":
			continue
		case "GEDC":
			g := &gedcom.Structure{Tag: c.Tag, Payload: c.Payload}
			for _, cc := range c.Substructures {
				if cc.Tag != "VERS" {
					g.Substructures = append(g.Substructures, cc)
				}
			}
			c = g
		}
		out.Substructures = append(out.Substructures, c)
	}
	return out
}
