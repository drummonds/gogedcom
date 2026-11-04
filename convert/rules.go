package convert

// Rule is a row of the decision table in docs/convert-551-to-7.md: what
// To7 does with a 5.5.1 structure. The rules below are keyed by tag; the
// rules a condition other than the tag selects (an extension tag, a tag
// or identifier 7 rejects, a standard tag out of its place) are applied
// around them in convert.go.
type Rule string

const (
	reader      Rule = "reader"      // never reaches To7: CONC and CONT are folded, TRLR ends the file
	header      Rule = "header"      // HEAD: the rules of its substructures, then SCHMA after GEDC
	unchanged   Rule = "unchanged"   // tag and payload as the file has them
	dropped     Rule = "dropped"     // removed from 7, or a constant of the 5.5.1 format
	version     Rule = "version"     // GEDC.VERS becomes 7.0
	snote       Rule = "snote"       // a NOTE record, and a pointer to one, become SNOTE
	exid        Rule = "exid"        // RFN, AFN, RIN become EXID with TYPE
	tran        Rule = "tran"        // FONE, ROMN become TRAN with LANG
	role        Rule = "role"        // RELA and a citation ROLE: 7's ROLE set, else OTHER with PHRASE
	inil        Rule = "inil"        // WAC becomes INIL
	enumeration Rule = "enumeration" // upper-cased when that is in 7's set; DNS/CAN and PRE-1970 renamed
	mediaType   Rule = "media type"  // a FILE's FORM becomes a media type
	medi        Rule = "medi"        // a FORM's TYPE becomes MEDI, an enumeration
	objeRecord  Rule = "obje record" // an inline OBJE becomes a record
	citation    Rule = "citation"    // an inline SOUR becomes SOUR @VOID@, or a record when described
	date        Rule = "date"        // 7's spelling with PHRASE the original when it differs; no payload when 7 rejects it
	age         Rule = "age"         // CHILD, INFANT, STILLBORN as bounds; else as date
	name        Rule = "name"        // slashes and tabs to spaces when 7 rejects the name, with _PHRASE
	address     Rule = "address"     // an empty ADDR payload composed from its pieces
)

// tagRules is every tag of the 5.5.1 specification's Appendix A in the
// places 5.5.1 gives it, with its rule; WAC is the guide's addition. A key
// is "TAG" at record level, "SUPER.TAG" under one superstructure, "*.TAG"
// under any; ruleFor tries them in that order, and a tag with no entry is
// unchanged. The tag table in docs/convert-551-to-7.md is this map, row for
// row; TestRulesMatchTheDocument holds them together.
var tagRules = map[string]Rule{
	"HEAD": header,
	"TRLR": reader,
	"INDI": unchanged,
	"FAM":  unchanged,
	"SUBM": unchanged,
	"REPO": unchanged,
	"SOUR": unchanged,
	"OBJE": unchanged,
	"NOTE": snote,
	"SUBN": dropped,

	"*.CONC": reader,
	"*.CONT": reader,

	"HEAD.CHAR": dropped,
	"HEAD.FILE": dropped,
	"HEAD.SUBN": dropped,
	"HEAD.SOUR": unchanged,
	"GEDC.FORM": dropped,
	"GEDC.VERS": version,
	"*.GEDC":    unchanged,
	"*.VERS":    unchanged,
	"*.DEST":    unchanged,
	"*.COPR":    unchanged,
	"*.CORP":    unchanged,
	"*.DATA":    unchanged,
	"*.LANG":    unchanged,

	"SUBN.ANCE": dropped,
	"SUBN.DESC": dropped,
	"SUBN.ORDI": dropped,
	"SUBN.FAMF": dropped,
	"SUBN.TEMP": dropped,

	"*.NOTE": snote,
	"*.RFN":  exid,
	"*.AFN":  exid,
	"*.RIN":  exid,
	"*.FONE": tran,
	"*.ROMN": tran,
	"*.RELA": role,
	"*.ROLE": role,
	"*.WAC":  inil,

	"*.SEX":     enumeration,
	"*.PEDI":    enumeration,
	"*.RESN":    enumeration,
	"*.STAT":    enumeration,
	"*.QUAY":    enumeration,
	"*.MEDI":    enumeration,
	"FAMC.ADOP": enumeration,
	"*.ADOP":    unchanged,
	"NAME.TYPE": enumeration,
	"FORM.TYPE": medi,
	"*.TYPE":    unchanged,

	"FILE.FORM": mediaType,
	"*.FORM":    unchanged,
	"*.FILE":    unchanged,
	"*.OBJE":    objeRecord,
	"*.SOUR":    citation,

	"*.DATE": date,
	"*.AGE":  age,
	"*.NAME": name,
	"*.ADDR": address,

	"*.ABBR":  unchanged,
	"*.ADR1":  unchanged,
	"*.ADR2":  unchanged,
	"*.ADR3":  unchanged,
	"*.AGNC":  unchanged,
	"*.ALIA":  unchanged,
	"*.ANCI":  unchanged,
	"*.ANUL":  unchanged,
	"*.ASSO":  unchanged,
	"*.AUTH":  unchanged,
	"*.BAPL":  unchanged,
	"*.BAPM":  unchanged,
	"*.BARM":  unchanged,
	"*.BASM":  unchanged,
	"*.BIRT":  unchanged,
	"*.BLES":  unchanged,
	"*.BURI":  unchanged,
	"*.CALN":  unchanged,
	"*.CAST":  unchanged,
	"*.CAUS":  unchanged,
	"*.CENS":  unchanged,
	"*.CHAN":  unchanged,
	"*.CHIL":  unchanged,
	"*.CHR":   unchanged,
	"*.CHRA":  unchanged,
	"*.CITY":  unchanged,
	"*.CONF":  unchanged,
	"*.CONL":  unchanged,
	"*.CREM":  unchanged,
	"*.CTRY":  unchanged,
	"*.DEAT":  unchanged,
	"*.DESI":  unchanged,
	"*.DIV":   unchanged,
	"*.DIVF":  unchanged,
	"*.DSCR":  unchanged,
	"*.EDUC":  unchanged,
	"*.EMAIL": unchanged,
	"*.EMIG":  unchanged,
	"*.ENDL":  unchanged,
	"*.ENGA":  unchanged,
	"*.EVEN":  unchanged,
	"*.FACT":  unchanged,
	"*.FAMC":  unchanged,
	"*.FAMS":  unchanged,
	"*.FAX":   unchanged,
	"*.FCOM":  unchanged,
	"*.GIVN":  unchanged,
	"*.GRAD":  unchanged,
	"*.HUSB":  unchanged,
	"*.IDNO":  unchanged,
	"*.IMMI":  unchanged,
	"*.LATI":  unchanged,
	"*.LONG":  unchanged,
	"*.MAP":   unchanged,
	"*.MARB":  unchanged,
	"*.MARC":  unchanged,
	"*.MARL":  unchanged,
	"*.MARR":  unchanged,
	"*.MARS":  unchanged,
	"*.NATI":  unchanged,
	"*.NATU":  unchanged,
	"*.NCHI":  unchanged,
	"*.NICK":  unchanged,
	"*.NMR":   unchanged,
	"*.NPFX":  unchanged,
	"*.NSFX":  unchanged,
	"*.OCCU":  unchanged,
	"*.ORDN":  unchanged,
	"*.PAGE":  unchanged,
	"*.PHON":  unchanged,
	"*.PLAC":  unchanged,
	"*.POST":  unchanged,
	"*.PROB":  unchanged,
	"*.PROP":  unchanged,
	"*.PUBL":  unchanged,
	"*.REFN":  unchanged,
	"*.RELI":  unchanged,
	"*.REPO":  unchanged,
	"*.RESI":  unchanged,
	"*.RETI":  unchanged,
	"*.SLGC":  unchanged,
	"*.SLGS":  unchanged,
	"*.SPFX":  unchanged,
	"*.SSN":   unchanged,
	"*.STAE":  unchanged,
	"*.SUBM":  unchanged,
	"*.SURN":  unchanged,
	"*.TEMP":  unchanged,
	"*.TEXT":  unchanged,
	"*.TIME":  unchanged,
	"*.TITL":  unchanged,
	"*.WIFE":  unchanged,
	"*.WILL":  unchanged,
	"*.WWW":   unchanged,
}

// ruleFor is the rule for tag under the superstructure with tag super, ""
// at record level.
func ruleFor(super, tag string) Rule {
	if super == "" {
		if r, ok := tagRules[tag]; ok {
			return r
		}
		return unchanged
	}
	if r, ok := tagRules[super+"."+tag]; ok {
		return r
	}
	if r, ok := tagRules["*."+tag]; ok {
		return r
	}
	return unchanged
}

// languages is the migration guide's table: a FONE or ROMN TYPE to the
// TRAN's LANG.
var languages = map[string]string{
	"hangul":    "ko-hang",
	"kana":      "ja-hrkt",
	"pinyin":    "und-Latn-pinyin",
	"romaji":    "ja-Latn",
	"wadegiles": "zh-Latn-wadegile",
}

// mediaTypes is a 5.5.1 FORM word to 7's media type: the guide's table
// with jpg, png, pdf and txt, which the Geni export writes.
var mediaTypes = map[string]string{
	"bmp":  "image/bmp",
	"gif":  "image/gif",
	"jpg":  "image/jpeg",
	"jpeg": "image/jpeg",
	"ole":  "application/ole",
	"pcx":  "image/vnd.zbrush.pcx",
	"pdf":  "application/pdf",
	"png":  "image/png",
	"tif":  "image/tiff",
	"tiff": "image/tiff",
	"txt":  "text/plain",
	"wav":  "audio/wav",
}

// ages is the 5.5.1 age words as 7 bounds.
var ages = map[string]string{
	"CHILD":     "< 8y",
	"INFANT":    "< 1y",
	"STILLBORN": "0y",
}

// calendars is a 5.5.1 calendar escape's name as 7 spells it: the guide
// makes ROMAN and UNKNOWN extensions.
var calendars = map[string]string{
	"GREGORIAN": "GREGORIAN",
	"JULIAN":    "JULIAN",
	"HEBREW":    "HEBREW",
	"FRENCH R":  "FRENCH_R",
	"ROMAN":     "_ROMAN",
	"UNKNOWN":   "_UNKNOWN",
}

// renamed is the enumeration values 7 spells differently.
var renamed = map[string]string{
	"DNS/CAN":  "DNS_CAN",
	"PRE-1970": "PRE_1970",
}
