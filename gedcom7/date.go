package gedcom7

import (
	"fmt"
	"strconv"
	"strings"

	"git.bytestone.uk/hum3/gogedcom/abnf"
)

// Kind is which of the DateValue rule's alternatives a value is.
type Kind int

const (
	NoDate          Kind = iota // the empty payload
	DateKind                    // date: one calendar date
	PeriodKind                  // DatePeriod: FROM and/or TO
	RangeKind                   // dateRange: BET ... AND, AFT, BEF
	ApproximateKind             // dateApprox: ABT, CAL, EST
)

// Date is the spec's date rule: [calendar] [[day] month] year [epoch].
// Day is 0 and Month, Calendar, Epoch "" when absent.
type Date struct {
	Calendar string
	Day      int
	Month    string
	Year     int
	Epoch    string
}

// String is the date as GEDCOM writes it.
func (d *Date) String() string {
	if d == nil {
		return ""
	}
	var parts []string
	if d.Calendar != "" {
		parts = append(parts, d.Calendar)
	}
	if d.Day != 0 {
		parts = append(parts, strconv.Itoa(d.Day))
	}
	if d.Month != "" {
		parts = append(parts, d.Month)
	}
	parts = append(parts, strconv.Itoa(d.Year))
	if d.Epoch != "" {
		parts = append(parts, d.Epoch)
	}
	return strings.Join(parts, " ")
}

func (d *Date) equal(o *Date) bool {
	if d == nil || o == nil {
		return d == o
	}
	return *d == *o
}

// DateValue is a DateValue payload. Qualifier is the range or
// approximation word as written (ABT, CAL, EST, BET, AFT, BEF); From and To
// are the bounds: a plain or approximate date and the lower bound of a
// range or period in From, the upper bound in To.
type DateValue struct {
	Kind      Kind
	Qualifier string
	From      *Date
	To        *Date
}

// Equal is whether two values say the same thing.
func (v DateValue) Equal(o DateValue) bool {
	return v.Kind == o.Kind && v.Qualifier == o.Qualifier && v.From.equal(o.From) && v.To.equal(o.To)
}

// String is the value as GEDCOM writes it.
func (v DateValue) String() string {
	switch v.Kind {
	case DateKind:
		return v.From.String()
	case ApproximateKind:
		return v.Qualifier + " " + v.From.String()
	case RangeKind:
		switch v.Qualifier {
		case "BET":
			return "BET " + v.From.String() + " AND " + v.To.String()
		case "AFT":
			return "AFT " + v.From.String()
		default:
			return "BEF " + v.To.String()
		}
	case PeriodKind:
		switch {
		case v.From == nil:
			return "TO " + v.To.String()
		case v.To == nil:
			return "FROM " + v.From.String()
		default:
			return "FROM " + v.From.String() + " TO " + v.To.String()
		}
	}
	return ""
}

// months of each calendar, as the specification's registry lists them.
// Roadmap #2 reads them from the registry's YAML instead.
var months = map[string][]string{
	"GREGORIAN": {"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"},
	"JULIAN":    {"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"},
	"FRENCH_R":  {"VEND", "BRUM", "FRIM", "NIVO", "PLUV", "VENT", "GERM", "FLOR", "PRAI", "MESS", "THER", "FRUC", "COMP"},
	"HEBREW":    {"TSH", "CSH", "KSL", "TVT", "SHV", "ADR", "ADS", "NSN", "IYR", "SVN", "TMZ", "AAV", "ELL"},
}

// wellFormedDate is what the grammar says in comments: a date's month is
// one of its calendar's (any, for an extension calendar) and its day is a
// day of a month. Without it ABT 1850 reads as the month ABT of 1850.
func wellFormedDate(n *abnf.Node) bool {
	if n.Rule != "date" {
		return true
	}
	if d := n.Find("day"); d != nil {
		if v, _ := strconv.Atoi(d.Text); v < 1 || v > 31 {
			return false
		}
	}
	m := n.Find("month")
	if m == nil {
		return true
	}
	cal := "GREGORIAN"
	if c := n.Find("calendar"); c != nil {
		cal = c.Text
	}
	known, ok := months[cal]
	if !ok { // an extension calendar: any tag may be a month
		return true
	}
	for _, k := range known {
		if k == m.Text {
			return true
		}
	}
	return false
}

// ParseDate reads a DateValue payload. Where the grammar allows two
// readings (JULIAN 1700 is a calendar or a month) the rule's written
// order decides: the calendar.
func ParseDate(s string) (DateValue, error) {
	n, err := Grammar().MatchWhere("DateValue", s, wellFormedDate)
	if err != nil {
		return DateValue{}, fmt.Errorf("gedcom7: date %q: %w", s, err)
	}
	if len(n.Children) == 0 {
		return DateValue{}, nil
	}
	c := n.Children[0]
	dates := c.All("date")
	switch c.Rule {
	case "date":
		return DateValue{Kind: DateKind, From: toDate(c)}, nil
	case "dateApprox":
		return DateValue{Kind: ApproximateKind, Qualifier: c.Text[:3], From: toDate(dates[0])}, nil
	case "dateRange":
		v := DateValue{Kind: RangeKind, Qualifier: c.Text[:3]}
		switch v.Qualifier {
		case "BET":
			v.From, v.To = toDate(dates[0]), toDate(dates[1])
		case "AFT":
			v.From = toDate(dates[0])
		case "BEF":
			v.To = toDate(dates[0])
		}
		return v, nil
	case "DatePeriod":
		v := DateValue{Kind: PeriodKind}
		if strings.HasPrefix(c.Text, "TO") {
			v.To = toDate(dates[0])
		} else {
			v.From = toDate(dates[0])
			if len(dates) > 1 {
				v.To = toDate(dates[1])
			}
		}
		return v, nil
	}
	return DateValue{}, fmt.Errorf("gedcom7: date %q: unexpected rule %s", s, c.Rule)
}

func toDate(n *abnf.Node) *Date {
	d := &Date{}
	if c := n.Find("calendar"); c != nil {
		d.Calendar = c.Text
	}
	if c := n.Find("day"); c != nil {
		d.Day, _ = strconv.Atoi(c.Text)
	}
	if c := n.Find("month"); c != nil {
		d.Month = c.Text
	}
	if c := n.Find("year"); c != nil {
		d.Year, _ = strconv.Atoi(c.Text)
	}
	if c := n.Find("epoch"); c != nil {
		d.Epoch = c.Text
	}
	return d
}
