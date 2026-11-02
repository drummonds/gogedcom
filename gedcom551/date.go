package gedcom551

import (
	"fmt"
	"strconv"
	"strings"

	"git.bytestone.uk/hum3/gotreesitter"

	"git.bytestone.uk/hum3/gogedcom/internal/tstree"
)

// Kind is which of the DATE_VALUE alternatives a value is.
type Kind int

const (
	NoDate          Kind = iota // the empty payload
	DateKind                    // DATE: one calendar date
	PeriodKind                  // DATE_PERIOD: FROM and/or TO
	RangeKind                   // DATE_RANGE: BET ... AND, AFT, BEF
	ApproximateKind             // DATE_APPROXIMATED: ABT, CAL, EST
	InterpretedKind             // INT <DATE> (<DATE_PHRASE>)
	PhraseKind                  // (<DATE_PHRASE>)
)

// Date is the specification's DATE: [calendar escape] [[day] month] year
// [B.C.]. Calendar is the escape's name (JULIAN, HEBREW, FRENCH R,
// GREGORIAN, ROMAN, UNKNOWN) or "" when absent; Dual is the dual-year
// modifier after the slash in 1699/00, "" when absent; Epoch is "B.C." or
// "". Day is 0 and Month "" when absent.
type Date struct {
	Calendar string
	Day      int
	Month    string
	Year     int
	Dual     string
	Epoch    string
}

// String is the date as GEDCOM writes it.
func (d *Date) String() string {
	if d == nil {
		return ""
	}
	var parts []string
	if d.Calendar != "" {
		parts = append(parts, "@#D"+d.Calendar+"@")
	}
	if d.Day != 0 {
		parts = append(parts, strconv.Itoa(d.Day))
	}
	if d.Month != "" {
		parts = append(parts, d.Month)
	}
	year := strconv.Itoa(d.Year)
	if d.Dual != "" {
		year += "/" + d.Dual
	}
	parts = append(parts, year)
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

// DateValue is a DATE_VALUE payload. Qualifier is the range or
// approximation word as written (ABT, CAL, EST, BET, AFT, BEF); From and
// To are the bounds: a plain, approximate or interpreted date and the
// lower bound of a range or period in From, the upper bound in To. Phrase
// is the text inside the parentheses of an interpreted or phrase value.
type DateValue struct {
	Kind      Kind
	Qualifier string
	From      *Date
	To        *Date
	Phrase    string
}

// Equal is whether two values say the same thing.
func (v DateValue) Equal(o DateValue) bool {
	return v.Kind == o.Kind && v.Qualifier == o.Qualifier && v.From.equal(o.From) && v.To.equal(o.To) && v.Phrase == o.Phrase
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
	case InterpretedKind:
		return "INT " + v.From.String() + " (" + v.Phrase + ")"
	case PhraseKind:
		return "(" + v.Phrase + ")"
	}
	return ""
}

// ParseDate reads a DATE_VALUE payload: the grammar's typed value
// "DATE <s>". A day over 31 or with a dual-year modifier is an error.
func ParseDate(s string) (DateValue, error) {
	if strings.ContainsAny(s, "\r\n") {
		return DateValue{}, fmt.Errorf("gedcom551: date %q: line break", s)
	}
	line := "DATE"
	if s != "" {
		line += " " + s
	}
	// gotreesitter (v0.7.0) mis-parses input that ends right after a GLR
	// fork; a trailing terminator avoids it.
	src := []byte(line + "\n")
	lang := Gedcom551Language()
	tree, err := gotreesitter.NewParser(lang).Parse(src)
	if err != nil {
		return DateValue{}, fmt.Errorf("gedcom551: date %q: %w", s, err)
	}
	defer tree.Release()
	root := tree.RootNode()
	if err := tstree.ParseError(root, lang, 1); err != nil {
		return DateValue{}, fmt.Errorf("gedcom551: date %q: %w", s, err)
	}
	typed := tstree.Child(root, lang, "date_line")
	if typed == nil {
		return DateValue{}, fmt.Errorf("gedcom551: date %q: not a DATE value", s)
	}
	value := tstree.Child(typed, lang, "date_value")
	if value == nil {
		return DateValue{}, nil
	}
	v := dateValue(value, src, lang)
	for _, d := range []*Date{v.From, v.To} {
		if d != nil && d.Day < 0 || d != nil && d.Day > 31 {
			return DateValue{}, fmt.Errorf("gedcom551: date %q: day %d", s, d.Day)
		}
	}
	return v, nil
}

// dateValue reads a date_value node.
func dateValue(n *gotreesitter.Node, src []byte, lang *gotreesitter.Language) DateValue {
	c := tstree.FirstNamed(n)
	dates := tstree.Children(c, lang, "date")
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
	case "date_interp":
		return DateValue{Kind: InterpretedKind, From: toDate(dates[0], src, lang), Phrase: phrase(tstree.Child(c, lang, "date_phrase"), src)}
	case "date_phrase":
		return DateValue{Kind: PhraseKind, Phrase: phrase(c, src)}
	}
	return DateValue{}
}

// phrase is the text inside a date_phrase token's parentheses. The token
// carries its leading delimiter when it follows a date.
func phrase(n *gotreesitter.Node, src []byte) string {
	t := strings.TrimPrefix(n.Text(src), " ")
	return strings.TrimSuffix(strings.TrimPrefix(t, "("), ")")
}

// toDate reads a date node. A day with a dual-year modifier (5/00) is
// reported as day -1, which ParseDate rejects.
func toDate(n *gotreesitter.Node, src []byte, lang *gotreesitter.Language) *Date {
	d := &Date{}
	if c := tstree.Child(n, lang, "calendar"); c != nil {
		d.Calendar = strings.TrimSuffix(strings.TrimPrefix(c.Text(src), "@#D"), "@")
	}
	if c := tstree.Child(n, lang, "day"); c != nil {
		var err error
		if d.Day, err = strconv.Atoi(c.Text(src)); err != nil {
			d.Day = -1
		}
	}
	if c := tstree.Child(n, lang, "month"); c != nil {
		d.Month = c.Text(src)
	}
	if c := tstree.Child(n, lang, "year"); c != nil {
		year, dual, _ := strings.Cut(c.Text(src), "/")
		d.Year, _ = strconv.Atoi(year)
		d.Dual = dual
	}
	if c := tstree.Child(n, lang, "epoch"); c != nil {
		// The epoch token carries its leading delimiter so the grammar stays
		// LR(1); the delimiter is not part of the epoch.
		d.Epoch = strings.TrimPrefix(c.Text(src), " ")
	}
	return d
}
