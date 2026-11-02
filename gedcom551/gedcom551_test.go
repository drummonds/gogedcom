package gedcom551

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"git.bytestone.uk/hum3/gogedcom/gedcom"
)

const small = "\xef\xbb\xbf0 HEAD\r\n1 GEDC\r\n2 VERS 5.5.1\r\n0 @I1@ INDI\r\n1 NAME John /Smith/\r\n0 TRLR\r\n"

func TestRead(t *testing.T) {
	doc, err := Read(strings.NewReader(small))
	if err != nil {
		t.Fatal(err)
	}
	if doc.Header == nil || doc.Header.Tag != "HEAD" || doc.Header.Find("GEDC").Find("VERS").Payload != "5.5.1" {
		t.Fatalf("header = %+v", doc.Header)
	}
	if len(doc.Records) != 1 {
		t.Fatalf("records = %d", len(doc.Records))
	}
	r := doc.Records[0]
	if r.XRef != "@I1@" || r.Tag != "INDI" || r.Line != 4 || len(r.Substructures) != 1 {
		t.Fatalf("record = %+v", r)
	}
	if n := r.Substructures[0]; n.Tag != "NAME" || n.Payload != "John /Smith/" || n.Line != 5 {
		t.Fatalf("name = %+v", n)
	}
}

// file wraps lines in a header and trailer.
func file(lines ...string) string {
	return "0 HEAD\n" + strings.Join(lines, "\n") + "\n0 TRLR\n"
}

func TestReadPayloads(t *testing.T) {
	cases := []struct {
		name  string
		lines []string
		want  string // the payload of the first record's first substructure
	}{
		{"CONC joins without a separator, CONT with a line break", []string{"0 @N1@ NOTE He was", "1 CONC  awarded the MC", "1 CONT in 1916"}, "He was awarded the MC\nin 1916"},
		{"an empty CONT", []string{"0 @N1@ NOTE a", "1 CONT", "1 CONT b"}, "a\n\nb"},
		{"a trailing space is kept", []string{"0 @N1@ NOTE ends in space "}, "ends in space "},
		{"a tab is kept", []string{"0 @N1@ NOTE a\ttab"}, "a\ttab"},
		{"a lone at sign is text", []string{"0 @N1@ NOTE humphrey.drummond@gmail.com"}, "humphrey.drummond@gmail.com"},
		{"an xref followed by text is text", []string{"0 @N1@ NOTE @R1203558009@  Ancestry Family Trees"}, "@R1203558009@  Ancestry Family Trees"},
		{"a doubled at sign is kept as written", []string{"0 @N1@ NOTE mail @@home"}, "mail @@home"},
		{"a pointer", []string{"0 @N1@ FAMC @F1@"}, "@F1@"},
		{"a DATE its data type rejects is text", []string{"0 @N1@ DATE 5  1801"}, "5  1801"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			doc, err := Read(strings.NewReader(file(c.lines...)))
			if err != nil {
				t.Fatal(err)
			}
			r := doc.Records[0]
			if r.Payload != c.want {
				t.Fatalf("payload = %q, want %q", r.Payload, c.want)
			}
			if len(r.Substructures) != 0 {
				t.Fatalf("substructures = %+v, want none", r.Substructures)
			}
		})
	}
}

func TestReadLines(t *testing.T) {
	cases := []struct {
		name string
		src  string
		ok   bool
		err  string // a substring of the error when !ok
	}{
		{"a two-digit level", file("0 @X@ _T", strings.Repeat("1 A\n2 B\n3 C\n4 D\n5 E\n6 F\n7 G\n8 H\n9 I\n10 J", 1)), true, ""},
		{"leading white space is ignored", "0 HEAD\n  1 GEDC\n\t2 VERS 5.5.1\n0 TRLR\n", true, ""},
		{"an empty line is ignored", "0 HEAD\n1 GEDC\n\n2 VERS 5.5.1\n0 TRLR\n", true, ""},
		{"LF CR line ends", "0 HEAD\n\r1 GEDC\n\r0 TRLR\n\r", true, ""},
		{"CR line ends", "0 HEAD\r1 GEDC\r0 TRLR\r", true, ""},
		{"no line end on the last line", "0 HEAD\n0 TRLR", true, ""},
		{"a lower-case tag", file("0 @X@ _T", "1 Note lower"), true, ""},
		{"an xref with a space, bang and colon", file("0 @I 1!2:a@ INDI"), true, ""},
		{"two spaces after the level", file("0 @X@ _T", "1  NAME x"), false, "line 3"},
		{"a leading zero on the level", file("0 @X@ _T", "01 NAME x"), false, "line 3"},
		{"a delimiter with no value", file("0 @X@ _T", "1 NAME "), false, "line 3"},
		{"level 2 under level 0", file("0 @X@ _T", "2 NAME x"), false, "line 3"},
		{"no TRLR", "0 HEAD\n0 @X@ _T\n", false, "TRLR"},
		{"a record before HEAD", "0 @X@ _T\n0 HEAD\n0 TRLR\n", false, "HEAD"},
		{"a line after TRLR", "0 HEAD\n0 TRLR\n0 @X@ _T\n", false, "TRLR"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := Read(strings.NewReader(c.src))
			switch {
			case c.ok && err != nil:
				t.Fatalf("Read: %v", err)
			case !c.ok && err == nil:
				t.Fatal("Read should fail")
			case !c.ok && !strings.Contains(err.Error(), c.err):
				t.Fatalf("Read: %v, want %q in the error", err, c.err)
			}
		})
	}
}

func TestReadLeadingWhitespaceAndEmptyLinesKeepLineNumbers(t *testing.T) {
	doc, err := Read(strings.NewReader("0 HEAD\n\n  1 GEDC\n0 TRLR\n"))
	if err != nil {
		t.Fatal(err)
	}
	if g := doc.Header.Find("GEDC"); g == nil || g.Line != 3 {
		t.Fatalf("GEDC = %+v, want line 3", g)
	}
}

// TestReadGeni reads the author's Geni.com export when GOGEDCOM_GENI names
// it (story 006: the file is not in the repo).
func TestReadGeni(t *testing.T) {
	path := os.Getenv("GOGEDCOM_GENI")
	if path == "" {
		t.Skip("GOGEDCOM_GENI not set")
	}
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := Read(bytes.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, r := range doc.Records {
		counts[r.Tag]++
	}
	for tag, want := range map[string]int{"INDI": 1976, "FAM": 722, "NOTE": 164, "SUBM": 187} {
		if counts[tag] != want {
			t.Errorf("%s records = %d, want %d", tag, counts[tag], want)
		}
	}
	// Every file line is a structure, a fold (CONC, CONT) or the trailer.
	lines, folds := 0, 0
	for _, l := range bytes.Split(bytes.TrimRight(src, "\r\n"), []byte("\n")) {
		lines++
		f := bytes.Fields(l)
		if len(f) > 1 && (string(f[1]) == "CONC" || string(f[1]) == "CONT") {
			folds++
		}
	}
	var structures int
	var walk func(s *gedcom.Structure)
	walk = func(s *gedcom.Structure) {
		structures++
		if s.Tag == "CONC" || s.Tag == "CONT" {
			t.Errorf("line %d: %s not folded", s.Line, s.Tag)
		}
		for _, c := range s.Substructures {
			walk(c)
		}
	}
	walk(doc.Header)
	for _, r := range doc.Records {
		walk(r)
	}
	if structures+folds+1 != lines {
		t.Errorf("structures %d + folds %d + TRLR != lines %d", structures, folds, lines)
	}
}

func TestParseDate(t *testing.T) {
	d := func(cal string, day int, month string, year int, dual, epoch string) *Date {
		return &Date{Calendar: cal, Day: day, Month: month, Year: year, Dual: dual, Epoch: epoch}
	}
	cases := []struct {
		in   string
		want DateValue
	}{
		{"", DateValue{}},
		{"11 MAR 1961", DateValue{Kind: DateKind, From: d("", 11, "MAR", 1961, "", "")}},
		{"MAR 1961", DateValue{Kind: DateKind, From: d("", 0, "MAR", 1961, "", "")}},
		{"1961", DateValue{Kind: DateKind, From: d("", 0, "", 1961, "", "")}},
		{"@#DJULIAN@ 1 JAN 1700", DateValue{Kind: DateKind, From: d("JULIAN", 1, "JAN", 1700, "", "")}},
		{"@#DFRENCH R@ 1 VEND 8", DateValue{Kind: DateKind, From: d("FRENCH R", 1, "VEND", 8, "", "")}},
		{"@#DHEBREW@ 1 TSH 5700", DateValue{Kind: DateKind, From: d("HEBREW", 1, "TSH", 5700, "", "")}},
		{"100 B.C.", DateValue{Kind: DateKind, From: d("", 0, "", 100, "", "B.C.")}},
		{"15 APR 1699/00", DateValue{Kind: DateKind, From: d("", 15, "APR", 1699, "00", "")}},
		{"ABT 1850", DateValue{Kind: ApproximateKind, Qualifier: "ABT", From: d("", 0, "", 1850, "", "")}},
		{"CAL 1850", DateValue{Kind: ApproximateKind, Qualifier: "CAL", From: d("", 0, "", 1850, "", "")}},
		{"EST 1850", DateValue{Kind: ApproximateKind, Qualifier: "EST", From: d("", 0, "", 1850, "", "")}},
		{"BET 1850 AND 1860", DateValue{Kind: RangeKind, Qualifier: "BET", From: d("", 0, "", 1850, "", ""), To: d("", 0, "", 1860, "", "")}},
		{"AFT 1850", DateValue{Kind: RangeKind, Qualifier: "AFT", From: d("", 0, "", 1850, "", "")}},
		{"BEF 1850", DateValue{Kind: RangeKind, Qualifier: "BEF", To: d("", 0, "", 1850, "", "")}},
		{"FROM 1850 TO 1860", DateValue{Kind: PeriodKind, From: d("", 0, "", 1850, "", ""), To: d("", 0, "", 1860, "", "")}},
		{"FROM 1850", DateValue{Kind: PeriodKind, From: d("", 0, "", 1850, "", "")}},
		{"TO 1860", DateValue{Kind: PeriodKind, To: d("", 0, "", 1860, "", "")}},
		{"INT 1850 (about the war)", DateValue{Kind: InterpretedKind, From: d("", 0, "", 1850, "", ""), Phrase: "about the war"}},
		{"(spring 1850)", DateValue{Kind: PhraseKind, Phrase: "spring 1850"}},
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
			if got.String() != c.in {
				t.Fatalf("ParseDate(%q).String() = %q", c.in, got.String())
			}
		})
	}
	for _, in := range []string{"5  1801", "12 JAN", "Louise", "BCE 100", "1850 B.C.E.", "32 JAN 1990", "1699/00 JAN 1700", "12 jan 1990"} {
		if _, err := ParseDate(in); err == nil {
			t.Errorf("ParseDate(%q) should fail", in)
		}
	}
}
