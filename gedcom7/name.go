package gedcom7

import "strings"

// PersonalName is the spec's PersonalName rule: the surname between
// slashes, with the text before it (a prefix and the given names, as the
// culture orders them) in Given and the text after it in Suffix, each
// trimmed of the spaces that separate it from the slashes. A name without
// slashes is all Given.
type PersonalName struct {
	Given   string
	Surname string
	Suffix  string
}

// String is the name as GEDCOM writes it. A name with no surname and no
// suffix writes without slashes.
func (n PersonalName) String() string {
	if n.Surname == "" && n.Suffix == "" {
		return n.Given
	}
	var parts []string
	if n.Given != "" {
		parts = append(parts, n.Given)
	}
	parts = append(parts, "/"+n.Surname+"/")
	if n.Suffix != "" {
		parts = append(parts, n.Suffix)
	}
	return strings.Join(parts, " ")
}

// ParsePersonalName reads a PersonalName payload.
func ParsePersonalName(s string) (PersonalName, error) {
	p, err := parsePayload("name", "NAME", s)
	if err != nil {
		return PersonalName{}, err
	}
	defer p.release()
	var n PersonalName
	parts := []*string{&n.Given, &n.Surname, &n.Suffix}
	slashes := 0
	for _, c := range p.value("personal_name").Children() {
		switch {
		case c.IsNamed():
			*parts[slashes] = strings.Trim(p.text(c), " ")
		case p.text(c) == "/":
			slashes++
		}
	}
	return n, nil
}
