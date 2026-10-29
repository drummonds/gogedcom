package gedcom7

import (
	"fmt"
	"strconv"
)

// Time is the spec's Time rule: hour:minute[:second[.fraction]][Z]. Second
// is 0 and Fraction "" when absent; Fraction is the digits as written, so
// .5 and .50 differ; UTC is the Z.
type Time struct {
	Hour     int
	Minute   int
	Second   int
	Fraction string
	UTC      bool
}

// String is the time as GEDCOM writes it.
func (t Time) String() string {
	s := fmt.Sprintf("%02d:%02d", t.Hour, t.Minute)
	if t.Second != 0 || t.Fraction != "" {
		s += fmt.Sprintf(":%02d", t.Second)
		if t.Fraction != "" {
			s += "." + t.Fraction
		}
	}
	if t.UTC {
		s += "Z"
	}
	return s
}

// ParseTime reads a Time payload.
func ParseTime(s string) (Time, error) {
	p, err := parsePayload("time", "TIME", s)
	if err != nil {
		return Time{}, err
	}
	defer p.release()
	n := p.value("time")
	t := Time{Fraction: p.childText(n, "fraction")}
	t.Hour, _ = strconv.Atoi(p.childText(n, "hour"))
	t.Minute, _ = strconv.Atoi(p.childText(n, "minute"))
	t.Second, _ = strconv.Atoi(p.childText(n, "second"))
	text := p.text(n)
	t.UTC = text[len(text)-1] == 'Z'
	return t, nil
}
