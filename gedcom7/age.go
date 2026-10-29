package gedcom7

import (
	"strconv"
	"strings"
)

// Age is the spec's Age rule: [bound] years months weeks days, each unit
// optional. Bound is "<", ">" or ""; a unit not given is 0. The zero Age is
// the empty payload.
type Age struct {
	Bound  string
	Years  int
	Months int
	Weeks  int
	Days   int
}

// String is the age as GEDCOM writes it. An age of zero in every unit
// writes as "0y" when it has a bound and as the empty payload otherwise.
func (a Age) String() string {
	var parts []string
	if a.Bound != "" {
		parts = append(parts, a.Bound)
	}
	for _, u := range []struct {
		n    int
		unit string
	}{{a.Years, "y"}, {a.Months, "m"}, {a.Weeks, "w"}, {a.Days, "d"}} {
		if u.n != 0 {
			parts = append(parts, strconv.Itoa(u.n)+u.unit)
		}
	}
	if len(parts) == 1 && a.Bound != "" {
		parts = append(parts, "0y")
	}
	return strings.Join(parts, " ")
}

// ParseAge reads an Age payload.
func ParseAge(s string) (Age, error) {
	p, err := parsePayload("age", "AGE", s)
	if err != nil {
		return Age{}, err
	}
	defer p.release()
	n := p.value("age")
	if n == nil {
		return Age{}, nil
	}
	a := Age{Bound: p.childText(n, "age_bound")}
	d := child(n, p.lang, "age_duration")
	for _, u := range []struct {
		typ string
		dst *int
	}{{"years", &a.Years}, {"months", &a.Months}, {"weeks", &a.Weeks}, {"days", &a.Days}} {
		if t := p.childText(d, u.typ); t != "" {
			*u.dst, _ = strconv.Atoi(t[:len(t)-1])
		}
	}
	return a, nil
}
