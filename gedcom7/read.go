package gedcom7

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"git.bytestone.uk/hum3/gotreesitter"
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

// Read reads a file with the grammar. A leading byte order mark is
// accepted; the grammar accepts CR, LF and CR LF line ends. The first line
// that the grammar rejects is the error.
//
// Each line is parsed on its own. The grammar is line-oriented, and
// gotreesitter (v0.7.0) flattens a repetition into its parent on every
// reduce, so parsing a file whole costs memory quadratic in its lines: a
// 100k-line export took 28 GB.
func Read(r io.Reader) (*Document, error) {
	src, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	src = bytes.TrimPrefix(src, []byte("\xef\xbb\xbf"))
	lang := Gedcom7Language()
	parser := gotreesitter.NewParser(lang)

	doc := &Document{}
	var stack []*Structure // stack[i] is the open structure at level i
	ended := false
	lineNo := 0
	for rest := src; len(rest) > 0; {
		lineNo++
		line, eol := nextLine(rest)
		rest = rest[len(line)+eol:]
		if len(line) == 0 {
			return nil, fmt.Errorf("gedcom7: line %d: empty line", lineNo)
		}
		if ended {
			return nil, fmt.Errorf("gedcom7: line %d: after TRLR", lineNo)
		}
		s, level, err := parseLine(parser, lang, line, lineNo)
		if err != nil {
			return nil, err
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
	if doc.Header == nil {
		return nil, fmt.Errorf("gedcom7: no HEAD")
	}
	if !ended {
		return nil, fmt.Errorf("gedcom7: no TRLR")
	}
	return doc, nil
}

// nextLine is the first line of src without its line end, and the length
// of that line end: 0 at the end of input.
func nextLine(src []byte) (line []byte, eol int) {
	i := bytes.IndexAny(src, "\r\n")
	switch {
	case i < 0:
		return src, 0
	case src[i] == '\r' && i+1 < len(src) && src[i+1] == '\n':
		return src[:i], 2
	default:
		return src[:i], 1
	}
}

// parseLine parses one line with the grammar and returns its structure and
// level.
func parseLine(parser *gotreesitter.Parser, lang *gotreesitter.Language, line []byte, lineNo int) (*Structure, int, error) {
	// gotreesitter (v0.7.0) mis-parses input that ends right after a GLR
	// fork; the grammar allows a trailing EOL.
	src := append(line[:len(line):len(line)], '\n')
	tree, err := parser.Parse(src)
	if err != nil {
		return nil, 0, fmt.Errorf("gedcom7: line %d: %w", lineNo, err)
	}
	defer tree.Release()
	root := tree.RootNode()
	if e := firstError(root); e != nil {
		return nil, 0, fmt.Errorf("gedcom7: %w", errorAt(e, lang, lineNo))
	}
	l := child(root, lang, "line")
	if l == nil {
		return nil, 0, fmt.Errorf("gedcom7: line %d: not a line", lineNo)
	}
	level, _ := strconv.Atoi(child(l, lang, "level").Text(src))
	s := &Structure{}
	if x := child(l, lang, "xref"); x != nil {
		s.XRef = x.Text(src)
	}
	// The payload line is the last named child: tag_line or a typed
	// *_line; its text is "TAG" or "TAG payload".
	payloadLine := l.NamedChild(l.NamedChildCount() - 1)
	s.Tag, s.Payload, _ = strings.Cut(payloadLine.Text(src), " ")
	return s, level, nil
}
