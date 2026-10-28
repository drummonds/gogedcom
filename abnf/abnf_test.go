package abnf

import (
	"strings"
	"testing"
)

func mustParse(t *testing.T, src string) *Grammar {
	t.Helper()
	g, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return g
}

func TestMatchConstructs(t *testing.T) {
	cases := []struct {
		name    string
		grammar string
		input   string
		ok      bool
	}{
		{"bare string is case-insensitive", `a = "ab"`, "AB", true},
		{"%s string is case-sensitive", `a = %s"ab"`, "AB", false},
		{"%s string exact", `a = %s"ab"`, "ab", true},
		{"%i string", `a = %i"ab"`, "Ab", true},
		{"numeric range in", `a = %x41-43`, "B", true},
		{"numeric range out", `a = %x41-43`, "D", false},
		{"numeric sequence", `a = %x0D.0A`, "\r\n", true},
		{"decimal value", `a = %d65`, "A", true},
		{"binary value", `a = %b01000001`, "A", true},
		{"repeat n*m low", `a = 2*3DIGIT`, "12", true},
		{"repeat n*m high", `a = 2*3DIGIT`, "123", true},
		{"repeat n*m too few", `a = 2*3DIGIT`, "1", false},
		{"repeat n*m too many", `a = 2*3DIGIT`, "1234", false},
		{"repeat exact", `a = 3DIGIT`, "123", true},
		{"repeat exact wrong", `a = 3DIGIT`, "12", false},
		{"repeat star then literal", `a = *DIGIT "x"`, "12x", true},
		{"repeat star empty", `a = *DIGIT "x"`, "x", true},
		{"repeat 1*", `a = 1*DIGIT`, "", false},
		{"house-style pipe alternation", "a = b | c\nb = \"1\"\nc = \"2\"", "2", true},
		{"slash alternation", "a = b / c\nb = \"1\"\nc = \"2\"", "2", true},
		{"full backtracking over alternatives", `a = "1" / "12"`, "12", true},
		{"backtracking into repetition", `a = *DIGIT "1"`, "121", true},
		{"incremental alternative", "a = b\nb = \"y\"\nb =/ \"z\"", "z", true},
		{"optional absent", `a = [ "x" ] "y"`, "y", true},
		{"optional present", `a = [ "x" ] "y"`, "xy", true},
		{"group", `a = ("x" / "y") "z"`, "yz", true},
		{"continuation line", "a = \"x\"\n    \"y\"", "xy", true},
		{"comment", "a = \"x\" ; a comment\n", "x", true},
		{"rule names are case-insensitive", "a = digit\n", "7", true},
		{"core rule ALPHA", `a = 1*ALPHA`, "abcXYZ", true},
		{"core rule overridden", "a = DIGIT\nDIGIT = \"q\"", "q", true},
		{"unicode code point", `a = %x2192`, "→", true},
		{"unicode in string", `a = %s"→"`, "→", true},
		{"input left over", `a = "x"`, "xy", false},
		{"empty rule matches empty", `a = [ "x" ]`, "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := mustParse(t, c.grammar)
			_, err := g.Match("a", c.input)
			if (err == nil) != c.ok {
				t.Fatalf("Match(%q) err = %v, want ok=%v", c.input, err, c.ok)
			}
		})
	}
}

func TestMatchTree(t *testing.T) {
	g := mustParse(t, "a = b c\nb = \"1\"\nc = *DIGIT\n")
	n, err := g.Match("a", "123")
	if err != nil {
		t.Fatal(err)
	}
	if n.Rule != "a" || n.Start != 0 || n.End != 3 {
		t.Fatalf("root = %+v", n)
	}
	if len(n.Children) != 2 || n.Children[0].Rule != "b" || n.Children[1].Rule != "c" {
		t.Fatalf("children = %+v", n.Children)
	}
	if got := n.Find("c").Text; got != "23" {
		t.Fatalf("c text = %q", got)
	}
	if n.Find("zzz") != nil {
		t.Fatal("Find of an absent rule should be nil")
	}
	if got := len(n.Find("c").All("DIGIT")); got != 2 {
		t.Fatalf("DIGIT nodes under c = %d", got)
	}
}

func TestMatchTreeUsesDefinedCase(t *testing.T) {
	g := mustParse(t, "a = Foo\nfoo = \"x\"\n")
	n, err := g.Match("a", "x")
	if err != nil {
		t.Fatal(err)
	}
	if n.Children[0].Rule != "foo" {
		t.Fatalf("node rule = %q, want the name as defined", n.Children[0].Rule)
	}
}

func TestUndefined(t *testing.T) {
	g := mustParse(t, "a = b c DIGIT\nc = \"1\"\n")
	if got := g.Undefined(); len(got) != 1 || got[0] != "b" {
		t.Fatalf("Undefined = %v, want [b]", got)
	}
	if _, err := g.Match("a", "1"); err == nil || !strings.Contains(err.Error(), "b") {
		t.Fatalf("Match with an undefined rule: err = %v", err)
	}
	if _, err := g.Match("nosuch", "1"); err == nil {
		t.Fatal("Match of an unknown rule should error")
	}
}

func TestMatchErrorPosition(t *testing.T) {
	g := mustParse(t, `a = "ab" "cd"`)
	_, err := g.Match("a", "abxd")
	if err == nil || !strings.Contains(err.Error(), "2") {
		t.Fatalf("err = %v, want the furthest position 2", err)
	}
}

func TestParseErrors(t *testing.T) {
	for _, src := range []string{
		`a = `,
		`a = "unterminated`,
		`a = %x4G`,
		`= "x"`,
		`a = ( "x"`,
		`a = 3*2DIGIT`,
	} {
		if _, err := Parse(src); err == nil {
			t.Errorf("Parse(%q) should fail", src)
		}
	}
}

func TestRules(t *testing.T) {
	g := mustParse(t, "b = \"1\"\na = b\n")
	if got := g.Rules(); len(got) != 2 || got[0] != "b" || got[1] != "a" {
		t.Fatalf("Rules = %v, want definition order", got)
	}
}

func TestMatchWhere(t *testing.T) {
	g := mustParse(t, "a = b / c\nb = 1*DIGIT\nc = \"4\" 1*DIGIT\n")
	notFour := func(n *Node) bool { return n.Rule != "b" || !strings.HasPrefix(n.Text, "4") }
	n, err := g.MatchWhere("a", "42", notFour)
	if err != nil {
		t.Fatal(err)
	}
	if n.Children[0].Rule != "c" {
		t.Fatalf("derivation = %s, want c after b was rejected", n.Children[0].Rule)
	}
	if _, err := g.MatchWhere("a", "4", notFour); err == nil {
		t.Fatal("4 has no accepted derivation")
	}
}
