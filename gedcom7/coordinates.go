package gedcom7

import (
	"fmt"
	"math"
	"strconv"
)

// Latitude is the spec's Latitude rule as signed degrees, north positive,
// within ±90.
type Latitude float64

// Longitude is the spec's Longitude rule as signed degrees, east positive,
// within ±180.
type Longitude float64

// String is the latitude as GEDCOM writes it: N or S and the degrees.
func (l Latitude) String() string { return hemisphere(float64(l), "N", "S") }

// String is the longitude as GEDCOM writes it: E or W and the degrees.
func (l Longitude) String() string { return hemisphere(float64(l), "E", "W") }

func hemisphere(degrees float64, positive, negative string) string {
	letter := positive
	if degrees < 0 {
		letter = negative
	}
	return letter + strconv.FormatFloat(math.Abs(degrees), 'f', -1, 64)
}

// ParseLatitude reads a Latitude payload.
func ParseLatitude(s string) (Latitude, error) {
	d, err := parseDegrees("latitude", "LATI", s, 'S', 90)
	return Latitude(d), err
}

// ParseLongitude reads a Longitude payload.
func ParseLongitude(s string) (Longitude, error) {
	d, err := parseDegrees("longitude", "LONG", s, 'W', 180)
	return Longitude(d), err
}

// parseDegrees reads a hemisphere letter and degrees; negative is the
// letter of the negative hemisphere. The grammar bounds the whole degrees
// but not the fraction (N90.5 is grammatical), so the bound is checked
// here, as ParseDate checks the day.
func parseDegrees(what, tag, s string, negative byte, limit float64) (float64, error) {
	p, err := parsePayload(what, tag, s)
	if err != nil {
		return 0, err
	}
	defer p.release()
	text := p.text(p.value(what))
	d, _ := strconv.ParseFloat(text[1:], 64)
	if d > limit {
		return 0, fmt.Errorf("gedcom7: %s %q: over %v", what, s, limit)
	}
	if text[0] == negative && d != 0 {
		d = -d
	}
	return d, nil
}
