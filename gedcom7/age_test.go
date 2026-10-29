package gedcom7

import "testing"

func TestParseAge(t *testing.T) {
	cases := []struct {
		in   string
		want Age
	}{
		{"", Age{}},
		{"35y", Age{Years: 35}},
		{"> 35y 2m", Age{Bound: ">", Years: 35, Months: 2}},
		{"< 8w 3d", Age{Bound: "<", Weeks: 8, Days: 3}},
		{"1y 2m 3w 4d", Age{Years: 1, Months: 2, Weeks: 3, Days: 4}},
		{"3d", Age{Days: 3}},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got, err := ParseAge(c.in)
			if err != nil {
				t.Fatalf("ParseAge(%q): %v", c.in, err)
			}
			if got != c.want {
				t.Fatalf("ParseAge(%q) = %+v, want %+v", c.in, got, c.want)
			}
			if again, err := ParseAge(got.String()); err != nil || again != got {
				t.Fatalf("String() = %q reads as %+v, %v", got.String(), again, err)
			}
		})
	}
	for _, in := range []string{"35", "<35y", "2m 35y", "35Y", "35y2m"} {
		if _, err := ParseAge(in); err == nil {
			t.Errorf("ParseAge(%q) should fail", in)
		}
	}
}
