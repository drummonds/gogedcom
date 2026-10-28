package gedcom7

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Structure is a line with its substructures: the lines one level deeper
// that follow it. CONT lines are folded into the payload.
type Structure struct {
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

// Read reads a file. Every line is checked against the specification's
// Line rule; a leading byte order mark and CR LF line ends are accepted.
func Read(r io.Reader) (*Document, error) {
	g := Grammar()
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<24)
	doc := &Document{}
	var stack []*Structure // stack[i] is the open structure at level i
	lineNo := 0
	ended := false
	for sc.Scan() {
		lineNo++
		text := strings.TrimRight(sc.Text(), "\r")
		if lineNo == 1 {
			text = strings.TrimPrefix(text, "\xef\xbb\xbf")
		}
		if ended {
			return nil, fmt.Errorf("gedcom7: line %d: after TRLR", lineNo)
		}
		n, err := g.Match("Line", text+"\n")
		if err != nil {
			return nil, fmt.Errorf("gedcom7: line %d: %w", lineNo, err)
		}
		level, _ := strconv.Atoi(n.Find("Level").Text)
		s := &Structure{Tag: n.Find("Tag").Text}
		if x := n.Find("Xref"); x != nil {
			s.XRef = x.Text
		}
		if v := n.Find("LineVal"); v != nil {
			s.Payload = v.Text
		}
		if level > len(stack) {
			return nil, fmt.Errorf("gedcom7: line %d: level %d after level %d", lineNo, level, len(stack)-1)
		}
		stack = stack[:level]
		if s.Tag == "CONT" && level > 0 {
			parent := stack[level-1]
			parent.Payload += "\n" + s.Payload
			continue
		}
		switch {
		case level > 0:
			parent := stack[level-1]
			parent.Substructures = append(parent.Substructures, s)
		case s.Tag == "HEAD":
			if lineNo != 1 {
				return nil, fmt.Errorf("gedcom7: line %d: HEAD must be the first line", lineNo)
			}
			doc.Header = s
		case s.Tag == "TRLR":
			ended = true
		default:
			if doc.Header == nil {
				return nil, fmt.Errorf("gedcom7: line %d: no HEAD before the first record", lineNo)
			}
			doc.Records = append(doc.Records, s)
		}
		stack = append(stack, s)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if doc.Header == nil {
		return nil, fmt.Errorf("gedcom7: no HEAD")
	}
	if !ended {
		return nil, fmt.Errorf("gedcom7: no TRLR")
	}
	return doc, nil
}
