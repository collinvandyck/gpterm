package markdown

import (
	"bytes"
	"fmt"
	"io"

	"github.com/muesli/reflow/wordwrap"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
)

var _ renderer.Renderer = &Renderer{}

// An experimental renderer that will build its own ast
type Renderer struct {
	width  int
	w      io.Writer
	source []byte
}

func NewRenderer(width int) *Renderer {
	width = width - 2
	return &Renderer{
		width: width,
	}
}

// AddOptions implements renderer.Renderer
func (r *Renderer) AddOptions(...renderer.Option) {
}

// Render implements renderer.Renderer
func (r *Renderer) Render(w io.Writer, source []byte, node ast.Node) error {
	ww := wordwrap.NewWriter(r.width)
	r.w = ww
	r.source = source
	expr := r.parse(node)
	visitor := NewPrinter(ww, r.width)
	expr.Visit(visitor)
	ww.Close()
	io.Copy(w, bytes.NewReader(ww.Bytes()))
	return nil
}

func (r *Renderer) newPrinter() *Printer {
	return NewPrinter(r.w, r.width)
}

func (r *Renderer) parse(node ast.Node) Expr {
	switch node := node.(type) {
	case *ast.Document:
		return r.document(node)
	case *ast.TextBlock:
		return r.textBlock(node)
	case *ast.CodeBlock:
		return r.codeBlock(node)
	case *ast.FencedCodeBlock:
		return r.fencedCodeBlock(node)
	case *ast.HTMLBlock:
		return r.htmlBlock(node)
	case *ast.Paragraph:
		return r.paragraph(node)
	case *ast.Text:
		return r.text(node)
	case *ast.List:
		return r.list(node)
	case *ast.ListItem:
		return r.listItem(node)
	case *ast.CodeSpan:
		return r.codeSpan(node)
	case *ast.String:
		return r.string(node)
	case *ast.RawHTML:
		return r.rawHTML(node)
	case *ast.Link:
		return r.link(node)
	case *ast.AutoLink:
		return r.autoLink(node)
	case *ast.Heading:
		return r.heading(node)
	case *ast.Emphasis:
		return r.emphasis(node)
	}
	panic(fmt.Sprintf("unhandled node type: %T", node))
}

func (r *Renderer) document(node *ast.Document) Expr {
	doc := Document{node: node}
	doc.children = r.parseChildren(node)
	return doc
}

func (r *Renderer) textBlock(node *ast.TextBlock) Expr {
	res := TextBlock{node: node}
	res.blanksBefore = r.blanksBefore(node)
	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		bs := line.Value(r.source)
		res.lines = append(res.lines, string(bs))
	}
	return res
}

func (r *Renderer) codeBlock(node *ast.CodeBlock) Expr {
	res := TextBlock{node: node, code: true}
	res.blanksBefore = r.blanksBefore(node)
	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		bs := line.Value(r.source)
		res.lines = append(res.lines, string(bs))
	}
	return res
}

func (r *Renderer) fencedCodeBlock(node *ast.FencedCodeBlock) Expr {
	res := TextBlock{node: node, code: true}
	res.language = string(node.Language(r.source))
	res.blanksBefore = r.blanksBefore(node)
	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		bs := line.Value(r.source)
		res.lines = append(res.lines, string(bs))
	}
	return res
}

func (r *Renderer) htmlBlock(node *ast.HTMLBlock) Expr {
	res := HTMLBlock{node: node}
	res.blanksBefore = r.blanksBefore(node)
	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		bs := line.Value(r.source)
		res.lines = append(res.lines, string(bs))
	}
	return res
}

func (r *Renderer) paragraph(node *ast.Paragraph) Expr {
	res := Paragraph{node: node}
	res.blanksBefore = r.blanksBefore(node)
	res.children = r.parseChildren(node)
	return res
}

func (r *Renderer) text(node *ast.Text) Expr {
	res := Text{node: node}
	res.text = string(node.Segment.Value(r.source))
	res.softBreak = node.SoftLineBreak()
	res.hardBreak = node.HardLineBreak()
	return res
}

func (r *Renderer) string(node *ast.String) Expr {
	return String{node: node, text: string(node.Value)}
}

func (r *Renderer) list(node *ast.List) Expr {
	res := List{node: node, marker: string([]byte{node.Marker}), tight: node.IsTight}
	res.children = r.parseChildren(node)
	return res
}

func (r *Renderer) listItem(node *ast.ListItem) Expr {
	res := ListItem{node: node}
	res.children = r.parseChildren(node)
	return res
}

func (r *Renderer) codeSpan(node *ast.CodeSpan) Expr {
	res := CodeSpan{node: node}
	res.children = r.parseChildren(node)
	return res
}

func (r *Renderer) rawHTML(node *ast.RawHTML) Expr {
	res := RawHTML{node: node}
	for i := 0; i < node.Segments.Len(); i++ {
		segment := node.Segments.At(i)
		bs := segment.Value(r.source)
		res.lines = append(res.lines, string(bs))
	}
	return res
}

func (r *Renderer) link(node *ast.Link) Expr {
	res := Link{node: node}
	res.destination = string(node.Destination)
	res.title = string(node.Title)
	return res
}

func (r *Renderer) autoLink(node *ast.AutoLink) Expr {
	res := AutoLink{node: node}
	res.autoLinkType = node.AutoLinkType
	res.protocol = string(node.Protocol)
	res.value = string(node.Text(r.source))
	return res
}

func (r *Renderer) heading(node *ast.Heading) Expr {
	res := Heading{node: node}
	res.blanksBefore = r.blanksBefore(node)
	res.text = string(node.Text(r.source))
	res.level = node.Level
	return res
}

func (r *Renderer) emphasis(node *ast.Emphasis) Expr {
	res := Emphasis{node: node}
	res.level = node.Level
	res.children = r.parseChildren(node)
	return res
}

func (r *Renderer) parseChildren(node ast.Node) []Expr {
	var res []Expr
	child := node.FirstChild()
	for child != nil {
		res = append(res, r.parse(child))
		child = child.NextSibling()
	}
	return res
}

func (r *Renderer) blanksBefore(node ast.Node) bool {
	if node.Type() == ast.TypeBlock {
		return node.HasBlankPreviousLines()
	}
	return false
}
