package gedcom7

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// Violation is a structure that breaks the registry: the line it is on and
// what it breaks.
type Violation struct {
	Line    int
	Message string
}

func (v Violation) String() string { return fmt.Sprintf("line %d: %s", v.Line, v.Message) }

// Validate is every violation of the registry in the document, in line
// order: a standard tag where the registry allows none, a singleton
// repeated, a required substructure missing (reported at the
// superstructure's line), an enumeration payload not in its set. An
// extension tag and everything under it passes unchecked, and so does
// everything under a structure the registry does not place: one violation,
// not a cascade.
func (r *Registry) Validate(doc *Document) []Violation {
	v := &validator{r: r}
	if doc.Header != nil {
		v.structure(nil, doc.Header)
	}
	for _, rec := range doc.Records {
		v.structure(nil, rec)
	}
	slices.SortStableFunc(v.out, func(a, b Violation) int { return a.Line - b.Line })
	return v.out
}

type validator struct {
	r   *Registry
	out []Violation
}

func (v *validator) add(line int, format string, args ...any) {
	v.out = append(v.out, Violation{Line: line, Message: fmt.Sprintf(format, args...)})
}

// structure checks s under a superstructure of type super (nil at document
// level) and returns s's type, or nil when it has none or is an extension.
func (v *validator) structure(super *StructureType, s *Structure) *StructureType {
	if strings.HasPrefix(s.Tag, "_") {
		return nil
	}
	typ := v.r.Substructure(super, s.Tag)
	switch {
	case typ != nil:
	case super == nil:
		v.add(s.Line, "%s is not a record type", s.Tag)
		return nil
	default:
		v.add(s.Line, "%s is not a substructure of %s", s.Tag, super.Name())
		return nil
	}
	v.payload(typ, s)
	counts := map[string]int{}
	for _, c := range s.Substructures {
		ct := v.structure(typ, c)
		if ct == nil {
			continue
		}
		counts[ct.URI]++
		if counts[ct.URI] == 2 && !typ.Substructures[ct.URI].Repeatable {
			v.add(c.Line, "%s has more than one %s", typ.Name(), c.Tag)
		}
	}
	for _, uri := range slices.Sorted(maps.Keys(typ.Substructures)) {
		if typ.Substructures[uri].Required && counts[uri] == 0 {
			v.add(s.Line, "%s lacks %s", typ.Name(), v.r.structures[uri].Tag)
		}
	}
	return typ
}

// payload checks an enumeration payload against its set.
func (v *validator) payload(typ *StructureType, s *Structure) {
	if typ.EnumerationSet == "" {
		return
	}
	if s.Payload == "" {
		v.add(s.Line, "%s has no payload", s.Tag)
		return
	}
	items := []string{s.Payload}
	if strings.HasSuffix(typ.Payload, "#Enum") {
		items = strings.Split(s.Payload, ",")
	}
	values := v.r.Enumeration(typ.EnumerationSet)
	for _, item := range items {
		if item = strings.TrimSpace(item); !slices.Contains(values, item) {
			v.add(s.Line, "%s is not in %s", item, name(typ.EnumerationSet))
		}
	}
}
