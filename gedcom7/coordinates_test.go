package gedcom7

import "testing"

func TestParseLatitude(t *testing.T) {
	cases := map[string]Latitude{"N51.5": 51.5, "S33.86": -33.86, "N90": 90, "S0": 0}
	for in, want := range cases {
		got, err := ParseLatitude(in)
		if err != nil {
			t.Errorf("ParseLatitude(%q): %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("ParseLatitude(%q) = %v, want %v", in, got, want)
		}
		if again, err := ParseLatitude(got.String()); err != nil || again != got {
			t.Errorf("String() = %q reads as %v, %v", got.String(), again, err)
		}
	}
	for _, in := range []string{"N90.1", "N91", "51.5", "E51.5", "n51.5", "N51."} {
		if _, err := ParseLatitude(in); err == nil {
			t.Errorf("ParseLatitude(%q) should fail", in)
		}
	}
}

func TestParseLongitude(t *testing.T) {
	cases := map[string]Longitude{"W1.5": -1.5, "E180": 180, "E179.99": 179.99, "W0.000001": -0.000001}
	for in, want := range cases {
		got, err := ParseLongitude(in)
		if err != nil {
			t.Errorf("ParseLongitude(%q): %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("ParseLongitude(%q) = %v, want %v", in, got, want)
		}
		if again, err := ParseLongitude(got.String()); err != nil || again != got {
			t.Errorf("String() = %q reads as %v, %v", got.String(), again, err)
		}
	}
	for _, in := range []string{"E180.1", "E181", "N1.5", "1.5"} {
		if _, err := ParseLongitude(in); err == nil {
			t.Errorf("ParseLongitude(%q) should fail", in)
		}
	}
}
