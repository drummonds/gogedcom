package convert

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"git.bytestone.uk/hum3/gogedcom/gedcom"
	"git.bytestone.uk/hum3/gogedcom/gedcom551"
	"git.bytestone.uk/hum3/gogedcom/gedcom7"
)

// head551 is a 5.5.1 header of five lines, so a record's first line is
// line 6.
const head551 = "0 HEAD\n1 GEDC\n2 VERS 5.5.1\n2 FORM LINEAGE-LINKED\n1 CHAR UTF-8\n"

const head7 = "0 HEAD\n1 GEDC\n2 VERS 7.0\n"

// to7 reads src as 5.5.1, converts it and writes it as 7.
func to7(t *testing.T, src string) string {
	t.Helper()
	doc, err := gedcom551.Read(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	out, err := To7(doc)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := gedcom7.Write(&buf, out); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func lines(l ...string) string { return strings.Join(l, "\n") + "\n" }

func TestTo7(t *testing.T) {
	const v7 = "https://gedcom.io/terms/v7/"
	cases := []struct {
		name  string
		in    string // 5.5.1 record lines
		schma string // the SCHMA TAG lines expected in the header, "" for no SCHMA
		want  string // 7 record lines
	}{
		{"NOTE record and pointer become SNOTE",
			lines("0 @N1@ NOTE a", "0 @I1@ INDI", "1 NOTE @N1@", "1 NOTE inline"), "",
			lines("0 @N1@ SNOTE a", "0 @I1@ INDI", "1 SNOTE @N1@", "1 NOTE inline")},
		{"RFN, AFN, RIN become EXID",
			lines("0 @I1@ INDI", "1 RFN geni:42", "1 AFN 123", "1 RIN 9"), "",
			lines("0 @I1@ INDI", "1 EXID 42", "2 TYPE "+v7+"RFN#geni", "1 EXID 123", "2 TYPE "+v7+"AFN", "1 EXID 9", "2 TYPE "+v7+"RIN")},
		{"RELA becomes ROLE, from the set or OTHER with PHRASE",
			lines("0 @I1@ INDI", "1 ASSO @I2@", "2 RELA Honorary uncle", "1 ASSO @I3@", "2 RELA godp"), "",
			lines("0 @I1@ INDI", "1 ASSO @I2@", "2 ROLE OTHER", "3 PHRASE Honorary uncle", "1 ASSO @I3@", "2 ROLE GODP")},
		{"a citation ROLE the same way",
			lines("0 @I1@ INDI", "1 SOUR @S1@", "2 EVEN BIRT", "3 ROLE (Godfather)", "1 SOUR @S1@", "2 EVEN BIRT", "3 ROLE CHIL"), "",
			lines("0 @I1@ INDI", "1 SOUR @S1@", "2 EVEN BIRT", "3 ROLE OTHER", "4 PHRASE (Godfather)", "1 SOUR @S1@", "2 EVEN BIRT", "3 ROLE CHIL")},
		{"WAC becomes INIL",
			lines("0 @I1@ INDI", "1 WAC", "2 DATE 1 JAN 1990"), "",
			lines("0 @I1@ INDI", "1 INIL", "2 DATE 1 JAN 1990")},
		{"enumerations upper-cased when in the set, renamed, else unchanged",
			lines("0 @I1@ INDI", "1 SEX M", "1 RESN privacy", "1 FAMC @F1@", "2 PEDI birth", "2 STAT challenged", "1 FAMC @F2@", "2 PEDI step",
				"1 BAPL", "2 STAT DNS/CAN", "1 ENDL", "2 STAT PRE-1970", "1 ADOP", "2 FAMC @F1@", "3 ADOP both"), "",
			lines("0 @I1@ INDI", "1 SEX M", "1 RESN PRIVACY", "1 FAMC @F1@", "2 PEDI BIRTH", "2 STAT CHALLENGED", "1 FAMC @F2@", "2 PEDI step",
				"1 BAPL", "2 STAT DNS_CAN", "1 ENDL", "2 STAT PRE_1970", "1 ADOP", "2 FAMC @F1@", "3 ADOP BOTH")},
		{"an inline OBJE becomes a record, TITL under FILE, FORM a media type, _PRIM on the pointer",
			lines("0 @I1@ INDI", "1 OBJE", "2 FILE a.jpg", "3 FORM jpg", "4 TYPE photo", "2 TITL t", "2 _PRIM Y", "2 NOTE n"), "",
			lines("0 @I1@ INDI", "1 OBJE @O1@", "2 _PRIM Y", "0 @O1@ OBJE", "1 FILE a.jpg", "2 FORM image/jpeg", "3 MEDI PHOTO", "2 TITL t", "1 NOTE n")},
		{"an OBJE record keeps its identifier; a new one skips it",
			lines("0 @O1@ OBJE", "1 FILE f", "2 FORM png", "0 @I1@ INDI", "1 OBJE", "2 FILE g", "3 FORM gif", "1 OBJE @O1@"), "",
			lines("0 @O1@ OBJE", "1 FILE f", "2 FORM image/png", "0 @I1@ INDI", "1 OBJE @O2@", "1 OBJE @O1@", "0 @O2@ OBJE", "1 FILE g", "2 FORM image/gif")},
		{"an empty inline SOUR becomes SOUR @VOID@ with its substructures",
			lines("0 @I1@ INDI", "1 SOUR", "2 DATA", "3 DATE 1 JAN 1990", "3 TEXT t", "2 OBJE", "3 FILE f", "4 FORM pdf"), "",
			lines("0 @I1@ INDI", "1 SOUR @VOID@", "2 DATA", "3 DATE 1 JAN 1990", "3 TEXT t", "2 OBJE @O1@", "0 @O1@ OBJE", "1 FILE f", "2 FORM application/pdf")},
		{"a described inline SOUR becomes a record with TITL",
			lines("0 @S1@ SOUR", "1 TITL first", "0 @I1@ INDI", "1 SOUR My book", "2 PAGE 3"), "",
			lines("0 @S1@ SOUR", "1 TITL first", "0 @I1@ INDI", "1 SOUR @S2@", "2 PAGE 3", "0 @S2@ SOUR", "1 TITL My book")},
		{"dates re-spelt keep the original in PHRASE; 7 spellings pass",
			lines("0 @I1@ INDI", "1 BIRT", "2 DATE @#DJULIAN@ 1 JAN 1700", "1 BIRT", "2 DATE 100 B.C.", "1 BIRT", "2 DATE 15 APR 1699/00",
				"1 BIRT", "2 DATE INT 1850 (war)", "1 BIRT", "2 DATE (spring)", "1 BIRT", "2 DATE 5  1801", "1 BIRT", "2 DATE @#DFRENCH R@ 1 VEND 8",
				"1 BIRT", "2 DATE @#DROMAN@ 71", "1 BIRT", "2 DATE 11 MAR 1961", "1 BIRT", "2 DATE ABT 1850", "1 BIRT", "2 DATE BET 1850 AND 1860", "1 BIRT", "2 DATE"), "",
			lines("0 @I1@ INDI", "1 BIRT", "2 DATE JULIAN 1 JAN 1700", "3 PHRASE @@#DJULIAN@ 1 JAN 1700", "1 BIRT", "2 DATE 100 BCE", "3 PHRASE 100 B.C.",
				"1 BIRT", "2 DATE 15 APR 1699", "3 PHRASE 15 APR 1699/00", "1 BIRT", "2 DATE 1850", "3 PHRASE INT 1850 (war)", "1 BIRT", "2 DATE", "3 PHRASE (spring)",
				"1 BIRT", "2 DATE", "3 PHRASE 5  1801", "1 BIRT", "2 DATE FRENCH_R 1 VEND 8", "3 PHRASE @@#DFRENCH R@ 1 VEND 8",
				"1 BIRT", "2 DATE _ROMAN 71", "3 PHRASE @@#DROMAN@ 71", "1 BIRT", "2 DATE 11 MAR 1961", "1 BIRT", "2 DATE ABT 1850", "1 BIRT", "2 DATE BET 1850 AND 1860", "1 BIRT", "2 DATE")},
		{"an exact date and a zero-padded day pass",
			lines("0 @I1@ INDI", "1 CHAN", "2 DATE 02 NOV 2023", "3 TIME 16:01:35"), "",
			lines("0 @I1@ INDI", "1 CHAN", "2 DATE 02 NOV 2023", "3 TIME 16:01:35")},
		{"age words and forms 7 rejects keep the original in PHRASE",
			lines("0 @I1@ INDI", "1 DEAT", "2 AGE CHILD", "1 DEAT", "2 AGE INFANT", "1 DEAT", "2 AGE STILLBORN", "1 DEAT", "2 AGE twenty", "1 DEAT", "2 AGE > 25y 3m"), "",
			lines("0 @I1@ INDI", "1 DEAT", "2 AGE < 8y", "3 PHRASE CHILD", "1 DEAT", "2 AGE < 1y", "3 PHRASE INFANT", "1 DEAT", "2 AGE 0y", "3 PHRASE STILLBORN",
				"1 DEAT", "2 AGE", "3 PHRASE twenty", "1 DEAT", "2 AGE > 25y 3m")},
		{"a NAME 7 rejects is repaired with _PHRASE",
			lines("0 @I1@ INDI", "1 NAME Louise/Louisa", "2 GIVN Louise", "1 NAME A /B/ C /D/", "1 NAME A\tB", "1 NAME John /Smith/"),
			"2 TAG _PHRASE " + v7 + "PHRASE\n",
			lines("0 @I1@ INDI", "1 NAME Louise Louisa", "2 GIVN Louise", "2 _PHRASE Louise/Louisa", "1 NAME A B C D", "2 _PHRASE A /B/ C /D/", "1 NAME A B", "2 _PHRASE A\tB", "1 NAME John /Smith/")},
		{"FONE and ROMN become TRAN with LANG; an unknown TYPE stays beside LANG und",
			lines("0 @I1@ INDI", "1 NAME 橘 /逸勢/", "2 ROMN Tachibana /Hayanari/", "3 TYPE romaji", "3 SURN Tachibana", "2 FONE たちばな /はやなり/", "3 TYPE kana", "2 ROMN x", "3 TYPE other",
				"1 BIRT", "2 PLAC 京都", "3 ROMN Kyoto", "4 TYPE romaji"),
			"2 TAG _TYPE " + v7 + "TYPE\n",
			lines("0 @I1@ INDI", "1 NAME 橘 /逸勢/", "2 TRAN Tachibana /Hayanari/", "3 LANG ja-Latn", "3 SURN Tachibana", "2 TRAN たちばな /はやなり/", "3 LANG ja-hrkt", "2 TRAN x", "3 LANG und", "3 _TYPE other",
				"1 BIRT", "2 PLAC 京都", "3 TRAN Kyoto", "4 LANG ja-Latn")},
		{"an empty ADDR payload is composed from the pieces; one with no pieces stays empty",
			lines("0 @I1@ INDI", "1 BIRT", "2 ADDR", "3 CITY Errol", "3 CTRY Scotland", "1 DEAT", "2 ADDR", "1 BURI", "2 ADDR Here", "3 CITY There"), "",
			lines("0 @I1@ INDI", "1 BIRT", "2 ADDR Errol", "3 CONT Scotland", "3 CITY Errol", "3 CTRY Scotland", "1 DEAT", "2 ADDR", "1 BURI", "2 ADDR Here", "3 CITY There")},
		{"a standard tag out of place is relocated and declared in SCHMA",
			lines("0 @I1@ INDI", "1 ADDR x@y", "2 EMAIL x@y", "2 PHON 1", "2 CITY c", "1 NAME a", "2 _MARNM Smith"),
			"2 TAG _ADDR " + v7 + "ADDR\n2 TAG _EMAIL " + v7 + "EMAIL\n2 TAG _PHON " + v7 + "PHON\n",
			lines("0 @I1@ INDI", "1 _ADDR x@y", "2 _EMAIL x@y", "2 _PHON 1", "2 CITY c", "1 NAME a", "2 _MARNM Smith")},
		{"an extension tag passes with everything under it",
			lines("0 @I1@ INDI", "1 _UID x", "2 DATE 5  1801", "2 WHATEVER y"), "",
			lines("0 @I1@ INDI", "1 _UID x", "2 DATE 5  1801", "2 WHATEVER y")},
		{"a SUBN record is dropped",
			lines("0 @U1@ SUBN", "1 SUBM @S1@", "1 ANCE 3", "0 @I1@ INDI", "1 SEX F"), "",
			lines("0 @I1@ INDI", "1 SEX F")},
		{"what the registry rejects but the grammar accepts is unchanged",
			lines("0 @I1@ INDI", "1 OCCU", "1 SEX M", "1 SEX F", "1 SUBM @S1@"), "",
			lines("0 @I1@ INDI", "1 OCCU", "1 SEX M", "1 SEX F", "1 SUBM @S1@")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := to7(t, head551+c.in+"0 TRLR\n")
			want := head7
			if c.schma != "" {
				want += "1 SCHMA\n" + c.schma
			}
			want += c.want + "0 TRLR\n"
			if got != want {
				t.Errorf("\n got %q\nwant %q", got, want)
			}
		})
	}
}

func TestTo7Header(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{"CHAR, GEDC.FORM, FILE and SUBN dropped; VERS 7.0; the rest kept",
			lines("0 HEAD", "1 SOUR Geni.com", "2 VERS 1.0", "1 DATE 24 OCT 2026", "2 TIME 05:15:00", "1 SUBM @S1@", "1 SUBN @U1@", "1 FILE x.ged", "1 GEDC", "2 VERS 5.5.1", "2 FORM LINEAGE-LINKED", "1 CHAR UTF-8", "1 LANG English", "0 TRLR"),
			lines("0 HEAD", "1 SOUR Geni.com", "2 VERS 1.0", "1 DATE 24 OCT 2026", "2 TIME 05:15:00", "1 SUBM @S1@", "1 GEDC", "2 VERS 7.0", "1 LANG English", "0 TRLR")},
		{"SCHMA follows GEDC",
			lines("0 HEAD", "1 GEDC", "2 VERS 5.5.1", "1 LANG English", "0 @I1@ INDI", "1 ADDR a", "0 TRLR"),
			lines("0 HEAD", "1 GEDC", "2 VERS 7.0", "1 SCHMA", "2 TAG _ADDR https://gedcom.io/terms/v7/ADDR", "1 LANG English", "0 @I1@ INDI", "1 _ADDR a", "0 TRLR")},
		{"a header without GEDC gets one",
			lines("0 HEAD", "0 TRLR"),
			lines("0 HEAD", "1 GEDC", "2 VERS 7.0", "0 TRLR")},
		{"RIN's TYPE carries the header's system",
			lines("0 HEAD", "1 SOUR MySystem", "1 GEDC", "2 VERS 5.5.1", "0 @I1@ INDI", "1 RIN 9876", "0 TRLR"),
			lines("0 HEAD", "1 SOUR MySystem", "1 GEDC", "2 VERS 7.0", "0 @I1@ INDI", "1 EXID 9876", "2 TYPE https://gedcom.io/terms/v7/RIN#MySystem", "0 TRLR")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := to7(t, c.src); got != c.want {
				t.Errorf("\n got %q\nwant %q", got, c.want)
			}
		})
	}
}

func TestTo7Errors(t *testing.T) {
	cases := []struct {
		name string
		in   string
		err  string
	}{
		{"an xref 7 rejects", "0 @I 1@ INDI\n", "line 6"},
		{"a tag 7 rejects", "0 @I1@ INDI\n1 Note x\n", "line 7"},
		{"a tag 7 rejects under an extension", "0 @I1@ INDI\n1 _UID x\n2 Note x\n", "line 8"},
		{"a pointer 7 rejects", "0 @I1@ INDI\n1 FAMC @F 1@\n", "line 7"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			doc, err := gedcom551.Read(strings.NewReader(head551 + c.in + "0 TRLR\n"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = To7(doc)
			if err == nil || !strings.Contains(err.Error(), c.err) {
				t.Fatalf("To7: %v, want %q in the error", err, c.err)
			}
		})
	}
}

func TestTo7LeavesTheInput(t *testing.T) {
	src := head551 + lines("0 @I1@ INDI", "1 NAME Louise/Louisa", "1 OBJE", "2 FILE f", "3 FORM jpg", "1 ADDR a", "2 EMAIL e", "1 BIRT", "2 DATE 100 B.C.", "2 ADDR", "3 CITY c") + "0 TRLR\n"
	doc, err := gedcom551.Read(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	before, _ := gedcom551.Read(strings.NewReader(src))
	if _, err := To7(doc); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(doc, before) {
		t.Error("To7 changed its input")
	}
}

// TestTo7ReadsBack: what To7 makes, gedcom7 writes and reads back the same
// in every field but Line.
func TestTo7ReadsBack(t *testing.T) {
	src := head551 + lines("0 @N1@ NOTE a", "0 @I1@ INDI", "1 NAME Louise/Louisa", "1 NOTE @N1@", "1 RFN geni:1", "1 OBJE", "2 FILE f", "3 FORM jpg", "2 TITL t", "1 ADDR a", "2 EMAIL e",
		"1 BIRT", "2 DATE @#DJULIAN@ 1 JAN 1700", "2 ADDR", "3 CITY c", "1 SOUR", "2 DATA", "3 TEXT t") + "0 TRLR\n"
	doc, err := gedcom551.Read(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	out, err := To7(doc)
	if err != nil {
		t.Fatal(err)
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
		t.Errorf("read back differs:\n%s", buf.String())
	}
}

func clearLines(s *gedcom.Structure) {
	s.Line = 0
	for _, c := range s.Substructures {
		clearLines(c)
	}
}
