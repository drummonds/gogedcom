package convert

// The rules of docs/convert-7-to-551.md that rules.go does not name: what
// To551 does with a 7 structure. The rules both tables share (reader,
// header, version, exid, tran, role, enumeration, media type, medi,
// citation, date, age, name, address, unchanged) keep their names; each
// table gives its own meaning, the inverse of the other's.
const (
	note       Rule = "note"        // an SNOTE record, and a pointer to one, become NOTE
	schema     Rule = "schema"      // SCHMA is consumed: its TAGs drive the relocation back, nothing is written
	wac        Rule = "wac"         // INIL becomes WAC
	inlineObje Rule = "inline obje" // an OBJE record exactly one OBJE pointer names comes back inline
	extension  Rule = "extension"   // a tag 5.5.1 does not define becomes _TAG, everything under it unchanged
)

// tagRules551 is every tag the 7 registry has a structure term for, in
// the places 7 gives it, with its rule, keyed as tagRules is. A child row
// named after its parent's rule (EXID.TYPE, TRAN.LANG, DATE.PHRASE,
// NAME.PHRASE) is consumed by the parent; it is reached only under a
// parent that did not consume it, and then it is unchanged. The tag table
// in docs/convert-7-to-551.md is this map, row for row;
// TestRules551MatchTheDocument holds them together.
var tagRules551 = map[string]Rule{
	"HEAD":  header,
	"TRLR":  reader,
	"INDI":  unchanged,
	"FAM":   unchanged,
	"SUBM":  unchanged,
	"REPO":  unchanged,
	"SOUR":  unchanged,
	"OBJE":  unchanged,
	"SNOTE": note,

	"*.CONT": reader,

	"HEAD.SCHMA": schema,
	"SCHMA.TAG":  schema,
	"GEDC.VERS":  version,
	"*.GEDC":     unchanged,
	"*.VERS":     unchanged,

	"*.SNOTE":   note,
	"*.NOTE":    unchanged,
	"*.EXID":    exid,
	"EXID.TYPE": exid,
	"*.TRAN":    tran,
	"TRAN.LANG": tran,
	"*.LANG":    unchanged,
	"*.ROLE":    role,
	"*.INIL":    wac,

	"*.SEX":     enumeration,
	"*.PEDI":    enumeration,
	"*.RESN":    enumeration,
	"FAMC.STAT": enumeration,
	"*.STAT":    enumeration,
	"*.QUAY":    enumeration,
	"*.MEDI":    enumeration,
	"FAMC.ADOP": enumeration,
	"*.ADOP":    unchanged,
	"NAME.TYPE": enumeration,
	"*.TYPE":    unchanged,

	"FILE.FORM": mediaType,
	"FORM.MEDI": medi,
	"*.FORM":    unchanged,
	"*.FILE":    unchanged,
	"*.OBJE":    inlineObje,
	"*.SOUR":    citation,

	"*.DATE":      date,
	"DATE.PHRASE": date,
	"*.AGE":       age,
	"*.NAME":      name,
	"NAME.PHRASE": name,
	"*.PHRASE":    extension,
	"*.ADDR":      address,

	"*.CREA":   extension,
	"*.CROP":   extension,
	"*.HEIGHT": extension,
	"*.LEFT":   extension,
	"*.TOP":    extension,
	"*.WIDTH":  extension,
	"*.MIME":   extension,
	"*.NO":     extension,
	"*.SDATE":  extension,
	"*.UID":    extension,

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
	"*.COPR":  unchanged,
	"*.CORP":  unchanged,
	"*.CREM":  unchanged,
	"*.CTRY":  unchanged,
	"*.DATA":  unchanged,
	"*.DEAT":  unchanged,
	"*.DESI":  unchanged,
	"*.DEST":  unchanged,
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

// ruleFor551 is the rule for tag under the superstructure with tag super
// ("" at record level), and the key that gave it.
func ruleFor551(super, tag string) (Rule, string) {
	if super == "" {
		if r, ok := tagRules551[tag]; ok {
			return r, tag
		}
		return unchanged, tag
	}
	if r, ok := tagRules551[super+"."+tag]; ok {
		return r, super + "." + tag
	}
	if r, ok := tagRules551["*."+tag]; ok {
		return r, "*." + tag
	}
	return unchanged, "*." + tag
}

// enumerationSets is the 7 set an enumeration row's payload is judged
// against, by key, and whether 5.5.1 spells the set in lower case.
var enumerationSets = map[string]struct {
	set   string
	lower bool
}{
	"*.SEX":     {"enumset-SEX", false},
	"*.PEDI":    {"enumset-PEDI", true},
	"*.RESN":    {"enumset-RESN", true},
	"FAMC.STAT": {"enumset-FAMC-STAT", true},
	"*.STAT":    {"enumset-ord-STAT", false},
	"*.QUAY":    {"enumset-QUAY", false},
	"*.MEDI":    {"enumset-MEDI", true},
	"FORM.MEDI": {"enumset-MEDI", true},
	"FAMC.ADOP": {"enumset-ADOP", false},
	"NAME.TYPE": {"enumset-NAME-TYPE", true},
}

// transliterations is the guide's table backwards: a TRAN's LANG to the
// 5.5.1 tag and its TYPE.
var transliterations = map[string]struct{ tag, word string }{
	"ko-hang":          {"FONE", "hangul"},
	"ja-hrkt":          {"FONE", "kana"},
	"und-Latn-pinyin":  {"ROMN", "pinyin"},
	"ja-Latn":          {"ROMN", "romaji"},
	"zh-Latn-wadegile": {"ROMN", "wadegiles"},
}

// mediaWords is mediaTypes backwards, one word per type: jpg for
// image/jpeg, tif for image/tiff.
var mediaWords = map[string]string{
	"image/bmp":            "bmp",
	"image/gif":            "gif",
	"image/jpeg":           "jpg",
	"application/ole":      "ole",
	"image/vnd.zbrush.pcx": "pcx",
	"application/pdf":      "pdf",
	"image/png":            "png",
	"image/tiff":           "tif",
	"text/plain":           "txt",
	"audio/wav":            "wav",
}

// renamedBack is renamed backwards.
var renamedBack = map[string]string{
	"DNS_CAN":  "DNS/CAN",
	"PRE_1970": "PRE-1970",
}
