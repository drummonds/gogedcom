// Package gedcom holds the tree a GEDCOM file reads to, whichever version
// wrote it: lines nested by level. The gedcom7 and gedcom551 readers both
// return it; what a tag or payload means is the version's business.
package gedcom

// Structure is a line with its substructures: the lines one level deeper
// that follow it. CONT lines (and in 5.5.1 CONC lines) are folded into the
// payload.
type Structure struct {
	Line          int    // the file line it starts on, from 1
	XRef          string // "@I1@" on a record, else ""
	Tag           string
	Payload       string
	Substructures []*Structure
}

// Find is the first substructure with the tag, or nil.
func (s *Structure) Find(tag string) *Structure {
	for _, c := range s.Substructures {
		if c.Tag == tag {
			return c
		}
	}
	return nil
}

// All is every substructure with the tag, in order.
func (s *Structure) All(tag string) []*Structure {
	var out []*Structure
	for _, c := range s.Substructures {
		if c.Tag == tag {
			out = append(out, c)
		}
	}
	return out
}

// Document is a file: the header and the records between it and the
// trailer.
type Document struct {
	Header  *Structure
	Records []*Structure
}
