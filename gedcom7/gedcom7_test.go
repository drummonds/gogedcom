package gedcom7

import (
	"strings"
	"testing"
)

func TestGrammarIsComplete(t *testing.T) {
	g := Grammar()
	if got := g.Undefined(); len(got) != 1 || got[0] != "URI-reference" {
		t.Fatalf("Undefined = %v; only URI-reference (RFC 3986) is external", got)
	}
	// The 7.0.17 typo: Longitude took N/S.
	if _, err := g.Match("Longitude", "W1.5"); err != nil {
		t.Fatalf("Longitude W1.5: %v", err)
	}
	if _, err := g.Match("Longitude", "N1.5"); err == nil {
		t.Fatal("Longitude N1.5 should not match")
	}
}

func TestParseDate(t *testing.T) {
	d := func(cal string, day int, month string, year int, epoch string) *Date {
		return &Date{Calendar: cal, Day: day, Month: month, Year: year, Epoch: epoch}
	}
	cases := []struct {
		in   string
		want DateValue
	}{
		{"", DateValue{}},
		{"12 JAN 1990", DateValue{Kind: DateKind, From: d("", 12, "JAN", 1990, "")}},
		{"JAN 1990", DateValue{Kind: DateKind, From: d("", 0, "JAN", 1990, "")}},
		{"1990", DateValue{Kind: DateKind, From: d("", 0, "", 1990, "")}},
		{"JULIAN 1 JAN 1700", DateValue{Kind: DateKind, From: d("JULIAN", 1, "JAN", 1700, "")}},
		{"JULIAN 1700", DateValue{Kind: DateKind, From: d("JULIAN", 0, "", 1700, "")}},
		{"100 BCE", DateValue{Kind: DateKind, From: d("", 0, "", 100, "BCE")}},
		{"ABT 1850", DateValue{Kind: ApproximateKind, Qualifier: "ABT", From: d("", 0, "", 1850, "")}},
		{"CAL 1850", DateValue{Kind: ApproximateKind, Qualifier: "CAL", From: d("", 0, "", 1850, "")}},
		{"EST 1850", DateValue{Kind: ApproximateKind, Qualifier: "EST", From: d("", 0, "", 1850, "")}},
		{"BET 1850 AND 1860", DateValue{Kind: RangeKind, Qualifier: "BET", From: d("", 0, "", 1850, ""), To: d("", 0, "", 1860, "")}},
		{"AFT 1850", DateValue{Kind: RangeKind, Qualifier: "AFT", From: d("", 0, "", 1850, "")}},
		{"BEF 1850", DateValue{Kind: RangeKind, Qualifier: "BEF", To: d("", 0, "", 1850, "")}},
		{"FROM 1850 TO 1860", DateValue{Kind: PeriodKind, From: d("", 0, "", 1850, ""), To: d("", 0, "", 1860, "")}},
		{"FROM 1850", DateValue{Kind: PeriodKind, From: d("", 0, "", 1850, "")}},
		{"TO 1860", DateValue{Kind: PeriodKind, To: d("", 0, "", 1860, "")}},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got, err := ParseDate(c.in)
			if err != nil {
				t.Fatalf("ParseDate(%q): %v", c.in, err)
			}
			if !got.Equal(c.want) {
				t.Fatalf("ParseDate(%q) = %s, want %s", c.in, got, c.want)
			}
		})
	}
	for _, in := range []string{"12 JAN", "1990 JAN 12", "ABT", "BET 1850", "12 jan 1990"} {
		if _, err := ParseDate(in); err == nil {
			t.Errorf("ParseDate(%q) should fail", in)
		}
	}
}

const small = "0 HEAD\n1 GEDC\n2 VERS 7.0\n0 @I1@ INDI\n1 NAME John /Smith/\n0 TRLR\n"

func TestRead(t *testing.T) {
	doc, err := Read(strings.NewReader(small))
	if err != nil {
		t.Fatal(err)
	}
	if doc.Header == nil || doc.Header.Tag != "HEAD" || doc.Header.Find("GEDC").Find("VERS").Payload != "7.0" {
		t.Fatalf("header = %+v", doc.Header)
	}
	if len(doc.Records) != 1 {
		t.Fatalf("records = %d", len(doc.Records))
	}
	r := doc.Records[0]
	if r.XRef != "@I1@" || r.Tag != "INDI" || len(r.Substructures) != 1 {
		t.Fatalf("record = %+v", r)
	}
	if n := r.Find("NAME"); n == nil || n.Payload != "John /Smith/" {
		t.Fatalf("NAME = %+v", n)
	}
	if r.Find("BIRT") != nil {
		t.Fatal("Find of an absent tag should be nil")
	}
}

func TestReadFoldsCONT(t *testing.T) {
	src := "0 HEAD\n0 @N1@ SNOTE first\n1 CONT second\n1 CONT\n0 TRLR\n"
	doc, err := Read(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	r := doc.Records[0]
	if r.Payload != "first\nsecond\n" || len(r.Substructures) != 0 {
		t.Fatalf("payload = %q, substructures = %d", r.Payload, len(r.Substructures))
	}
}

func TestReadAcceptsBOMAndCRLF(t *testing.T) {
	src := "\xef\xbb\xbf0 HEAD\r\n0 TRLR\r\n"
	if _, err := Read(strings.NewReader(src)); err != nil {
		t.Fatal(err)
	}
}

func TestReadErrors(t *testing.T) {
	cases := map[string]string{
		"two spaces":    "0 HEAD\n0 @I1@ INDI\n1  NAME x\n0 TRLR\n",
		"no trailer":    "0 HEAD\n0 @I1@ INDI\n",
		"level skipped": "0 HEAD\n0 @I1@ INDI\n2 NAME x\n0 TRLR\n",
		"no header":     "0 @I1@ INDI\n0 TRLR\n",
	}
	for name, src := range cases {
		_, err := Read(strings.NewReader(src))
		if err == nil {
			t.Errorf("%s: should fail", name)
			continue
		}
		if name == "two spaces" && !strings.Contains(err.Error(), "line 3") {
			t.Errorf("%s: err = %v, want the line number", name, err)
		}
	}
}
