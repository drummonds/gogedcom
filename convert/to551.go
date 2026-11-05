package convert

import (
	"slices"
	"strconv"
	"strings"

	"git.bytestone.uk/hum3/gogedcom/gedcom"
	"git.bytestone.uk/hum3/gogedcom/gedcom551"
	"git.bytestone.uk/hum3/gogedcom/gedcom7"
)

// To551 is doc, read as 7, as a GEDCOM 5.5.1 tree: a new tree, the input
// untouched, by the decision table in docs/convert-7-to-551.md. 5.5.1's
// line syntax decides what must change (a tag 5.5.1 does not define
// begins with "_"); the 5.5.1 tag list decides what comes back, by To7's
// rules read backwards, the original wording in PHRASE first; everything
// else passes through as the tree has it. The header gets GEDC.VERS
// 5.5.1, GEDC.FORM and CHAR, and loses SCHMA; an OBJE record exactly one
// OBJE pointer names comes back inline and leaves the records. There is
// no error: a 7 tree has nothing a 5.5.1 line cannot carry.
func To551(doc *gedcom.Document) *gedcom.Document {
	c := &inverter{
		reg:       gedcom7.Standard(),
		out:       &gedcom.Document{},
		relocated: map[string]string{},
		records:   map[string]*gedcom.Structure{},
		inlined:   map[string]bool{},
	}
	for _, r := range doc.Records {
		if r.XRef != "" {
			c.records[r.XRef] = r
		}
	}
	// A record is inlined when the one payload naming it is an OBJE's.
	refs := map[string]int{}
	referrer := map[string]string{}
	var count func(s *gedcom.Structure)
	count = func(s *gedcom.Structure) {
		if isPointer(s.Payload) {
			refs[s.Payload]++
			referrer[s.Payload] = s.Tag
		}
		for _, ch := range s.Substructures {
			count(ch)
		}
	}
	if doc.Header != nil {
		count(doc.Header)
	}
	for _, r := range doc.Records {
		count(r)
	}
	for x, n := range refs {
		if r := c.records[x]; n == 1 && referrer[x] == "OBJE" && r != nil && r.Tag == "OBJE" {
			c.inlined[x] = true
		}
	}
	if doc.Header != nil {
		if schma := doc.Header.Find("SCHMA"); schma != nil {
			for _, t := range schma.All("TAG") {
				tag, uri, _ := strings.Cut(t.Payload, " ")
				if typ := c.reg.Structure(uri); typ != nil {
					c.relocated[tag] = typ.Tag
				}
			}
		}
		c.out.Header = c.header(doc.Header)
	}
	for _, r := range doc.Records {
		if c.inlined[r.XRef] {
			continue
		}
		c.out.Records = append(c.out.Records, c.structure(scope551{}, r)...)
	}
	return c.out
}

type inverter struct {
	reg       *gedcom7.Registry
	out       *gedcom.Document
	relocated map[string]string            // a SCHMA-declared tag -> the standard tag it relocates
	records   map[string]*gedcom.Structure // by identifier
	inlined   map[string]bool              // the OBJE records that come back inline
}

// scope551 is the superstructure a structure is judged under: its 7 tag
// ("" at record level) and whether it is inside an inlined OBJE.
type scope551 struct {
	tag    string
	inline bool
}

// header converts HEAD: its substructures by their rules, then 5.5.1's
// constants added where absent, in 5.5.1's places.
func (c *inverter) header(h *gedcom.Structure) *gedcom.Structure {
	out := c.structure(scope551{}, h)[0]
	add := func(tag, payload string) *gedcom.Structure {
		return &gedcom.Structure{Line: h.Line, Tag: tag, Payload: payload}
	}
	gedc := out.Find("GEDC")
	if gedc == nil {
		gedc = add("GEDC", "")
		out.Substructures = append(out.Substructures, gedc)
	}
	vers := gedc.Find("VERS")
	if vers == nil {
		vers = add("VERS", "5.5.1")
		gedc.Substructures = append(gedc.Substructures, vers)
	}
	if gedc.Find("FORM") == nil {
		i := slices.Index(gedc.Substructures, vers) + 1
		gedc.Substructures = slices.Insert(gedc.Substructures, i, add("FORM", "LINEAGE-LINKED"))
	}
	if out.Find("CHAR") == nil {
		i := slices.Index(out.Substructures, gedc) + 1
		out.Substructures = slices.Insert(out.Substructures, i, add("CHAR", "UTF-8"))
	}
	return out
}

// tagOf is s's tag with a SCHMA-declared relocation undone.
func (c *inverter) tagOf(s *gedcom.Structure) string {
	if std, ok := c.relocated[s.Tag]; ok {
		return std
	}
	return s.Tag
}

// find is the first substructure of s whose tag, relocation undone, is tag.
func (c *inverter) find(s *gedcom.Structure, tag string) *gedcom.Structure {
	for _, ch := range s.Substructures {
		if c.tagOf(ch) == tag {
			return ch
		}
	}
	return nil
}

// structure converts s under sc: the structures that take its place, none
// when it is consumed, with their substructures converted.
func (c *inverter) structure(sc scope551, s *gedcom.Structure) []*gedcom.Structure {
	tag := s.Tag
	if strings.HasPrefix(tag, "_") {
		std, ok := c.relocated[tag]
		if !ok {
			return []*gedcom.Structure{clone(s)}
		}
		tag = std
	}
	rule, key := ruleFor551(sc.tag, tag)
	out := &gedcom.Structure{Line: s.Line, XRef: s.XRef, Tag: tag, Payload: s.Payload}
	children := s.Substructures
	consume := func(ch *gedcom.Structure) {
		children = slices.DeleteFunc(slices.Clone(children), func(x *gedcom.Structure) bool { return x == ch })
	}
	switch rule {
	case schema:
		return nil
	case version:
		out.Payload = "5.5.1"
	case note:
		out.Tag = "NOTE"
	case exid:
		if tag != "EXID" {
			break
		}
		typ := s.Find("TYPE")
		if typ == nil || !strings.HasPrefix(typ.Payload, v7) {
			return []*gedcom.Structure{extensionOf(s)}
		}
		name, fragment, _ := strings.Cut(typ.Payload[len(v7):], "#")
		switch name {
		case "RFN", "AFN", "RIN":
		default:
			return []*gedcom.Structure{extensionOf(s)}
		}
		out.Tag = name
		if name == "RFN" && fragment != "" {
			out.Payload = fragment + ":" + s.Payload
		}
		consume(typ)
	case tran:
		if tag != "TRAN" {
			break
		}
		lang := s.Find("LANG")
		if lang == nil || (sc.tag != "NAME" && sc.tag != "PLAC") {
			return []*gedcom.Structure{extensionOf(s)}
		}
		t, ok := transliterations[lang.Payload]
		if !ok {
			return []*gedcom.Structure{extensionOf(s)}
		}
		out.Tag = t.tag
		children = slices.Clone(children)
		children[slices.Index(children, lang)] = &gedcom.Structure{Line: lang.Line, Tag: "TYPE", Payload: t.word}
	case role:
		if sc.tag == "ASSO" {
			out.Tag = "RELA"
		}
		if ph := s.Find("PHRASE"); ph != nil && s.Payload == "OTHER" {
			out.Payload = ph.Payload
			consume(ph)
		}
	case wac:
		out.Tag = "WAC"
	case enumeration:
		out.Payload = c.enumerated551(key, s.Payload)
	case mediaType:
		if w, ok := mediaWords[s.Payload]; ok {
			out.Payload = w
		}
	case medi:
		out.Tag = "TYPE"
		if sc.inline {
			out.Tag = "MEDI"
		}
		out.Payload = c.enumerated551(key, s.Payload)
	case inlineObje:
		if c.inlined[s.Payload] {
			return []*gedcom.Structure{c.inline(s)}
		}
	case citation:
		if s.Payload == "@VOID@" {
			out.Payload = ""
		}
	case date:
		if tag != "DATE" {
			break
		}
		ph := s.Find("PHRASE")
		switch {
		case ph == nil:
			out.Payload = date551(s.Payload)
		default:
			if spelt, _ := date7(ph.Payload); spelt == s.Payload {
				out.Payload = ph.Payload
			} else if d := date551(s.Payload); d == "" {
				out.Payload = "(" + ph.Payload + ")"
			} else {
				out.Payload = "INT " + d + " (" + ph.Payload + ")"
			}
			consume(ph)
		}
	case age:
		if ph := s.Find("PHRASE"); ph != nil {
			if _, err := gedcom7.ParseAge(ph.Payload); err != nil && ages[ph.Payload] == s.Payload {
				out.Payload = ph.Payload
				consume(ph)
			}
		}
	case name:
		if tag != "NAME" {
			break
		}
		if ph := c.find(s, "PHRASE"); ph != nil {
			out.Payload = ph.Payload
			consume(ph)
		}
	case address:
		if s.Payload == composeAddress(s) {
			out.Payload = ""
		}
	case extension:
		return []*gedcom.Structure{extensionOf(s)}
	}
	inner := scope551{tag: tag, inline: sc.inline}
	for _, ch := range children {
		out.Substructures = append(out.Substructures, c.structure(inner, ch)...)
	}
	return []*gedcom.Structure{out}
}

// inline is the OBJE pointer's record as an inline OBJE: the record's
// FILEs, each followed by its TITL, then the pointer's own substructures,
// then the rest of the record.
func (c *inverter) inline(ptr *gedcom.Structure) *gedcom.Structure {
	rec := c.records[ptr.Payload]
	out := &gedcom.Structure{Line: ptr.Line, Tag: "OBJE"}
	sc := scope551{tag: "OBJE", inline: true}
	var rest []*gedcom.Structure
	for _, ch := range rec.Substructures {
		if ch.Tag != "FILE" {
			rest = append(rest, ch)
			continue
		}
		file := *ch
		titl := ch.Find("TITL")
		if titl != nil {
			file.Substructures = slices.DeleteFunc(slices.Clone(ch.Substructures), func(x *gedcom.Structure) bool { return x == titl })
		}
		out.Substructures = append(out.Substructures, c.structure(sc, &file)...)
		if titl != nil {
			out.Substructures = append(out.Substructures, c.structure(sc, titl)...)
		}
	}
	for _, ch := range ptr.Substructures {
		out.Substructures = append(out.Substructures, c.structure(sc, ch)...)
	}
	for _, ch := range rest {
		out.Substructures = append(out.Substructures, c.structure(sc, ch)...)
	}
	return out
}

// enumerated551 is payload in 5.5.1's spelling when every value (of a
// list) is in the row's 7 set: lower-cased where 5.5.1 spells the set so,
// DNS_CAN and PRE_1970 renamed; else payload as it is.
func (c *inverter) enumerated551(key, payload string) string {
	e, ok := enumerationSets[key]
	if !ok {
		return payload
	}
	values := c.reg.Enumeration(v7 + e.set)
	var items []string
	for _, item := range strings.Split(payload, ",") {
		item = strings.TrimSpace(item)
		if !slices.Contains(values, item) {
			return payload
		}
		if r, ok := renamedBack[item]; ok {
			item = r
		}
		if e.lower {
			item = strings.ToLower(item)
		}
		items = append(items, item)
	}
	return strings.Join(items, ", ")
}

// extensionOf is s as a 5.5.1 extension: its tag under "_", everything
// under it as it is.
func extensionOf(s *gedcom.Structure) *gedcom.Structure {
	out := clone(s)
	out.Tag = "_" + s.Tag
	return out
}

// clone is a deep copy of s.
func clone(s *gedcom.Structure) *gedcom.Structure {
	out := *s
	out.Substructures = nil
	for _, ch := range s.Substructures {
		out.Substructures = append(out.Substructures, clone(ch))
	}
	return &out
}

// date551 is a 7 DATE payload as 5.5.1 spells it: the payload itself when
// 5.5.1's data type accepts it; else, when 7's does, the calendar as an
// escape and BCE as B.C.; else the payload as it is.
func date551(payload string) string {
	if _, err := gedcom551.ParseDate(payload); err == nil {
		return payload
	}
	v, err := gedcom7.ParseDate(payload)
	if err != nil {
		return payload
	}
	switch v.Kind {
	case gedcom7.DateKind:
		return date551Of(v.From)
	case gedcom7.ApproximateKind:
		return v.Qualifier + " " + date551Of(v.From)
	case gedcom7.RangeKind:
		switch v.Qualifier {
		case "BET":
			return "BET " + date551Of(v.From) + " AND " + date551Of(v.To)
		case "AFT":
			return "AFT " + date551Of(v.From)
		default:
			return "BEF " + date551Of(v.To)
		}
	case gedcom7.PeriodKind:
		switch {
		case v.From == nil:
			return "TO " + date551Of(v.To)
		case v.To == nil:
			return "FROM " + date551Of(v.From)
		default:
			return "FROM " + date551Of(v.From) + " TO " + date551Of(v.To)
		}
	}
	return payload
}

// date551Of is one 7 date as 5.5.1 spells it: the calendar as an escape
// (FRENCH_R as FRENCH R, an extension calendar without its underscore),
// BCE as B.C.
func date551Of(d *gedcom7.Date) string {
	var parts []string
	if d.Calendar != "" {
		cal := strings.ReplaceAll(strings.TrimPrefix(d.Calendar, "_"), "_", " ")
		parts = append(parts, "@#D"+cal+"@")
	}
	if d.Day != 0 {
		parts = append(parts, strconv.Itoa(d.Day))
	}
	if d.Month != "" {
		parts = append(parts, d.Month)
	}
	parts = append(parts, strconv.Itoa(d.Year))
	if d.Epoch != "" {
		parts = append(parts, "B.C.")
	}
	return strings.Join(parts, " ")
}
