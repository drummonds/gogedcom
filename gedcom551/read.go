// Package gedcom551 reads GEDCOM 5.5.1 with the grammar under
// grammar/gedcom551, written from the 5.5.1 specification. The file is
// the contract: a line types nothing, so every payload is kept as the
// file has it, and a data type is parsed on demand (ParseDate).
package gedcom551

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"git.bytestone.uk/hum3/gotreesitter"

	"git.bytestone.uk/hum3/gogedcom/gedcom"
	"git.bytestone.uk/hum3/gogedcom/internal/tstree"
)

// Read reads a file with the grammar into the tree the gedcom7 reader
// returns. A leading byte order mark, leading white space on a line and
// empty lines are ignored, as the specification instructs readers; the
// four terminators (CR, LF, CR LF, LF CR) are accepted. CONC lines are
// folded into their parent's payload as written and CONT lines with a
// line break; nothing else in a payload is interpreted. The first line
// that the grammar rejects is the error.
//
// Each line is parsed on its own, as the 7 reader does: gotreesitter
// (v0.7.0) flattens a repetition into its parent on every reduce, so a
// file parsed whole costs memory quadratic in its lines.
func Read(r io.Reader) (*gedcom.Document, error) {
	src, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	src = bytes.TrimPrefix(src, []byte("\xef\xbb\xbf"))
	lang := Gedcom551Language()
	parser := gotreesitter.NewParser(lang)

	doc := &gedcom.Document{}
	var stack []*gedcom.Structure // stack[i] is the open structure at level i
	ended := false
	lineNo := 0
	for rest := src; len(rest) > 0; {
		lineNo++
		line, eol := nextLine(rest)
		rest = rest[len(line)+eol:]
		line = bytes.TrimLeft(line, " \t")
		if len(line) == 0 {
			continue
		}
		if ended {
			return nil, fmt.Errorf("gedcom551: line %d: after TRLR", lineNo)
		}
		s, level, err := parseLine(parser, lang, line, lineNo)
		if err != nil {
			return nil, err
		}
		if level > len(stack) {
			return nil, fmt.Errorf("gedcom551: line %d: level %d after level %d", lineNo, level, len(stack)-1)
		}
		stack = stack[:level]
		if level > 0 && (s.Tag == "CONC" || s.Tag == "CONT") {
			parent := stack[level-1]
			if s.Tag == "CONT" {
				parent.Payload += "\n"
			}
			parent.Payload += s.Payload
			continue
		}
		switch {
		case level > 0:
			parent := stack[level-1]
			parent.Substructures = append(parent.Substructures, s)
		case s.Tag == "HEAD":
			if doc.Header != nil {
				return nil, fmt.Errorf("gedcom551: line %d: second HEAD", lineNo)
			}
			if len(doc.Records) > 0 {
				return nil, fmt.Errorf("gedcom551: line %d: HEAD after a record", lineNo)
			}
			doc.Header = s
		case s.Tag == "TRLR":
			ended = true
		default:
			if doc.Header == nil {
				return nil, fmt.Errorf("gedcom551: line %d: no HEAD before the first record", lineNo)
			}
			doc.Records = append(doc.Records, s)
		}
		stack = append(stack, s)
	}
	if doc.Header == nil {
		return nil, fmt.Errorf("gedcom551: no HEAD")
	}
	if !ended {
		return nil, fmt.Errorf("gedcom551: no TRLR")
	}
	return doc, nil
}

// nextLine is the first line of src without its terminator, and the
// length of that terminator: CR, LF, CR LF or LF CR; 0 at the end of
// input.
func nextLine(src []byte) (line []byte, eol int) {
	i := bytes.IndexAny(src, "\r\n")
	switch {
	case i < 0:
		return src, 0
	case i+1 < len(src) && src[i+1] != src[i] && (src[i+1] == '\r' || src[i+1] == '\n'):
		return src[:i], 2
	default:
		return src[:i], 1
	}
}

// parseLine parses one line with the grammar and returns its structure and
// level.
func parseLine(parser *gotreesitter.Parser, lang *gotreesitter.Language, line []byte, lineNo int) (*gedcom.Structure, int, error) {
	// gotreesitter (v0.7.0) mis-parses input that ends right after a GLR
	// fork; the grammar allows a trailing terminator.
	src := append(line[:len(line):len(line)], '\n')
	tree, err := parser.Parse(src)
	if err != nil {
		return nil, 0, fmt.Errorf("gedcom551: line %d: %w", lineNo, err)
	}
	defer tree.Release()
	root := tree.RootNode()
	if err := tstree.ParseError(root, lang, lineNo); err != nil {
		return nil, 0, fmt.Errorf("gedcom551: %w", err)
	}
	l := tstree.Child(root, lang, "line")
	if l == nil {
		return nil, 0, fmt.Errorf("gedcom551: line %d: not a line", lineNo)
	}
	level, _ := strconv.Atoi(tstree.Child(l, lang, "level").Text(src))
	s := &gedcom.Structure{Line: lineNo}
	if x := tstree.Child(l, lang, "xref"); x != nil {
		s.XRef = x.Text(src)
	}
	s.Tag = tstree.Child(l, lang, "tag").Text(src)
	// The value, pointer or text, is the rest of the line after the
	// delimiter: the file's text, verbatim.
	if v := tstree.Child(l, lang, "pointer"); v != nil {
		s.Payload = v.Text(src)
	} else if v := tstree.Child(l, lang, "text"); v != nil {
		s.Payload = v.Text(src)
	}
	return s, level, nil
}
