package gedcom7

import (
	"fmt"
	"strconv"
	"strings"

	"git.bytestone.uk/hum3/gotreesitter"
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

// ParseDate reads a DateValue payload: the grammar's date_line, parsed as
// the line "0 DATE <s>". Where the grammar allows two readings (JULIAN
// 1700 is a calendar or a month) the first alternative wins: the calendar.
func ParseDate(s string) (DateValue, error) {
	if strings.ContainsAny(s, "\r\n") {
		return DateValue{}, fmt.Errorf("gedcom7: date %q: line break", s)
	}
	line := "0 DATE"
	if s != "" {
		line += " " + s
	}
	// gotreesitter (v0.7.0) mis-parses input that ends right after a GLR
	// fork; a trailing EOL avoids it.
	src := []byte(line + "\n")
	tree, lang, err := parse(src)
	if err != nil {
		return DateValue{}, fmt.Errorf("gedcom7: date %q: %w", s, err)
	}
	defer tree.Release()
	root := tree.RootNode()
	if e := firstError(root); e != nil {
		return DateValue{}, fmt.Errorf("gedcom7: date %q: %w", s, errorAt(e, lang, 1))
	}
	dateLine := child(child(root, lang, "line"), lang, "date_line")
	if dateLine == nil {
		return DateValue{}, fmt.Errorf("gedcom7: date %q: not a DATE line", s)
	}
	value := child(dateLine, lang, "date_value")
	if value == nil {
		return DateValue{}, nil
	}
	v := dateValue(value, src, lang)
	for _, d := range []*Date{v.From, v.To} {
		if d != nil && (d.Day < 0 || d.Day > 31 || (d.Day == 0 && child(value, lang, "day") != nil)) {
			return DateValue{}, fmt.Errorf("gedcom7: date %q: day %d", s, d.Day)
		}
	}
	return v, nil
}

// dateValue reads a date_value node.
func dateValue(n *gotreesitter.Node, src []byte, lang *gotreesitter.Language) DateValue {
	c := firstNamed(n)
	dates := children(c, lang, "date")
	text := c.Text(src)
	switch c.Type(lang) {
	case "date":
		return DateValue{Kind: DateKind, From: toDate(c, src, lang)}
	case "date_approx":
		return DateValue{Kind: ApproximateKind, Qualifier: text[:3], From: toDate(dates[0], src, lang)}
	case "date_range":
		v := DateValue{Kind: RangeKind, Qualifier: text[:3]}
		switch v.Qualifier {
		case "BET":
			v.From, v.To = toDate(dates[0], src, lang), toDate(dates[1], src, lang)
		case "AFT":
			v.From = toDate(dates[0], src, lang)
		case "BEF":
			v.To = toDate(dates[0], src, lang)
		}
		return v
	case "date_period":
		v := DateValue{Kind: PeriodKind}
		if strings.HasPrefix(text, "TO") {
			v.To = toDate(dates[0], src, lang)
		} else {
			v.From = toDate(dates[0], src, lang)
			if len(dates) > 1 {
				v.To = toDate(dates[1], src, lang)
			}
		}
		return v
	}
	return DateValue{}
}

func toDate(n *gotreesitter.Node, src []byte, lang *gotreesitter.Language) *Date {
	d := &Date{}
	if c := child(n, lang, "calendar"); c != nil {
		d.Calendar = c.Text(src)
	}
	if c := child(n, lang, "day"); c != nil {
		d.Day, _ = strconv.Atoi(c.Text(src))
	}
	if c := child(n, lang, "month"); c != nil {
		d.Month = c.Text(src)
	}
	if c := child(n, lang, "year"); c != nil {
		d.Year, _ = strconv.Atoi(c.Text(src))
	}
	if c := child(n, lang, "epoch"); c != nil {
		d.Epoch = c.Text(src)
	}
	return d
}
