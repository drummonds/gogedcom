package gedcom551

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"git.bytestone.uk/hum3/gogedcom/gedcom"
)

// The grammar's tag and xref rules.
var (
	tagRe  = regexp.MustCompile(`^[A-Za-z0-9_]+$`)
	xrefRe = regexp.MustCompile(`^@[A-Za-z0-9_][^@\r\n]*@$`)
)

// Write writes the document as GEDCOM 5.5.1 on gedcom7.Write's contract:
// the header, the records in order, then TRLR, with LF line ends and no
// byte order mark. A line break in a payload becomes a CONT line; every
// payload is otherwise written as it is, an at-sign included, since Read
// decodes none. Lines are not split with CONC. The tree is written as it
// is: the header's CHAR and GEDC.FORM are the converter's to add.
//
// Write refuses what no line can carry, naming the record: a carriage
// return in a payload, a tag or cross-reference identifier the grammar
// rejects, a document with no header.
func Write(w io.Writer, doc *gedcom.Document) error {
	if doc.Header == nil {
		return fmt.Errorf("gedcom551: no HEAD")
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
		wr.err = fmt.Errorf("gedcom551: %s: %s", wr.at, fmt.Sprintf(format, args...))
	}
}

func (wr *writer) record(r *gedcom.Structure) {
	wr.at = strings.TrimSpace(r.XRef + " " + r.Tag)
	if r.XRef != "" && !xrefRe.MatchString(r.XRef) {
		wr.fail("%q is not a cross-reference identifier", r.XRef)
		return
	}
	wr.structure(0, r)
}

func (wr *writer) structure(level int, s *gedcom.Structure) {
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

// line writes "level [xref] tag [payload]".
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
		wr.w.WriteString(payload)
	}
	wr.w.WriteByte('\n')
}
