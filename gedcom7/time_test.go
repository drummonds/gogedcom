package gedcom7

import "testing"

func TestParseTime(t *testing.T) {
	cases := []struct {
		in   string
		want Time
	}{
		{"12:30", Time{Hour: 12, Minute: 30}},
		{"9:05", Time{Hour: 9, Minute: 5}},
		{"12:30:05", Time{Hour: 12, Minute: 30, Second: 5}},
		{"12:30:05.5Z", Time{Hour: 12, Minute: 30, Second: 5, Fraction: "5", UTC: true}},
		{"23:59:59.999", Time{Hour: 23, Minute: 59, Second: 59, Fraction: "999"}},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got, err := ParseTime(c.in)
			if err != nil {
				t.Fatalf("ParseTime(%q): %v", c.in, err)
			}
			if got != c.want {
				t.Fatalf("ParseTime(%q) = %+v, want %+v", c.in, got, c.want)
			}
			if again, err := ParseTime(got.String()); err != nil || again != got {
				t.Fatalf("String() = %q reads as %+v, %v", got.String(), again, err)
			}
		})
	}
	for _, in := range []string{"", "24:00", "12:60", "1230", "12:30z", "12:30:05."} {
		if _, err := ParseTime(in); err == nil {
			t.Errorf("ParseTime(%q) should fail", in)
		}
	}
}
