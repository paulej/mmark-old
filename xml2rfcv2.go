// XML2RFC v2 rendering backend

package mmark

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
)

// References code in Xml2rfcv3.go

// XML renderer configuration options.
const (
	XML2_STANDALONE = 1 << iota // create standalone document
)

// Xml2 is a type that implements the Renderer interface for XML2RFV3 output.
//
// Do not create this directly, instead use the Xml2Renderer function.
type Xml2 struct {
	flags        int // XML_* options
	sectionLevel int // current section level
	docLevel     int // frontmatter/mainmatter or backmatter

	// Store the IAL we see for this block element
	ial []*IAL

	// TitleBlock in TOML
	titleBlock *title
}

// Xml2Renderer creates and configures a Xml object, which
// satisfies the Renderer interface.
//
// flags is a set of XML_* options ORed together
func Xml2Renderer(flags int) Renderer        { return &Xml2{flags: flags} }
func (options *Xml2) GetFlags() int          { return options.flags }
func (options *Xml2) GetState() int          { return 0 }
func (options *Xml2) SetIAL(i []*IAL)        { options.ial = append(options.ial, i...) }
func (options *Xml2) GetAndResetIAL() []*IAL { i := options.ial; options.ial = nil; return i }

// render code chunks using verbatim, or listings if we have a language
func (options *Xml2) BlockCode(out *bytes.Buffer, text []byte, lang string, caption []byte) {
	s := renderIAL(options.GetAndResetIAL())
	if lang == "" {
		out.WriteString("\n<figure" + s + "><artwork>\n")
	} else {
		out.WriteString("\n<figure" + s + "><artwork>\n")
	}
	out.Write(text)
	if lang == "" {
		out.WriteString("</artwork></figure>\n")
	} else {
		out.WriteString("</artwork></figure>\n")
	}
}

func (options *Xml2) TitleBlockTOML(out *bytes.Buffer, block *title) {
	if options.flags&XML_STANDALONE == 0 {
		return
	}
	options.titleBlock = block
	out.WriteString("<rfc ipr=\"" +
		options.titleBlock.Ipr + "\" category=\"" +
		options.titleBlock.Category + "\" docName=\"" + options.titleBlock.DocName + "\">\n")
	out.WriteString("<front>\n")
	out.WriteString("<title abbrev=\"" + options.titleBlock.Abbrev + "\">")
	out.WriteString(options.titleBlock.Title + "</title>\n\n")

	for _, a := range options.titleBlock.Author {
		out.WriteString("<author")
		out.WriteString(" initials=\"" + a.Initials + "\"")
		out.WriteString(" surname=\"" + a.Surname + "\"")
		out.WriteString(" fullname=\"" + a.Fullname + "\">\n")

		out.WriteString("<organization>" + a.Organization + "</organization>\n")
		out.WriteString("<address>\n")
		out.WriteString("<email>" + a.Address.Email + "</email>\n")
		out.WriteString("</address>\n")
		out.WriteString("</author>\n")
	}

	year := ""
	if options.titleBlock.Date.Year() > 0 {
		year = " year=\"" + strconv.Itoa(options.titleBlock.Date.Year()) + "\""
	}
	month := ""
	if options.titleBlock.Date.Month() > 0 {
		month = " month=\"" + time.Month(options.titleBlock.Date.Month()).String() + "\""
	}
	day := ""
	if options.titleBlock.Date.Day() > 0 {
		day = " day=\"" + strconv.Itoa(options.titleBlock.Date.Day()) + "\""
	}
	out.WriteString("<date" + year + month + day + "/>\n\n")

	out.WriteString("<area>" + options.titleBlock.Area + "</area>\n")
	out.WriteString("<workgroup>" + options.titleBlock.Workgroup + "</workgroup>\n")
	for _, k := range options.titleBlock.Keyword {
		out.WriteString("<keyword>" + k + "</keyword>\n")
	}
	out.WriteString("\n")
}

func (options *Xml2) BlockQuote(out *bytes.Buffer, text []byte) {
	renderIAL(options.GetAndResetIAL())
	// Fake a list paragraph
	out.WriteString("<t><list style=\"empty\">\n")
	out.Write(text)
	out.WriteString("</list></t>\n")
}

func (options *Xml2) Abstract(out *bytes.Buffer, text []byte) {
	s := renderIAL(options.GetAndResetIAL())
	out.WriteString("<abstract" + s + ">\n")
	out.Write(text)
	out.WriteString("</abstract>\n")
}

func (options *Xml2) Aside(out *bytes.Buffer, text []byte) {
	options.BlockQuote(out, text)
}

func (options *Xml2) Note(out *bytes.Buffer, text []byte) {
	options.BlockQuote(out, text)
}

func (options *Xml2) CommentHtml(out *bytes.Buffer, text []byte) {
	// nothing fancy any left of the first `:` will be used as the source="..."
	i := bytes.Index(text, []byte("-->"))
	if i > 0 {
		text = text[:i]
	}
	// strip, <!--
	text = text[4:]

	var source []byte
	l := len(text)
	if l > 20 {
		l = 20
	}
	for i := 0; i < l; i++ {
		if text[i] == ':' {
			source = text[:i]
			text = text[i+1:]
			break
		}
	}
	if len(source) != 0 {
		if source[0] == ' ' {
			source = source[1:]
		}
		out.WriteString("<t><cref source=\"")
		out.Write(source)
		out.WriteString("\">")
	} else {
		out.WriteString("<t><cref>\n")
	}
	out.Write(text)
	out.WriteString("</cref></t>\n")
	return
}

func (options *Xml2) BlockHtml(out *bytes.Buffer, text []byte) {
	// not supported, don't know yet if this is useful
	return
}

func (options *Xml2) Header(out *bytes.Buffer, text func() bool, level int, id string) {
	// set amount of open in options, so we know what to close after we finish
	// parsing the doc.
	//marker := out.Len()
	//out.Truncate(marker)
	if level <= options.sectionLevel {
		// close previous ones
		for i := options.sectionLevel - level + 1; i > 0; i-- {
			out.WriteString("</section>\n")
		}
	}
	// new section
	// Clashes with IAL, need to check ID
	renderIAL(options.GetAndResetIAL()) // Clear IAL here, so it will not pile up for following items
	out.WriteString("\n<section anchor=\"" + id + "\"")
	out.WriteString(" title=\"")
	text() // check bool here
	out.WriteString("\">\n")
	options.sectionLevel = level
	return
}

func (options *Xml2) HRule(out *bytes.Buffer) {
	// not used
}

func (options *Xml2) List(out *bytes.Buffer, text func() bool, flags, start int) {
	marker := out.Len()
	s := renderIAL(options.GetAndResetIAL())

	// inside lists we should drop the paragraph
	if flags&LIST_INSIDE_LIST == 0 {
		out.WriteString("<t>\n")
	}

	switch {
	case flags&LIST_TYPE_ORDERED != 0:
		if start <= 1 {
			out.WriteString("<list style=\"numbers\"" + s + ">\n")
		} else {
			out.WriteString(fmt.Sprintf("<list style=\"numbers\""+s+" start=\"%d\">\n", start))
		}
	case flags&LIST_TYPE_DEFINITION != 0:
		out.WriteString("<list style=\"hanging\"" + s + ">\n")
	default:
		out.WriteString("<list style=\"symbols\"" + s + ">\n")
	}

	if !text() {
		out.Truncate(marker)
		return
	}
	switch {
	case flags&LIST_TYPE_ORDERED != 0:
		out.WriteString("</list>\n")
	case flags&LIST_TYPE_DEFINITION != 0:
		out.WriteString("</t>\n</list>\n")
	default:
		out.WriteString("</list>\n")
	}
	if flags&LIST_INSIDE_LIST == 0 {
		out.WriteString("</t>\n")
	}
}

func (options *Xml2) ListItem(out *bytes.Buffer, text []byte, flags int) {
	if flags&LIST_TYPE_DEFINITION != 0 && flags&LIST_TYPE_TERM == 0 {
		//out.WriteString("<dd>")
		out.Write(text)
		//out.WriteString("</dd>\n")
		return
	}
	if flags&LIST_TYPE_TERM != 0 {
		if flags&LIST_ITEM_BEGINNING_OF_LIST == 0 {
			out.WriteString("</t>\n")
		}
		// close previous one?/
		out.WriteString("<t hangText=\"")
		out.Write(text)
		out.WriteString("\">\n")
		return
	}
	out.WriteString("<t>")
	out.Write(text)
	out.WriteString("</t>\n")
}

// Needs flags int, for in-list-detection xml2rfc v2
func (options *Xml2) Paragraph(out *bytes.Buffer, text func() bool, flags int) {
	marker := out.Len()
	if flags&LIST_TYPE_DEFINITION == 0 {
		out.WriteString("<t>")
	}
	if !text() {
		out.Truncate(marker)
		return
	}
	if flags&LIST_TYPE_DEFINITION == 0 {
		out.WriteString("</t>\n")
	}
}

func (options *Xml2) Table(out *bytes.Buffer, header []byte, body []byte, columnData []int, caption []byte) {
	s := renderIAL(options.GetAndResetIAL())
	out.WriteString("<texttable" + s + ">\n")
	out.Write(header)
	out.Write(body)
	out.WriteString("</texttable>\n")
}

func (options *Xml2) TableRow(out *bytes.Buffer, text []byte) {
	out.Write(text)
	out.WriteString("\n")
}

func (options *Xml2) TableHeaderCell(out *bytes.Buffer, text []byte, align int) {
	a := ""
	switch align {
	case TABLE_ALIGNMENT_LEFT:
		a = " align=\"left\""
	case TABLE_ALIGNMENT_RIGHT:
		a = " align=\"right\""
	default:
		a = " align=\"center\""
	}
	out.WriteString("<ttcol" + a + ">")
	out.Write(text)
	out.WriteString("</ttcol>\n")

}

func (options *Xml2) TableCell(out *bytes.Buffer, text []byte, align int) {
	out.WriteString("<c>")
	out.Write(text)
	out.WriteString("</c>")
}

func (options *Xml2) Footnotes(out *bytes.Buffer, text func() bool) {
	// not used
}

func (options *Xml2) FootnoteItem(out *bytes.Buffer, name, text []byte, flags int) {
	// not used
}

func (options *Xml2) Index(out *bytes.Buffer, primary, secondary []byte) {
	out.WriteString("<iref item=\"" + string(primary) + "\"")
	out.WriteString(" subitem=\"" + string(secondary) + "\"" + "/>")
}

func (options *Xml2) Citation(out *bytes.Buffer, link, title []byte) {
	if len(title) == 0 {
		out.WriteString("<xref target=\"" + string(link) + "\"/>")
		return
	}
	out.WriteString("<xref target=\"" + string(link) + "\"/>")
}

func (options *Xml2) References(out *bytes.Buffer, citations map[string]*citation) {
	if options.flags&XML_STANDALONE == 0 {
		return
	}
	// close any option section tags
	for i := options.sectionLevel; i > 0; i-- {
		out.WriteString("</section>\n")
		options.sectionLevel--
	}
	switch options.docLevel {
	case DOC_FRONT_MATTER:
		out.WriteString("</front>\n")
		out.WriteString("<back>\n")
	case DOC_MAIN_MATTER:
		out.WriteString("</middle>\n")
		out.WriteString("<back>\n")
	case DOC_BACK_MATTER:
		// nothing to do
	}
	options.docLevel = DOC_BACK_MATTER
	// count the references
	refi, refn := 0, 0
	for _, c := range citations {
		if c.typ == 'i' {
			refi++
		}
		if c.typ == 'n' {
			refn++
		}
	}
	// output <xi:include href="<references file>.xml"/>, we use file it its not empty, otherwise
	// we construct one for RFCNNNN and I-D.something something.
	if refi+refn > 0 {
		if refi > 0 {
			out.WriteString("<references title=\"Informative References\">\n")
			for _, c := range citations {
				if c.typ == 'i' {
					f := string(c.filename)
					if f == "" {
						f = referenceFile(c)
					}
					out.WriteString("\t<?rfc include=\"" + f + "\"?>\n")
				}
			}
			out.WriteString("</references>\n")
		}
		if refn > 0 {
			out.WriteString("<references title=\"Normative References\">\n")
			for _, c := range citations {
				if c.typ == 'n' {
					f := string(c.filename)
					if f == "" {
						f = referenceFile(c)
					}
					out.WriteString("\t<?rfc include=\"" + f + "\"?>\n")
				}
			}
			out.WriteString("</references>\n")
		}
	}
}

func (options *Xml2) AutoLink(out *bytes.Buffer, link []byte, kind int) {
	out.WriteString("<eref target=\"")
	if kind == LINK_TYPE_EMAIL {
		out.WriteString("mailto:")
	}
	out.Write(link)
	out.WriteString("\"/>")
}

func (options *Xml2) CodeSpan(out *bytes.Buffer, text []byte) {
	out.WriteString("<spanx style=\"verb\">")
	convertEntity(out, text)
	out.WriteString("</spanx>")
}

func (options *Xml2) DoubleEmphasis(out *bytes.Buffer, text []byte) {
	out.WriteString("<spanx style=\"strong\">")
	out.Write(text)
	out.WriteString("</spanx>")
}

func (options *Xml2) Emphasis(out *bytes.Buffer, text []byte) {
	out.WriteString("<spanx style=\"emph\">")
	out.Write(text)
	out.WriteString("</spanx>")
}

func (options *Xml2) Image(out *bytes.Buffer, link []byte, title []byte, alt []byte) {
	renderIAL(options.GetAndResetIAL()) // TODO(miek): useful?
	if bytes.HasPrefix(link, []byte("http://")) || bytes.HasPrefix(link, []byte("https://")) {
		// treat it like a link
		out.WriteString("\\href{")
		out.Write(link)
		out.WriteString("}{")
		out.Write(alt)
		out.WriteString("}")
	} else {
		out.WriteString("\\includegraphics{")
		out.Write(link)
		out.WriteString("}")
	}
}

func (options *Xml2) LineBreak(out *bytes.Buffer) {
	out.WriteString("\n<vspace/>\n")
}

func (options *Xml2) Link(out *bytes.Buffer, link []byte, title []byte, content []byte) {
	if link[0] == '#' {
		link = link[1:]
	}
	out.WriteString("<xref target=\"")
	out.Write(link)
	out.WriteString("\"/>")
	//	out.Write(content)
}

func (options *Xml2) RawHtmlTag(out *bytes.Buffer, tag []byte) {
}

func (options *Xml2) TripleEmphasis(out *bytes.Buffer, text []byte) {
	out.WriteString("<spanx style=\"strong\"><spanx style=\"emph\">")
	out.Write(text)
	out.WriteString("</spanx></spanx>")
}

func (options *Xml2) StrikeThrough(out *bytes.Buffer, text []byte) {
	out.Write(text)
}

func (options *Xml2) FootnoteRef(out *bytes.Buffer, ref []byte, id int) {
	// not used
}

func (options *Xml2) Entity(out *bytes.Buffer, entity []byte) {
	out.Write(entity)
}

func (options *Xml2) NormalText(out *bytes.Buffer, text []byte) {
	out.Write(text)
}

// header and footer
func (options *Xml2) DocumentHeader(out *bytes.Buffer, first bool) {
	if !first || options.flags&XML_STANDALONE == 0 {
		return
	}
	out.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	out.WriteString("<!DOCTYPE rfc SYSTEM 'rfc2629.dtd' [ ]>\n")
}

func (options *Xml2) DocumentFooter(out *bytes.Buffer, first bool) {
	if !first || options.flags&XML_STANDALONE == 0 {
		return
	}
	// close any option section tags
	for i := options.sectionLevel; i > 0; i-- {
		out.WriteString("</section>\n")
		options.sectionLevel--
	}
	switch options.docLevel {
	case DOC_FRONT_MATTER:
		out.WriteString("\n</front>\n")
	case DOC_MAIN_MATTER:
		out.WriteString("\n</middle>\n")
	case DOC_BACK_MATTER:
		out.WriteString("\n</back>\n")
	}
	out.WriteString("</rfc>\n")
}

func (options *Xml2) DocumentMatter(out *bytes.Buffer, matter int) {
	// we default to frontmatter already openened in the documentHeader
	for i := options.sectionLevel; i > 0; i-- {
		out.WriteString("</section>\n")
		options.sectionLevel--
	}
	switch matter {
	case DOC_FRONT_MATTER:
		// already open
	case DOC_MAIN_MATTER:
		out.WriteString("</front>\n")
		out.WriteString("\n<middle>\n")
	case DOC_BACK_MATTER:
		out.WriteString("\n</middle>\n")
		out.WriteString("<back>\n")
	}
	options.docLevel = matter
}