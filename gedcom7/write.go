package gedcom7

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// The grammar's Tag and Xref rules.
var (
	tagRe  = regexp.MustCompile(`^([A-Z]|_)[A-Z0-9_]*$`)
	xrefRe = regexp.MustCompile(`^@[A-Z0-9_]+@$`)
)

// Write writes the document as GEDCOM 7: the header, the records in order,
// then TRLR, with LF line ends and no byte order mark. A line break in a
// payload becomes a CONT line; a text payload beginning with "@" is
// escaped as "@@". The tree is written as it is; Validate judges it.
//
// Write refuses what no line can carry, naming the record: a carriage
// return in a payload, a tag or cross-reference identifier the grammar
// rejects, a document with no header.
func Write(w io.Writer, doc *Document) error {
	if doc.Header == nil {
		return fmt.Errorf("gedcom7: no HEAD")
	}
	bw := bufio.NewWriter(w)
	wr := &writer{w: bw}
	wr.record(doc.Header)
	for _, r := range doc.Records {
		wr.record(r)
	}
	wr.line(0, "", "TRLR", "")
	if wr.err != nil {
		return wr.err
	}
	return bw.Flush()
}

type writer struct {
	w   *bufio.Writer
	at  string // "@I1@ INDI": the record being written, for errors
	err error
}

func (wr *writer) fail(format string, args ...any) {
	if wr.err == nil {
		wr.err = fmt.Errorf("gedcom7: %s: %s", wr.at, fmt.Sprintf(format, args...))
	}
}

func (wr *writer) record(r *Structure) {
	wr.at = strings.TrimSpace(r.XRef + " " + r.Tag)
	if r.XRef != "" && (!xrefRe.MatchString(r.XRef) || r.XRef == "@VOID@") {
		wr.fail("%q is not a cross-reference identifier", r.XRef)
		return
	}
	wr.structure(0, r)
}

func (wr *writer) structure(level int, s *Structure) {
	if wr.err != nil {
		return
	}
	if !tagRe.MatchString(s.Tag) {
		wr.fail("%q is not a tag", s.Tag)
		return
	}
	if strings.Contains(s.Payload, "\r") {
		wr.fail("carriage return in payload")
		return
	}
	first, rest, more := strings.Cut(s.Payload, "\n")
	xref := ""
	if level == 0 {
		xref = s.XRef
	}
	wr.line(level, xref, s.Tag, first)
	for more {
		var cont string
		cont, rest, more = strings.Cut(rest, "\n")
		wr.line(level+1, "", "CONT", cont)
	}
	for _, c := range s.Substructures {
		wr.structure(level+1, c)
	}
}

// line writes "level [xref] tag [payload]", escaping a leading at-sign in
// a text payload. A pointer is written as it is.
func (wr *writer) line(level int, xref, tag, payload string) {
	if wr.err != nil {
		return
	}
	wr.w.WriteString(strconv.Itoa(level))
	wr.w.WriteByte(' ')
	if xref != "" {
		wr.w.WriteString(xref)
		wr.w.WriteByte(' ')
	}
	wr.w.WriteString(tag)
	if payload != "" {
		wr.w.WriteByte(' ')
		if strings.HasPrefix(payload, "@") && !isPointer(payload) {
			wr.w.WriteByte('@')
		}
		wr.w.WriteString(payload)
	}
	wr.w.WriteByte('\n')
}

// isPointer is whether a payload is a pointer: "@I1@" or "@VOID@".
func isPointer(payload string) bool { return xrefRe.MatchString(payload) }
