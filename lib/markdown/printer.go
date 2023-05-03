package markdown

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/yuin/goldmark/ast"
)

var _ Visitor = &Printer{}

type Printer struct {
	io.Writer
	width     int
	debug     bool
	exprDepth int
	lists     stack[ListContext]
	lineStart bool
	styles    stack[lipgloss.Style]
}

type ListContext struct {
	list  List
	items stack[ListItem]
}

func NewPrinter(w io.Writer, width int) *Printer {
	return &Printer{
		Writer: w,
		width:  width,
		debug:  false,
	}
}

func (p *Printer) debugExpr(expr Expr, node ast.Node) {
	if !p.debug {
		return
	}
	indent := strings.Repeat("  ", p.exprDepth)
	fmt.Printf("%s%s: \n", indent, expr)
}

// VisitDocument implements Visitor
func (p *Printer) VisitDocument(expr Document) {
	p.debugExpr(expr, expr.node)
	p.visitChildren(expr.children)
}

// VisitCodeSpan implements Visitor
func (p *Printer) VisitCodeSpan(expr CodeSpan) {
	p.debugExpr(expr, expr.node)
	p.styles.push(lipgloss.NewStyle().Foreground(lipgloss.Color("#f5a2ff")))
	defer p.styles.pop()
	p.visitChildren(expr.children)
}

// VisitHTMLBlock implements Visitor
func (p *Printer) VisitHTMLBlock(expr HTMLBlock) {
	p.debugExpr(expr, expr.node)
	p.startBlock(expr.blanksBefore)
}

// VisitList implements Visitor
func (p *Printer) VisitList(expr List) {
	p.debugExpr(expr, expr.node)
	p.lists.push(ListContext{list: expr})
	p.visitChildren(expr.children)
	p.lists.pop()
}

// VisitListItem implements Visitor
func (p *Printer) VisitListItem(expr ListItem) {
	p.debugExpr(expr, expr.node)
	p.lists.peek().items.push(expr)
	p.visitChildren(expr.children)
	p.lists.peek().items.pop()
}

// VisitParagraph implements Visitor
func (p *Printer) VisitParagraph(expr Paragraph) {
	p.debugExpr(expr, expr.node)
	p.startBlock(expr.blanksBefore)
	p.visitChildren(expr.children)
}

// VisitRawHTML implements Visitor
func (p *Printer) VisitRawHTML(expr RawHTML) {
	p.debugExpr(expr, expr.node)
}

// VisitString implements Visitor
func (p *Printer) VisitString(expr String) {
	p.debugExpr(expr, expr.node)
}

// VisitText implements Visitor
func (p *Printer) VisitText(expr Text) {
	p.debugExpr(expr, expr.node)
	p.string(expr.text)
	if expr.hardBreak {
		p.newline()
		p.newline()
	}
	if expr.softBreak {
		p.newline()
	}
}

// VisitTextBlock implements Visitor
func (p *Printer) VisitTextBlock(expr TextBlock) {
	p.debugExpr(expr, expr.node)
	p.startBlock(expr.blanksBefore)
	if expr.code {
		if expr.language != "" {
			lexer := lexers.Get(expr.language)
			if lexer == nil {
				lexer = lexers.Fallback
			}
			style := styles.Get("monokai")
			if style == nil {
				style = styles.Fallback
			}
			buf := new(bytes.Buffer)
			for _, line := range expr.lines {
				buf.WriteString(line)
			}
			iter, err := lexer.Tokenise(nil, buf.String())
			if err != nil {
				panic(err)
			}
			formatter := formatters.Get("terminal16m")
			if formatter == nil {
				formatter = formatters.Fallback
			}
			buf.Reset()
			err = formatter.Format(buf, style, iter)
			if err != nil {
				panic(err)
			}
			str := strings.TrimSpace(buf.String())
			p.string(str)
			return
		}
		p.styles.push(lipgloss.NewStyle().Foreground(lipgloss.Color("#ddff00")))
		defer p.styles.pop()
		for i, line := range expr.lines {
			p.string(strings.TrimRight(line, "\n"))
			if i < len(expr.lines)-1 {
				p.newline()
			}
		}
		return
	}
	for _, line := range expr.lines {
		p.string(line)
	}
}

// VisitAutoLink implements Visitor
func (p *Printer) VisitAutoLink(expr AutoLink) {
	p.debugExpr(expr, expr.node)
}

// VisitLink implements Visitor
func (p *Printer) VisitLink(expr Link) {
	p.debugExpr(expr, expr.node)
	if expr.title != "" {
		p.string(fmt.Sprintf("[%s](%s)", expr.destination, expr.destination))
		return
	}
	p.string(expr.destination)
}

func (p *Printer) VisitHeading(expr Heading) {
	p.debugExpr(expr, expr.node)
	p.startBlock(expr.blanksBefore)
	p.string(strings.Repeat("#", expr.level))
	p.string(" " + expr.text)
	p.visitChildren(expr.children)
}

func (p *Printer) VisitEmphasis(expr Emphasis) {
	p.debugExpr(expr, expr.node)
	p.styles.push(lipgloss.NewStyle().Italic(true))
	defer p.styles.pop()
	p.visitChildren(expr.children)
}

func (p *Printer) visitChildren(children []Expr) {
	p.exprDepth++
	for _, child := range children {
		child.Visit(p)
	}
	p.exprDepth--
}

func (p *Printer) startBlock(blanksBefore bool) {
	p.newline()
	if blanksBefore {
		p.newline()
	}
}

func (p *Printer) string(val string) {
	if p.lineStart && p.lists.peek() != nil {
		indent := strings.Repeat("  ", len(p.lists)-1)
		marker := p.lists.peek().list.marker
		p.write(indent)
		p.write(marker)
		p.write(" ")
	}
	style := p.styles.peek()
	if style != nil {
		val = style.Render(val)
	}
	p.write(val)
	p.lineStart = false
}

func (p *Printer) newline() {
	p.write("\n")
	p.lineStart = true
}

func (p *Printer) write(val string) {
	io.WriteString(p, val)
}
