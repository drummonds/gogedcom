package gedcom7

import (
	"slices"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	const head = "0 HEAD\n1 GEDC\n2 VERS 7.0\n"
	cases := []struct {
		name string
		src  string
		want []string
	}{
		{"valid", head + "0 @I1@ INDI\n1 SEX M\n1 RESN CONFIDENTIAL, LOCKED\n0 TRLR\n", nil},
		{"not a record type", head + "0 BIRT\n0 TRLR\n", []string{"line 4: BIRT is not a record type"}},
		{"not a substructure, no cascade", head + "0 @S1@ SOUR\n1 BIRT\n2 DATE 1 JAN 1900\n0 TRLR\n",
			[]string{"line 5: BIRT is not a substructure of record-SOUR"}},
		{"extension unchecked", head + "0 @I1@ INDI\n1 _UID x\n2 WHATEVER y\n0 TRLR\n", nil},
		{"singleton repeated", head + "0 @I1@ INDI\n1 SEX M\n1 SEX F\n0 TRLR\n",
			[]string{"line 6: record-INDI has more than one SEX"}},
		{"required missing", "0 HEAD\n1 GEDC\n0 TRLR\n", []string{"line 2: GEDC lacks VERS"}},
		{"required missing at the superstructure", "0 HEAD\n0 TRLR\n", []string{"line 1: HEAD lacks GEDC"}},
		{"enumeration", head + "0 @I1@ INDI\n1 SEX Q\n0 TRLR\n", []string{"line 5: Q is not in enumset-SEX"}},
		{"enumeration without a payload", head + "0 @I1@ INDI\n1 SEX\n0 TRLR\n", []string{"line 5: SEX has no payload"}},
		{"enumeration list", head + "0 @I1@ INDI\n1 RESN CONFIDENTIAL, BOGUS\n0 TRLR\n",
			[]string{"line 5: BOGUS is not in enumset-RESN"}},
		{"in line order", head + "0 @I1@ INDI\n1 SEX Q\n0 @S1@ SOUR\n1 BIRT\n0 TRLR\n",
			[]string{"line 5: Q is not in enumset-SEX", "line 7: BIRT is not a substructure of record-SOUR"}},
	}
	r := Standard()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			doc, err := Read(strings.NewReader(c.src))
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, v := range r.Validate(doc) {
				got = append(got, v.String())
			}
			if !slices.Equal(got, c.want) {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}
