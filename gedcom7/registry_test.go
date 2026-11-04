package gedcom7

import (
	"slices"
	"testing"
)

const v7 = "https://gedcom.io/terms/v7/"

func TestStandardRegistry(t *testing.T) {
	r := Standard()
	indi := r.Structure(v7 + "record-INDI")
	if indi == nil || indi.Tag != "INDI" || len(indi.Superstructures) != 0 {
		t.Fatalf("record-INDI = %+v", indi)
	}
	if c := indi.Substructures[v7+"SEX"]; c != (Cardinality{}) {
		t.Errorf("SEX under INDI = %v, want {0:1}", c)
	}
	if c := indi.Substructures[v7+"INDI-FAMC"]; c != (Cardinality{Repeatable: true}) {
		t.Errorf("FAMC under INDI = %v, want {0:M}", c)
	}
	if c := r.Structure(v7 + "HEAD").Substructures[v7+"GEDC"]; c != (Cardinality{Required: true}) {
		t.Errorf("GEDC under HEAD = %v, want {1:1}", c)
	}
	for _, c := range []struct {
		super *StructureType
		tag   string
		want  string
	}{
		{nil, "INDI", v7 + "record-INDI"},
		{nil, "HEAD", v7 + "HEAD"},
		{nil, "TRLR", v7 + "TRLR"},
		{nil, "BIRT", ""},
		{indi, "FAMC", v7 + "INDI-FAMC"},
		{r.Structure(v7 + "ADOP"), "FAMC", v7 + "ADOP-FAMC"},
		{r.Structure(v7 + "CHAN"), "DATE", v7 + "DATE-exact"},
		{indi, "ADDR", ""},
	} {
		got := r.Substructure(c.super, c.tag)
		switch {
		case c.want == "" && got != nil:
			t.Errorf("Substructure(%v, %s) = %s, want nil", c.super, c.tag, got.URI)
		case c.want != "" && (got == nil || got.URI != c.want):
			t.Errorf("Substructure(%v, %s) = %v, want %s", c.super, c.tag, got, c.want)
		}
	}
	if p := r.Substructure(r.Structure(v7+"CHAN"), "DATE").Payload; p != v7+"type-Date#exact" {
		t.Errorf("DATE-exact payload = %s", p)
	}
	if s := r.Structure(v7 + "SEX").EnumerationSet; s != v7+"enumset-SEX" {
		t.Errorf("SEX enumeration set = %s", s)
	}
	if got := r.Enumeration(v7 + "enumset-SEX"); !slices.Equal(got, []string{"M", "F", "X", "U"}) {
		t.Errorf("enumset-SEX = %v", got)
	}
	even := r.Enumeration(v7 + "enumset-EVEN")
	if !slices.Contains(even, "CENS") || !slices.Contains(even, "BIRT") {
		t.Errorf("enumset-EVEN = %v", even)
	}
}

func TestCardinalityString(t *testing.T) {
	for want, c := range map[string]Cardinality{
		"{0:1}": {}, "{0:M}": {Repeatable: true}, "{1:1}": {Required: true}, "{1:M}": {Required: true, Repeatable: true},
	} {
		if c.String() != want {
			t.Errorf("%+v = %s, want %s", c, c, want)
		}
	}
}

func TestTypes(t *testing.T) {
	r := Standard()
	var uris []string
	for _, s := range r.Types("DATE") {
		uris = append(uris, s.URI)
	}
	want := []string{v7 + "DATA-EVEN-DATE", v7 + "DATE", v7 + "DATE-exact", v7 + "HEAD-DATE", v7 + "NO-DATE"}
	if !slices.Equal(uris, want) {
		t.Errorf("Types(DATE) = %q, want %q", uris, want)
	}
	if r.Types("BOGUS") != nil {
		t.Error("Types(BOGUS) should be nil")
	}
}
