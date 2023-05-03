package markdown

import (
	"fmt"

	"github.com/yuin/goldmark/ast"
)

type Visitor interface {
	VisitDocument(expr Document)
	VisitTextBlock(expr TextBlock)
	VisitHTMLBlock(expr HTMLBlock)
	VisitParagraph(expr Paragraph)
	VisitText(expr Text)
	VisitString(expr String)
	VisitList(expr List)
	VisitListItem(expr ListItem)
	VisitCodeSpan(expr CodeSpan)
	VisitRawHTML(expr RawHTML)
	VisitLink(expr Link)
	VisitAutoLink(expr AutoLink)
	VisitHeading(expr Heading)
	VisitEmphasis(expr Emphasis)
}

type Expr interface {
	Visit(Visitor)
	String() string
}

type Document struct {
	node     ast.Node
	children []Expr
}

func (d Document) Visit(v Visitor) {
	v.VisitDocument(d)
}

func (d Document) String() string {
	return "Document"
}

type TextBlock struct {
	node         ast.Node
	blanksBefore bool
	code         bool
	language     string
	lines        []string
}

func (t TextBlock) Visit(v Visitor) {
	v.VisitTextBlock(t)
}

func (t TextBlock) String() string {
	return fmt.Sprintf("TextBlock{blanksBefore: %v, code: %v, language: %q, lines: %d}", t.blanksBefore, t.code, t.language, len(t.lines))
}

type HTMLBlock struct {
	node         ast.Node
	blanksBefore bool
	lines        []string
}

func (h HTMLBlock) Visit(v Visitor) {
	v.VisitHTMLBlock(h)
}

func (h HTMLBlock) String() string {
	return fmt.Sprintf("HTMLBlock{blanksBefore: %v, lines: %d}", h.blanksBefore, len(h.lines))
}

type Paragraph struct {
	node         ast.Node
	blanksBefore bool
	children     []Expr
}

func (p Paragraph) Visit(v Visitor) {
	v.VisitParagraph(p)
}

func (p Paragraph) String() string {
	return fmt.Sprintf("Paragraph{blanksBefore: %v, children: %d}", p.blanksBefore, len(p.children))
}

type Text struct {
	node      ast.Node
	hardBreak bool
	softBreak bool
	text      string
}

func (t Text) Visit(v Visitor) {
	v.VisitText(t)
}

func (t Text) String() string {
	return fmt.Sprintf("Text{hardBreak: %v, softBreak: %v, text: %q}", t.hardBreak, t.softBreak, t.text)
}

type String struct {
	node ast.Node
	text string
}

func (s String) Visit(v Visitor) {
	v.VisitString(s)
}

func (s String) String() string {
	return fmt.Sprintf("String{text: %q}", s.text)
}

type List struct {
	node         ast.Node
	blanksBefore bool
	marker       string
	tight        bool
	children     []Expr
}

func (l List) Visit(v Visitor) {
	v.VisitList(l)
}

func (l List) String() string {
	return fmt.Sprintf("List{blanksBefore: %v, marker:%s, tight:%v children: %d}", l.blanksBefore, l.marker, l.tight, len(l.children))
}

type ListItem struct {
	node         ast.Node
	blanksBefore bool
	children     []Expr
}

func (l ListItem) Visit(v Visitor) {
	v.VisitListItem(l)
}

func (l ListItem) String() string {
	return fmt.Sprintf("ListItem{blanksBefore: %v, children: %d}", l.blanksBefore, len(l.children))
}

type CodeSpan struct {
	node     ast.Node
	children []Expr
}

func (c CodeSpan) Visit(v Visitor) {
	v.VisitCodeSpan(c)
}

func (c CodeSpan) String() string {
	return fmt.Sprintf("CodeSpan{children: %d}", len(c.children))
}

type RawHTML struct {
	node  ast.Node
	lines []string
}

func (r RawHTML) Visit(v Visitor) {
	v.VisitRawHTML(r)
}

func (r RawHTML) String() string {
	return fmt.Sprintf("RawHTML{lines: %d}", len(r.lines))
}

type Link struct {
	node        ast.Node
	destination string
	title       string
}

func (l Link) Visit(v Visitor) {
	v.VisitLink(l)
}

func (l Link) String() string {
	return fmt.Sprintf("Link{destination: %v, title: %v}", l.destination, l.title)
}

type AutoLink struct {
	node         ast.Node
	autoLinkType ast.AutoLinkType
	protocol     string
	value        string
}

func (a AutoLink) Visit(v Visitor) {
	v.VisitAutoLink(a)
}

func (a AutoLink) String() string {
	return fmt.Sprintf("Autolink{type: %d, protocol: %s, value: %v}", a.autoLinkType, a.protocol, a.value)
}

type Heading struct {
	node         ast.Node
	text         string
	children     []Expr
	blanksBefore bool
	level        int
}

func (h Heading) Visit(v Visitor) {
	v.VisitHeading(h)
}

func (h Heading) String() string {
	return fmt.Sprintf("Heading{level: %d, blanksBefore:%v, children:%d}", h.level, h.blanksBefore, len(h.children))
}

type Emphasis struct {
	node     ast.Node
	children []Expr
	level    int
}

func (e Emphasis) Visit(v Visitor) {
	v.VisitEmphasis(e)
}

func (e Emphasis) String() string {
	return fmt.Sprintf("Emphasis{}")
}
