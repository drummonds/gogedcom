// Package convert turns a GEDCOM 5.5.1 tree into a GEDCOM 7 tree. The
// rule: 7's grammar decides what must change; the registry decides what
// may change where the change is lossless or the specification's
// migration guide prescribes it; everything else passes through as the
// file has it, and the original wording of what is re-spelt is kept in
// PHRASE. The decision table is docs/convert-551-to-7.md.
package convert

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"git.bytestone.uk/hum3/gogedcom/gedcom"
	"git.bytestone.uk/hum3/gogedcom/gedcom551"
	"git.bytestone.uk/hum3/gogedcom/gedcom7"
)

const v7 = "https://gedcom.io/terms/v7/"

// 7's Tag and Xref rules.
var (
	tagRe  = regexp.MustCompile(`^([A-Z]|_)[A-Z0-9_]*$`)
	xrefRe = regexp.MustCompile(`^@[A-Z0-9_]+@$`)
	spaces = regexp.MustCompile(` +`)
)

// To7 is doc, read as 5.5.1, as a GEDCOM 7 tree: a new tree, the input
// untouched. The header gets GEDC.VERS 7.0 and, after GEDC, a SCHMA
// declaring each relocated standard tag used; the records keep their
// order, with the OBJE and SOUR records made from inline structures
// appended. The error names the line of a tag or cross-reference
// identifier 7 cannot carry.
func To7(doc *gedcom.Document) (*gedcom.Document, error) {
	c := &converter{
		reg:    gedcom7.Standard(),
		out:    &gedcom.Document{},
		xrefs:  map[string]bool{},
		schema: map[string]string{},
	}
	for _, r := range doc.Records {
		c.xrefs[r.XRef] = true
	}
	if doc.Header != nil {
		if s := doc.Header.Find("SOUR"); s != nil {
			c.headSour = s.Payload
		}
		c.out.Header = c.header(doc.Header)
	}
	for _, r := range doc.Records {
		if r.XRef != "" && !xrefRe.MatchString(r.XRef) {
			return nil, fmt.Errorf("convert: line %d: %q is not a GEDCOM 7 cross-reference identifier", r.Line, r.XRef)
		}
		for _, s := range c.structure(scope{typed: true}, r) {
			c.out.Records = append(c.out.Records, s)
		}
		if c.err != nil {
			return nil, c.err
		}
	}
	c.out.Records = append(c.out.Records, c.made...)
	if c.err != nil {
		return nil, c.err
	}
	if c.out.Header != nil && len(c.schema) > 0 {
		schma := &gedcom.Structure{Line: c.out.Header.Line, Tag: "SCHMA"}
		for _, tag := range slices.Sorted(maps.Keys(c.schema)) {
			schma.Substructures = append(schma.Substructures, &gedcom.Structure{Line: c.out.Header.Line, Tag: "TAG", Payload: tag + " " + c.schema[tag]})
		}
		h := c.out.Header
		i := slices.IndexFunc(h.Substructures, func(s *gedcom.Structure) bool { return s.Tag == "GEDC" }) + 1
		h.Substructures = slices.Insert(h.Substructures, i, schma)
	}
	return c.out, nil
}

type converter struct {
	reg      *gedcom7.Registry
	out      *gedcom.Document
	made     []*gedcom.Structure // the OBJE and SOUR records made from inline structures
	xrefs    map[string]bool     // every identifier in use
	schema   map[string]string   // relocated tag -> the standard URI
	headSour string              // HEAD.SOUR, for RIN's TYPE
	err      error
}

func (c *converter) fail(line int, format string, args ...any) {
	if c.err == nil {
		c.err = fmt.Errorf("convert: line %d: %s", line, fmt.Sprintf(format, args...))
	}
}

// scope is the superstructure a structure is judged under: its 7 type
// (nil at record level) and standard tag, and whether it is typed at all.
// Under an untyped structure (a relocated tag with no term) nothing is
// relocated or checked.
type scope struct {
	typ   *gedcom7.StructureType
	tag   string
	typed bool
}

// header converts HEAD: its substructures by their rules, GEDC.VERS 7.0
// added when absent.
func (c *converter) header(h *gedcom.Structure) *gedcom.Structure {
	out := c.structure(scope{typed: true}, h)[0]
	gedc := out.Find("GEDC")
	if gedc == nil {
		gedc = &gedcom.Structure{Line: h.Line, Tag: "GEDC"}
		out.Substructures = append(out.Substructures, gedc)
	}
	if gedc.Find("VERS") == nil {
		gedc.Substructures = append(gedc.Substructures, &gedcom.Structure{Line: gedc.Line, Tag: "VERS", Payload: "7.0"})
	}
	return out
}

// structure converts s under sc: the structures that take its place, none
// when it is dropped, with their substructures converted.
func (c *converter) structure(sc scope, s *gedcom.Structure) []*gedcom.Structure {
	if c.err != nil {
		return nil
	}
	if !tagRe.MatchString(s.Tag) {
		c.fail(s.Line, "%q is not a GEDCOM 7 tag", s.Tag)
		return nil
	}
	if strings.HasPrefix(s.Tag, "_") || !sc.typed {
		return []*gedcom.Structure{c.copy(s)}
	}
	out := &gedcom.Structure{Line: s.Line, XRef: s.XRef, Tag: s.Tag, Payload: s.Payload}
	// children is what out gets, converted and placed like any 5.5.1
	// structure: s's substructures, and the PHRASE, TYPE or LANG a rule adds.
	children := s.Substructures
	add := func(tag, payload string) {
		children = append(slices.Clone(children), &gedcom.Structure{Line: s.Line, Tag: tag, Payload: payload})
	}
	rule := ruleFor(sc.tag, s.Tag)
	switch rule {
	case dropped:
		return nil
	case version:
		out.Payload = "7.0"
	case inil:
		out.Tag = "INIL"
	case snote:
		if sc.tag == "" || isPointer(s.Payload) {
			out.Tag = "SNOTE"
		}
	case exid:
		out.Tag = "EXID"
		typ := v7 + s.Tag
		switch s.Tag {
		case "RFN":
			if prefix, value, ok := strings.Cut(s.Payload, ":"); ok {
				out.Payload, typ = value, typ+"#"+prefix
			}
		case "RIN":
			if c.headSour != "" {
				typ += "#" + c.headSour
			}
		}
		add("TYPE", typ)
	case tran:
		// LANG takes the TYPE's place; a TYPE outside the guide's table
		// stays, after LANG und, for the placement rule to relocate.
		out.Tag = "TRAN"
		children = slices.Clone(children)
		lang := &gedcom.Structure{Line: s.Line, Tag: "LANG", Payload: "und"}
		i := slices.IndexFunc(children, func(t *gedcom.Structure) bool { return t.Tag == "TYPE" })
		if i < 0 {
			i = 0
		} else if l, ok := languages[strings.ToLower(children[i].Payload)]; ok {
			lang.Payload = l
			children = slices.Delete(children, i, i+1)
		}
		children = slices.Insert(children, i, lang)
	case role:
		out.Tag = "ROLE"
	case medi:
		out.Tag = "MEDI"
	case mediaType:
		if m, ok := mediaTypes[strings.ToLower(s.Payload)]; ok {
			out.Payload = m
		}
	case objeRecord:
		if !isPointer(s.Payload) {
			out.Payload = c.objeRecord(s)
			children = extensions(s.Substructures)
		}
	case citation:
		switch {
		case isPointer(s.Payload):
		case s.Payload == "":
			out.Payload = "@VOID@"
		default:
			rec := &gedcom.Structure{Line: s.Line, XRef: c.newXRef("S"), Tag: "SOUR",
				Substructures: []*gedcom.Structure{{Line: s.Line, Tag: "TITL", Payload: s.Payload}}}
			c.made = append(c.made, rec)
			out.Payload = rec.XRef
		}
	case date:
		if spelt, ok := date7(s.Payload); spelt != s.Payload || !ok {
			out.Payload = spelt
			add("PHRASE", s.Payload)
		}
	case age:
		if _, err := gedcom7.ParseAge(s.Payload); err != nil {
			out.Payload = ages[s.Payload]
			add("PHRASE", s.Payload)
		}
	case name:
		if _, err := gedcom7.ParsePersonalName(s.Payload); err != nil {
			out.Payload = name7(s.Payload)
			add("PHRASE", s.Payload)
		}
	case address:
		if s.Payload == "" {
			out.Payload = composeAddress(s)
		}
	}

	// Placement: a standard tag with no entry under the superstructure's
	// type is relocated, as an extension tag declared in SCHMA with the
	// standard URI, and its substructures judged against the standard type.
	typ := c.reg.Substructure(sc.typ, out.Tag)
	inner := scope{typ: typ, tag: out.Tag, typed: typ != nil}
	if typ == nil {
		inner.typ, inner.typed = c.relocate(out)
	}
	if typ != nil {
		switch rule {
		case enumeration, medi:
			if v, ok := c.enumerated(typ, out.Payload); ok {
				out.Payload = v
			}
		case role:
			if v, ok := c.enumerated(typ, out.Payload); ok {
				out.Payload = v
			} else {
				out.Payload = "OTHER"
				add("PHRASE", s.Payload)
			}
		}
		if strings.HasPrefix(typ.Payload, "@<") && out.Payload != "" && !xrefRe.MatchString(out.Payload) {
			c.fail(s.Line, "%q is not a GEDCOM 7 pointer", out.Payload)
			return nil
		}
	}
	for _, ch := range children {
		out.Substructures = append(out.Substructures, c.structure(inner, ch)...)
	}
	return []*gedcom.Structure{out}
}

// relocate makes out an extension tag and declares it; the type its
// substructures are judged against is the standard tag's term, when the
// registry has one.
func (c *converter) relocate(out *gedcom.Structure) (*gedcom7.StructureType, bool) {
	tag := out.Tag
	out.Tag = "_" + tag
	typ := c.reg.Structure(v7 + tag)
	if typ == nil {
		if types := c.reg.Types(tag); len(types) > 0 {
			typ = types[0]
		}
	}
	if typ == nil {
		return nil, false
	}
	c.schema[out.Tag] = typ.URI
	return typ, true
}

// enumerated is payload upper-cased, with 7's renamings, and whether that
// is a value of typ's set (every value, for a list); when it is not, the
// payload stays as it is, for Validate to report.
func (c *converter) enumerated(typ *gedcom7.StructureType, payload string) (string, bool) {
	if typ.EnumerationSet == "" {
		return payload, false
	}
	values := c.reg.Enumeration(typ.EnumerationSet)
	var items []string
	for _, item := range strings.Split(payload, ",") {
		item = strings.ToUpper(strings.TrimSpace(item))
		if r, ok := renamed[item]; ok {
			item = r
		}
		if !slices.Contains(values, item) {
			return payload, false
		}
		items = append(items, item)
	}
	return strings.Join(items, ", "), true
}

// objeRecord makes the OBJE record for an inline OBJE and returns its
// identifier: every standard substructure goes to the record, TITL under
// the record's first FILE; the extension substructures stay on the
// pointer.
func (c *converter) objeRecord(s *gedcom.Structure) string {
	rec := &gedcom.Structure{Line: s.Line, XRef: c.newXRef("O"), Tag: "OBJE"}
	c.made = append(c.made, rec)
	var titl *gedcom.Structure
	sc := scope{typ: c.reg.Structure(v7 + "record-OBJE"), tag: "OBJE", typed: true}
	for _, ch := range s.Substructures {
		switch {
		case strings.HasPrefix(ch.Tag, "_"):
		case ch.Tag == "TITL" && titl == nil:
			titl = ch
		default:
			rec.Substructures = append(rec.Substructures, c.structure(sc, ch)...)
		}
	}
	if titl != nil {
		if file := rec.Find("FILE"); file != nil {
			file.Substructures = append(file.Substructures, c.structure(scope{typ: c.reg.Structure(v7 + "FILE"), tag: "FILE", typed: true}, titl)...)
		} else {
			rec.Substructures = append(rec.Substructures, c.structure(sc, titl)...)
		}
	}
	return rec.XRef
}

// newXRef is the next unused identifier with the prefix: @O1@, @O2@, ...
func (c *converter) newXRef(prefix string) string {
	for n := 1; ; n++ {
		x := "@" + prefix + strconv.Itoa(n) + "@"
		if !c.xrefs[x] {
			c.xrefs[x] = true
			return x
		}
	}
}

// extensions is the extension-tagged structures among ss.
func extensions(ss []*gedcom.Structure) []*gedcom.Structure {
	var out []*gedcom.Structure
	for _, s := range ss {
		if strings.HasPrefix(s.Tag, "_") {
			out = append(out, s)
		}
	}
	return out
}

// isPointer is whether a 5.5.1 payload is a pointer: @X@ with X not
// beginning with an at-sign.
func isPointer(payload string) bool {
	return len(payload) > 2 && payload[0] == '@' && payload[1] != '@' && payload[len(payload)-1] == '@'
}

// copy is a deep copy of s, every tag checked against 7's rule.
func (c *converter) copy(s *gedcom.Structure) *gedcom.Structure {
	if !tagRe.MatchString(s.Tag) {
		c.fail(s.Line, "%q is not a GEDCOM 7 tag", s.Tag)
	}
	out := *s
	out.Substructures = nil
	for _, ch := range s.Substructures {
		out.Substructures = append(out.Substructures, c.copy(ch))
	}
	return &out
}

// date7 is a 5.5.1 DATE payload as 7 spells it, and whether 7 can spell
// it at all: the payload itself when 7's grammar accepts it; else the
// calendar escape as a word, B.C. as BCE, the first year of a dual year,
// the date of an interpreted value; nothing for a phrase or a payload
// 5.5.1's data type rejects.
func date7(payload string) (string, bool) {
	if _, err := gedcom7.ParseDate(payload); err == nil {
		return payload, true
	}
	v, err := gedcom551.ParseDate(payload)
	if err != nil {
		return "", false
	}
	var out string
	switch v.Kind {
	case gedcom551.NoDate:
		return "", true
	case gedcom551.PhraseKind:
		return "", false
	case gedcom551.DateKind, gedcom551.InterpretedKind:
		out = date7Of(v.From)
	case gedcom551.ApproximateKind:
		out = v.Qualifier + " " + date7Of(v.From)
	case gedcom551.RangeKind:
		switch v.Qualifier {
		case "BET":
			out = "BET " + date7Of(v.From) + " AND " + date7Of(v.To)
		case "AFT":
			out = "AFT " + date7Of(v.From)
		default:
			out = "BEF " + date7Of(v.To)
		}
	case gedcom551.PeriodKind:
		switch {
		case v.From == nil:
			out = "TO " + date7Of(v.To)
		case v.To == nil:
			out = "FROM " + date7Of(v.From)
		default:
			out = "FROM " + date7Of(v.From) + " TO " + date7Of(v.To)
		}
	}
	if _, err := gedcom7.ParseDate(out); err != nil {
		return "", false
	}
	return out, true
}

// date7Of is one 5.5.1 date as 7 spells it.
func date7Of(d *gedcom551.Date) string {
	var parts []string
	if d.Calendar != "" {
		cal, ok := calendars[d.Calendar]
		if !ok {
			cal = "_" + strings.ReplaceAll(d.Calendar, " ", "_")
		}
		parts = append(parts, cal)
	}
	if d.Day != 0 {
		parts = append(parts, strconv.Itoa(d.Day))
	}
	if d.Month != "" {
		parts = append(parts, d.Month)
	}
	parts = append(parts, strconv.Itoa(d.Year))
	if d.Epoch != "" {
		parts = append(parts, "BCE")
	}
	return strings.Join(parts, " ")
}

// name7 is a NAME payload 7 rejects with each slash and tab a space and
// runs of spaces one; "" when 7 still rejects it.
func name7(payload string) string {
	n := strings.NewReplacer("/", " ", "\t", " ").Replace(payload)
	n = strings.TrimSpace(spaces.ReplaceAllString(n, " "))
	if _, err := gedcom7.ParsePersonalName(n); err != nil {
		return ""
	}
	return n
}

// composeAddress is an ADDR payload made of the address pieces present,
// one per line, in the order the structure has them.
func composeAddress(s *gedcom.Structure) string {
	var lines []string
	for _, c := range s.Substructures {
		switch c.Tag {
		case "ADR1", "ADR2", "ADR3", "CITY", "STAE", "POST", "CTRY":
			if c.Payload != "" {
				lines = append(lines, c.Payload)
			}
		}
	}
	return strings.Join(lines, "\n")
}
