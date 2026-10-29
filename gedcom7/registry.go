package gedcom7

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
	"sync"

	"github.com/goccy/go-yaml"
)

// The registry is the specification's extracted-files/tags, one YAML file
// per URI-identified term, vendored at the commit in grammar.commit by
// task registry:fetch.
//
//go:embed registry
var registryFiles embed.FS

// Cardinality is how many of a substructure a superstructure takes: {0:1},
// {0:M}, {1:1} or {1:M}.
type Cardinality struct {
	Required   bool // at least one: {1:1}, {1:M}
	Repeatable bool // more than one allowed: {0:M}, {1:M}
}

// String is the cardinality as the specification writes it.
func (c Cardinality) String() string {
	s := "{0:"
	if c.Required {
		s = "{1:"
	}
	if c.Repeatable {
		return s + "M}"
	}
	return s + "1}"
}

func parseCardinality(s string) (Cardinality, error) {
	switch s {
	case "{0:1}":
		return Cardinality{}, nil
	case "{0:M}":
		return Cardinality{Repeatable: true}, nil
	case "{1:1}":
		return Cardinality{Required: true}, nil
	case "{1:M}":
		return Cardinality{Required: true, Repeatable: true}, nil
	}
	return Cardinality{}, fmt.Errorf("cardinality %q", s)
}

// StructureType is a term of type structure: what a tag means under a
// given superstructure. One tag is several types (DATE is DATE,
// DATE-exact or HEAD-DATE), so a structure is typed by its superstructure's
// type and its tag.
type StructureType struct {
	URI             string
	Tag             string                 // the standard tag
	Payload         string                 // the payload type as the registry writes it; "" for none
	Substructures   map[string]Cardinality // by URI
	Superstructures map[string]Cardinality // by URI; empty at document level
	EnumerationSet  string                 // the set's URI for a type-Enum or type-List#Enum payload
}

// Name is the URI's last segment: record-INDI, DATE-exact.
func (s *StructureType) Name() string { return name(s.URI) }

func name(uri string) string { return uri[strings.LastIndex(uri, "/")+1:] }

// Registry is a set of terms: the structure types and enumeration sets a
// document is validated against.
type Registry struct {
	structures   map[string]*StructureType
	enumerations map[string][]string                  // set URI -> the values' tags
	byTag        map[string]map[string]*StructureType // superstructure URI ("" at document level) -> tag -> type
}

var (
	standardOnce sync.Once
	standard     *Registry
)

// Standard is the GEDCOM 7 registry embedded in the package.
func Standard() *Registry {
	standardOnce.Do(func() {
		r, err := readRegistry(registryFiles, "registry")
		if err != nil {
			panic("gedcom7: registry: " + err.Error())
		}
		standard = r
	})
	return standard
}

// Structure is the structure type with the URI, or nil.
func (r *Registry) Structure(uri string) *StructureType { return r.structures[uri] }

// Substructure is the type a tag has under a superstructure of the type,
// or at document level when super is nil; nil when the registry allows no
// such substructure.
func (r *Registry) Substructure(super *StructureType, tag string) *StructureType {
	uri := ""
	if super != nil {
		uri = super.URI
	}
	return r.byTag[uri][tag]
}

// Enumeration is the tags of the set's values, in the registry's order, or
// nil for an unknown set.
func (r *Registry) Enumeration(set string) []string { return r.enumerations[set] }

// term is a registry file.
type term struct {
	Type            string            `yaml:"type"`
	URI             string            `yaml:"uri"`
	Tag             string            `yaml:"standard tag"`
	Payload         string            `yaml:"payload"`
	Substructures   map[string]string `yaml:"substructures"`
	Superstructures map[string]string `yaml:"superstructures"`
	EnumerationSet  string            `yaml:"enumeration set"`
	Values          []string          `yaml:"enumeration values"`
}

// readRegistry reads every file under dir as a term.
func readRegistry(files fs.FS, dir string) (*Registry, error) {
	entries, err := fs.ReadDir(files, dir)
	if err != nil {
		return nil, err
	}
	r := &Registry{
		structures:   map[string]*StructureType{},
		enumerations: map[string][]string{},
		byTag:        map[string]map[string]*StructureType{},
	}
	tags := map[string]string{} // every term's standard tag, by URI
	var sets []term
	for _, e := range entries {
		data, err := fs.ReadFile(files, dir+"/"+e.Name())
		if err != nil {
			return nil, err
		}
		var t term
		if err := yaml.Unmarshal(data, &t); err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		tags[t.URI] = t.Tag
		switch t.Type {
		case "structure":
			s, err := structureType(t)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", e.Name(), err)
			}
			r.structures[s.URI] = s
		case "enumeration set":
			sets = append(sets, t)
		}
	}
	for _, s := range r.structures {
		supers := []string{""}
		if len(s.Superstructures) > 0 {
			supers = supers[:0]
			for uri := range s.Superstructures {
				supers = append(supers, uri)
			}
		}
		for _, uri := range supers {
			if r.byTag[uri] == nil {
				r.byTag[uri] = map[string]*StructureType{}
			}
			r.byTag[uri][s.Tag] = s
		}
	}
	for _, t := range sets {
		values := make([]string, 0, len(t.Values))
		for _, uri := range t.Values {
			tag, ok := tags[uri]
			if !ok {
				return nil, fmt.Errorf("%s: value %s is not a term", name(t.URI), uri)
			}
			values = append(values, tag)
		}
		r.enumerations[t.URI] = values
	}
	return r, nil
}

func structureType(t term) (*StructureType, error) {
	s := &StructureType{
		URI:             t.URI,
		Tag:             t.Tag,
		Payload:         t.Payload,
		Substructures:   map[string]Cardinality{},
		Superstructures: map[string]Cardinality{},
		EnumerationSet:  t.EnumerationSet,
	}
	for dst, src := range map[*map[string]Cardinality]map[string]string{&s.Substructures: t.Substructures, &s.Superstructures: t.Superstructures} {
		for uri, c := range src {
			card, err := parseCardinality(c)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", uri, err)
			}
			(*dst)[uri] = card
		}
	}
	return s, nil
}
