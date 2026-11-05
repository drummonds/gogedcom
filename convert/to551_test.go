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

// head7Schma is a 7 header of eight lines declaring 008's relocated tags,
// so a record's first line is line 9.
const head7Schma = "0 HEAD\n1 GEDC\n2 VERS 7.0\n1 SCHMA\n2 TAG _ADDR https://gedcom.io/terms/v7/ADDR\n2 TAG _EMAIL https://gedcom.io/terms/v7/EMAIL\n2 TAG _PHON https://gedcom.io/terms/v7/PHON\n2 TAG _PHRASE https://gedcom.io/terms/v7/PHRASE\n"

const head551out = "0 HEAD\n1 GEDC\n2 VERS 5.5.1\n2 FORM LINEAGE-LINKED\n1 CHAR UTF-8\n"

// to551 reads src as 7, converts it and writes it as 5.5.1.
func to551(t *testing.T, src string) string {
	t.Helper()
	doc, err := gedcom7.Read(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	out := To551(doc)
	var buf bytes.Buffer
	if err := gedcom551.Write(&buf, out); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestTo551(t *testing.T) {
	const v7 = "https://gedcom.io/terms/v7/"
	cases := []struct {
		name string
		in   string // 7 record lines
		want string // 5.5.1 record lines
	}{
		{"SNOTE record and pointer become NOTE",
			lines("0 @N1@ SNOTE a", "0 @I1@ INDI", "1 SNOTE @N1@", "1 NOTE inline"),
			lines("0 @N1@ NOTE a", "0 @I1@ INDI", "1 NOTE @N1@", "1 NOTE inline")},
		{"EXID from RFN, AFN, RIN comes back; another EXID is an extension",
			lines("0 @I1@ INDI", "1 EXID 42", "2 TYPE "+v7+"RFN#geni", "1 EXID 1", "2 TYPE "+v7+"AFN", "1 EXID 9", "2 TYPE "+v7+"RIN#Geni.com", "1 EXID 7", "2 TYPE "+v7+"RFN", "1 EXID x", "2 TYPE https://example.com/id"),
			lines("0 @I1@ INDI", "1 RFN geni:42", "1 AFN 1", "1 RIN 9", "1 RFN 7", "1 _EXID x", "2 TYPE https://example.com/id")},
		{"TRAN with a LANG in the guide's table becomes ROMN or FONE; another is an extension",
			lines("0 @I1@ INDI", "1 NAME 橘 /逸勢/", "2 TRAN Tachibana /Hayanari/", "3 LANG ja-Latn", "3 SURN Tachibana", "2 TRAN たちばな /はやなり/", "3 LANG ja-hrkt", "2 TRAN x", "3 LANG fr",
				"1 BIRT", "2 PLAC 京都", "3 TRAN Kyoto", "4 LANG ja-Latn"),
			lines("0 @I1@ INDI", "1 NAME 橘 /逸勢/", "2 ROMN Tachibana /Hayanari/", "3 TYPE romaji", "3 SURN Tachibana", "2 FONE たちばな /はやなり/", "3 TYPE kana", "2 _TRAN x", "3 LANG fr",
				"1 BIRT", "2 PLAC 京都", "3 ROMN Kyoto", "4 TYPE romaji")},
		{"ROLE OTHER with PHRASE becomes RELA the phrase; a set word stays; another PHRASE is an extension",
			lines("0 @I1@ INDI", "1 ASSO @I2@", "2 ROLE OTHER", "3 PHRASE Honorary uncle", "1 ASSO @I3@", "2 ROLE GODP", "1 SOUR @S1@", "2 EVEN BIRT", "3 ROLE OTHER", "4 PHRASE (Godfather)", "1 SOUR @S1@", "2 EVEN BIRT", "3 ROLE CHIL", "4 PHRASE eldest"),
			lines("0 @I1@ INDI", "1 ASSO @I2@", "2 RELA Honorary uncle", "1 ASSO @I3@", "2 RELA GODP", "1 SOUR @S1@", "2 EVEN BIRT", "3 ROLE (Godfather)", "1 SOUR @S1@", "2 EVEN BIRT", "3 ROLE CHIL", "4 _PHRASE eldest")},
		{"INIL becomes WAC",
			lines("0 @I1@ INDI", "1 INIL", "2 DATE 1 JAN 1990"),
			lines("0 @I1@ INDI", "1 WAC", "2 DATE 1 JAN 1990")},
		{"enumerations in 5.5.1's spelling; a value outside 7's set unchanged",
			lines("0 @I1@ INDI", "1 SEX M", "1 SEX X", "1 RESN PRIVACY", "1 RESN CONFIDENTIAL, LOCKED", "1 FAMC @F1@", "2 PEDI BIRTH", "2 STAT CHALLENGED", "1 FAMC @F2@", "2 PEDI step",
				"1 BAPL", "2 STAT DNS_CAN", "1 ENDL", "2 STAT PRE_1970", "1 ADOP", "2 FAMC @F1@", "3 ADOP BOTH", "1 NAME a", "2 TYPE MAIDEN", "1 SOUR @S1@", "2 QUAY 3"),
			lines("0 @I1@ INDI", "1 SEX M", "1 SEX X", "1 RESN privacy", "1 RESN confidential, locked", "1 FAMC @F1@", "2 PEDI birth", "2 STAT challenged", "1 FAMC @F2@", "2 PEDI step",
				"1 BAPL", "2 STAT DNS/CAN", "1 ENDL", "2 STAT PRE-1970", "1 ADOP", "2 FAMC @F1@", "3 ADOP BOTH", "1 NAME a", "2 TYPE maiden", "1 SOUR @S1@", "2 QUAY 3")},
		{"an OBJE record pointed to once is inlined: FILE, its TITL, the pointer's substructures, the rest",
			lines("0 @I1@ INDI", "1 OBJE @O1@", "2 _PRIM Y", "0 @O1@ OBJE", "1 FILE a.jpg", "2 FORM image/jpeg", "3 MEDI PHOTO", "2 TITL t", "1 NOTE n", "1 NOTE m"),
			lines("0 @I1@ INDI", "1 OBJE", "2 FILE a.jpg", "3 FORM jpg", "4 MEDI photo", "2 TITL t", "2 _PRIM Y", "2 NOTE n", "2 NOTE m")},
		{"an OBJE record pointed to twice or never stays, MEDI as TYPE; a media type outside the table unchanged",
			lines("0 @O1@ OBJE", "1 FILE f", "2 FORM image/png", "3 MEDI PHOTO", "0 @O2@ OBJE", "1 FILE g", "2 FORM image/x-xcf", "0 @I1@ INDI", "1 OBJE @O1@", "1 OBJE @O1@"),
			lines("0 @O1@ OBJE", "1 FILE f", "2 FORM png", "3 TYPE photo", "0 @O2@ OBJE", "1 FILE g", "2 FORM image/x-xcf", "0 @I1@ INDI", "1 OBJE @O1@", "1 OBJE @O1@")},
		{"an OBJE record named once by something other than an OBJE stays",
			lines("0 @O1@ OBJE", "1 FILE f", "2 FORM image/png", "0 @I1@ INDI", "1 _MEDIA @O1@"),
			lines("0 @O1@ OBJE", "1 FILE f", "2 FORM png", "0 @I1@ INDI", "1 _MEDIA @O1@")},
		{"SOUR @VOID@ becomes an inline SOUR; a pointer citation stays",
			lines("0 @I1@ INDI", "1 SOUR @VOID@", "2 DATA", "3 DATE 1 JAN 1990", "3 TEXT t", "2 OBJE @O1@", "1 SOUR @S1@", "2 PAGE 3", "0 @O1@ OBJE", "1 FILE f", "2 FORM application/pdf"),
			lines("0 @I1@ INDI", "1 SOUR", "2 DATA", "3 DATE 1 JAN 1990", "3 TEXT t", "2 OBJE", "3 FILE f", "4 FORM pdf", "1 SOUR @S1@", "2 PAGE 3")},
		{"a PHRASE 008 would have written is the DATE",
			lines("0 @I1@ INDI", "1 BIRT", "2 DATE JULIAN 1 JAN 1700", "3 PHRASE @@#DJULIAN@ 1 JAN 1700", "1 BIRT", "2 DATE", "3 PHRASE 5  1801", "1 BIRT", "2 DATE 1850", "3 PHRASE INT 1850 (war)",
				"1 BIRT", "2 DATE", "3 PHRASE (spring)", "1 BIRT", "2 DATE 100 BCE", "3 PHRASE 100 B.C.", "1 BIRT", "2 DATE 15 APR 1699", "3 PHRASE 15 APR 1699/00"),
			lines("0 @I1@ INDI", "1 BIRT", "2 DATE @#DJULIAN@ 1 JAN 1700", "1 BIRT", "2 DATE 5  1801", "1 BIRT", "2 DATE INT 1850 (war)",
				"1 BIRT", "2 DATE (spring)", "1 BIRT", "2 DATE 100 B.C.", "1 BIRT", "2 DATE 15 APR 1699/00")},
		{"a 7-native PHRASE takes 5.5.1's phrase form; under an empty DATE it is the payload, as 008 wrote it",
			lines("0 @I1@ INDI", "1 BIRT", "2 DATE 1850", "3 PHRASE about the war", "1 BIRT", "2 DATE", "3 PHRASE sometime", "1 BIRT", "2 DATE JULIAN 1700", "3 PHRASE old style"),
			lines("0 @I1@ INDI", "1 BIRT", "2 DATE INT 1850 (about the war)", "1 BIRT", "2 DATE sometime", "1 BIRT", "2 DATE INT @#DJULIAN@ 1700 (old style)")},
		{"a DATE 5.5.1 rejects is re-spelt; one it accepts passes",
			lines("0 @I1@ INDI", "1 BIRT", "2 DATE JULIAN 1 JAN 1700", "1 BIRT", "2 DATE 100 BCE", "1 BIRT", "2 DATE FRENCH_R 1 VEND 8", "1 BIRT", "2 DATE _ROMAN 71", "1 BIRT", "2 DATE BET JULIAN 1700 AND 1800",
				"1 BIRT", "2 DATE FROM 1700 TO GREGORIAN 1800", "1 BIRT", "2 DATE ABT HEBREW 5000", "1 BIRT", "2 DATE BET 1850 AND 1860", "1 CHAN", "2 DATE 02 APR 2023", "3 TIME 10:00"),
			lines("0 @I1@ INDI", "1 BIRT", "2 DATE @#DJULIAN@ 1 JAN 1700", "1 BIRT", "2 DATE 100 B.C.", "1 BIRT", "2 DATE @#DFRENCH R@ 1 VEND 8", "1 BIRT", "2 DATE @#DROMAN@ 71", "1 BIRT", "2 DATE BET @#DJULIAN@ 1700 AND 1800",
				"1 BIRT", "2 DATE FROM 1700 TO @#DGREGORIAN@ 1800", "1 BIRT", "2 DATE ABT @#DHEBREW@ 5000", "1 BIRT", "2 DATE BET 1850 AND 1860", "1 CHAN", "2 DATE 02 APR 2023", "3 TIME 10:00")},
		{"an AGE PHRASE 008 would have written is the AGE; another is an extension",
			lines("0 @I1@ INDI", "1 DEAT", "2 AGE < 8y", "3 PHRASE CHILD", "1 DEAT", "2 AGE", "3 PHRASE twenty", "1 DEAT", "2 AGE 25y", "3 PHRASE about", "1 DEAT", "2 AGE > 25y 3m"),
			lines("0 @I1@ INDI", "1 DEAT", "2 AGE CHILD", "1 DEAT", "2 AGE twenty", "1 DEAT", "2 AGE 25y", "3 _PHRASE about", "1 DEAT", "2 AGE > 25y 3m")},
		{"a NAME's relocated PHRASE is the NAME",
			lines("0 @I1@ INDI", "1 NAME Louise Louisa", "2 GIVN Louise", "2 _PHRASE Louise/Louisa", "1 NAME A B", "2 _PHRASE A\tB", "1 NAME John /Smith/"),
			lines("0 @I1@ INDI", "1 NAME Louise/Louisa", "2 GIVN Louise", "1 NAME A\tB", "1 NAME John /Smith/")},
		{"an ADDR payload equal to its pieces is dropped; another stays",
			lines("0 @I1@ INDI", "1 BIRT", "2 ADDR Errol", "3 CONT Scotland", "3 CITY Errol", "3 CTRY Scotland", "1 DEAT", "2 ADDR", "1 BURI", "2 ADDR Here", "3 CITY There"),
			lines("0 @I1@ INDI", "1 BIRT", "2 ADDR", "3 CITY Errol", "3 CTRY Scotland", "1 DEAT", "2 ADDR", "1 BURI", "2 ADDR Here", "3 CITY There")},
		{"a relocated tag comes back; an undeclared extension passes with everything under it",
			lines("0 @I1@ INDI", "1 _ADDR x@y", "2 _EMAIL x@y", "2 _PHON 1", "2 CITY c", "1 NAME a", "2 _MARNM Smith", "1 _UID x", "2 DATE JULIAN 1700", "2 PHRASE p"),
			lines("0 @I1@ INDI", "1 ADDR x@y", "2 EMAIL x@y", "2 PHON 1", "2 CITY c", "1 NAME a", "2 _MARNM Smith", "1 _UID x", "2 DATE JULIAN 1700", "2 PHRASE p")},
		{"7-native tags become extensions with everything under them; a void pointer stays",
			lines("0 @I1@ INDI", "1 CREA", "2 DATE 1 JAN 2020", "1 UID 1234", "1 NO MARR", "2 DATE TO 1850", "1 FAMC @VOID@", "2 PEDI ADOPTED", "3 PHRASE by an aunt", "1 BIRT", "2 SDATE 1850"),
			lines("0 @I1@ INDI", "1 _CREA", "2 DATE 1 JAN 2020", "1 _UID 1234", "1 _NO MARR", "2 DATE TO 1850", "1 FAMC @VOID@", "2 PEDI adopted", "3 _PHRASE by an aunt", "1 BIRT", "2 _SDATE 1850")},
		{"what 5.5.1's schema would reject is unchanged",
			lines("0 @I1@ INDI", "1 OCCU", "1 SEX M", "1 SEX F", "1 SUBM @S1@"),
			lines("0 @I1@ INDI", "1 OCCU", "1 SEX M", "1 SEX F", "1 SUBM @S1@")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := to551(t, head7Schma+c.in+"0 TRLR\n")
			if want := head551out + c.want + "0 TRLR\n"; got != want {
				t.Errorf("\n got %q\nwant %q", got, want)
			}
		})
	}
}

func TestTo551Header(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{"VERS 5.5.1, FORM and CHAR added in place, SCHMA consumed, the rest kept",
			lines("0 HEAD", "1 SOUR Geni.com", "2 VERS 1.0", "1 DATE 24 OCT 2026", "2 TIME 05:15:00", "1 SUBM @S1@", "1 GEDC", "2 VERS 7.0", "1 SCHMA", "2 TAG _ADDR https://gedcom.io/terms/v7/ADDR", "1 LANG English", "0 TRLR"),
			lines("0 HEAD", "1 SOUR Geni.com", "2 VERS 1.0", "1 DATE 24 OCT 2026", "2 TIME 05:15:00", "1 SUBM @S1@", "1 GEDC", "2 VERS 5.5.1", "2 FORM LINEAGE-LINKED", "1 CHAR UTF-8", "1 LANG English", "0 TRLR")},
		{"a header with its own CHAR and FORM keeps them",
			lines("0 HEAD", "1 GEDC", "2 VERS 7.0", "2 FORM LINEAGE-LINKED", "1 CHAR ANSEL", "0 TRLR"),
			lines("0 HEAD", "1 GEDC", "2 VERS 5.5.1", "2 FORM LINEAGE-LINKED", "1 CHAR ANSEL", "0 TRLR")},
		{"a header without GEDC gets one",
			lines("0 HEAD", "0 TRLR"),
			lines("0 HEAD", "1 GEDC", "2 VERS 5.5.1", "2 FORM LINEAGE-LINKED", "1 CHAR UTF-8", "0 TRLR")},
		{"a HEAD SNOTE pointer becomes NOTE",
			lines("0 HEAD", "1 GEDC", "2 VERS 7.0", "1 SNOTE @N1@", "0 @N1@ SNOTE x", "0 TRLR"),
			lines("0 HEAD", "1 GEDC", "2 VERS 5.5.1", "2 FORM LINEAGE-LINKED", "1 CHAR UTF-8", "1 NOTE @N1@", "0 @N1@ NOTE x", "0 TRLR")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := to551(t, c.src); got != c.want {
				t.Errorf("\n got %q\nwant %q", got, c.want)
			}
		})
	}
}

func TestTo551LeavesTheInput(t *testing.T) {
	src := head7Schma + lines("0 @I1@ INDI", "1 NAME Louise Louisa", "2 _PHRASE Louise/Louisa", "1 OBJE @O1@", "2 _PRIM Y", "1 _ADDR a", "2 _EMAIL e", "1 BIRT", "2 DATE 100 BCE", "2 ADDR c", "3 CITY c",
		"1 EXID 1", "2 TYPE https://gedcom.io/terms/v7/RFN#x", "0 @O1@ OBJE", "1 FILE f", "2 FORM image/jpeg", "2 TITL t") + "0 TRLR\n"
	doc, err := gedcom7.Read(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	before, _ := gedcom7.Read(strings.NewReader(src))
	To551(doc)
	if !reflect.DeepEqual(doc, before) {
		t.Error("To551 changed its input")
	}
}

// TestTo551Inverts: what To7 makes of 5.5.1 lines, To551 gives back.
func TestTo551Inverts(t *testing.T) {
	src := head551 + lines("0 @N1@ NOTE a", "0 @I1@ INDI", "1 NAME Louise/Louisa", "2 GIVN Louise", "1 NOTE @N1@", "1 RFN geni:1", "1 RIN 9", "1 OBJE", "2 FILE f", "3 FORM jpg", "2 TITL t", "2 _PRIM Y", "2 NOTE n",
		"1 ADDR a", "2 EMAIL e", "2 PHON 1", "1 BIRT", "2 DATE @#DJULIAN@ 1 JAN 1700", "2 ADDR", "3 CITY c", "3 CTRY d", "2 AGE CHILD", "1 DEAT", "2 DATE 5  1801", "2 AGE twenty", "1 SOUR", "2 DATA", "3 TEXT t", "2 OBJE", "3 FILE g", "4 FORM pdf",
		"1 ASSO @I2@", "2 RELA Honorary uncle", "1 FAMC @F1@", "2 PEDI birth", "1 RESN privacy", "1 WAC", "2 DATE 1 JAN 1990") + "0 TRLR\n"
	in, err := gedcom551.Read(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	seven, err := To7(in)
	if err != nil {
		t.Fatal(err)
	}
	back := To551(seven)
	for _, d := range []*gedcom.Document{in, back} {
		clearLines(d.Header)
		for _, r := range d.Records {
			clearLines(r)
		}
	}
	if !reflect.DeepEqual(in, back) {
		var buf bytes.Buffer
		gedcom551.Write(&buf, back)
		t.Errorf("the trip changed the tree:\n%s", buf.String())
	}
}
