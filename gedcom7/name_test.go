package gedcom7

import "testing"

func TestParsePersonalName(t *testing.T) {
	cases := []struct {
		in   string
		want PersonalName
	}{
		{"John /Smith/", PersonalName{Given: "John", Surname: "Smith"}},
		{"John /Smith/ Jr.", PersonalName{Given: "John", Surname: "Smith", Suffix: "Jr."}},
		{"/Smith/", PersonalName{Surname: "Smith"}},
		{"John", PersonalName{Given: "John"}},
		{"Lt. Cmndr. Joseph /Allen/ jr.", PersonalName{Given: "Lt. Cmndr. Joseph", Surname: "Allen", Suffix: "jr."}},
		{"José /García Márquez/", PersonalName{Given: "José", Surname: "García Márquez"}},
		{"John //", PersonalName{Given: "John"}},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got, err := ParsePersonalName(c.in)
			if err != nil {
				t.Fatalf("ParsePersonalName(%q): %v", c.in, err)
			}
			if got != c.want {
				t.Fatalf("ParsePersonalName(%q) = %+v, want %+v", c.in, got, c.want)
			}
			if again, err := ParsePersonalName(got.String()); err != nil || again != got {
				t.Fatalf("String() = %q reads as %+v, %v", got.String(), again, err)
			}
		})
	}
	for _, in := range []string{"", "John /Smith", "John /Smith/ /Jones/", "John\t/Smith/"} {
		if _, err := ParsePersonalName(in); err == nil {
			t.Errorf("ParsePersonalName(%q) should fail", in)
		}
	}
}
