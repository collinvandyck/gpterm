package markdown

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

var (
	newline   = "\n"
	backtick  = "`"
	backticks = "```"
)

func RenderBytes(bs []byte, width int) ([]byte, error) {
	renderer := NewMarkdown(width)
	return renderer.RenderBytes(bs)
}

func RenderString(s string, width int) ([]byte, error) {
	bs, err := RenderBytes([]byte(s), width)
	return bs, err
}

type Markdown struct {
	width int
}

func NewMarkdown(width int) *Markdown {
	return &Markdown{
		width: width,
	}
}

func (r *Markdown) RenderBytes(bs []byte) ([]byte, error) {
	renderer := NewRenderer(r.width)
	md := goldmark.New(goldmark.WithRenderer(renderer))
	buf := &bytes.Buffer{}
	err := md.Convert(bs, buf)
	bs = buf.Bytes()
	bs = bytes.TrimSpace(bs)
	bs = append(bs, newline...)
	return bs, err
}

var _ renderer.NodeRenderer = &nodeRender{}

type nodeRender struct {
	writer      *writer
	debug       bool
	listDepth   int
	depth       int
	inCodeSpan  bool
	inCodeBlock bool
}

// RegisterFuncs implements renderer.NodeRenderer
func (n *nodeRender) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindDocument, n.renderDocument)

	reg.Register(ast.KindTextBlock, n.renderTextBlock)
	reg.Register(ast.KindCodeBlock, n.renderCodeBlock)
	reg.Register(ast.KindFencedCodeBlock, n.renderFencedCodeBlock)
	reg.Register(ast.KindHTMLBlock, n.renderHTMLBlock)
	reg.Register(ast.KindParagraph, n.renderParagraph)

	reg.Register(ast.KindText, n.renderText)
	reg.Register(ast.KindList, n.renderList)
	reg.Register(ast.KindListItem, n.renderListItem)
	reg.Register(ast.KindCodeSpan, n.renderCodeSpan)
	reg.Register(ast.KindString, n.renderString)
	reg.Register(ast.KindRawHTML, n.renderRawHTML)

	reg.Register(ast.KindThematicBreak, n.renderDebug)
	reg.Register(ast.KindLink, n.renderDebug)
	reg.Register(ast.KindAutoLink, n.renderDebug)
	reg.Register(ast.KindAutoLink, n.renderDebug)
}

func (n *nodeRender) renderDebug(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	fmt.Fprintf(os.Stderr, "renderDebug: %s\n", node.Kind().String())
	return ast.WalkContinue, nil
}

func (n *nodeRender) renderCodeBlock(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n.visited(writer, source, node, entering)
	n.inCodeBlock = entering
	block := node.(*ast.CodeBlock)
	if entering {
		n.startBlock(writer, node)
		io.WriteString(writer, backticks)
		io.WriteString(writer, newline)
		lines := block.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			writer.Write(line.Value(source))
		}
	} else {
		io.WriteString(writer, backticks)
	}
	return ast.WalkContinue, nil
}

func (n *nodeRender) renderFencedCodeBlock(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n.visited(writer, source, node, entering)
	n.inCodeBlock = entering
	block := node.(*ast.FencedCodeBlock)
	language := string(block.Language(source))
	style := lipgloss.NewStyle()
	style = style.Foreground(lipgloss.Color("#00ff00"))
	style = style.Background(lipgloss.Color("#333333"))
	if entering {
		n.startBlock(writer, node)
		io.WriteString(writer, style.Render(backticks))
		io.WriteString(writer, style.Render(language))
		io.WriteString(writer, style.Render(newline))
		lines := block.Lines()
		for i := 0; i < lines.Len(); i++ {
			at := lines.At(i)
			line := string(at.Value(source))
			line = strings.TrimRight(line, "\n")
			io.WriteString(writer, style.Render(line))
			io.WriteString(writer, newline)
		}
	} else {
		io.WriteString(writer, style.Render(backticks))
	}
	return ast.WalkContinue, nil
}

func (n *nodeRender) renderHTMLBlock(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n.visited(writer, source, node, entering)
	block := node.(*ast.HTMLBlock)
	if entering {
		n.startBlock(writer, node)
		for i := 0; i < block.Lines().Len(); i++ {
			line := block.Lines().At(i)
			writer.Write(line.Value(source))
		}
	}
	return ast.WalkContinue, nil
}

func (n *nodeRender) renderTextBlock(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n.visited(writer, source, node, entering)
	block := node.(*ast.TextBlock)
	if entering {
		n.startBlock(writer, node)
		for i := 0; i < block.Lines().Len(); i++ {
			line := block.Lines().At(i)
			writer.Write(line.Value(source))
		}
	}
	return ast.WalkContinue, nil
}

func (n *nodeRender) renderRawHTML(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	rawHTML := node.(*ast.RawHTML)
	if entering {
		for i := 0; i < rawHTML.Segments.Len(); i++ {
			segment := rawHTML.Segments.At(i)
			writer.Write(segment.Value(source))
		}
	}
	return ast.WalkContinue, nil
}

func (n *nodeRender) renderDocument(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering && n.debug {
		node.Dump(source, 0)
	}
	n.visited(writer, source, node, entering)
	if entering {
	} else {
		io.WriteString(writer, newline)
	}
	return ast.WalkContinue, nil
}

func (n *nodeRender) renderList(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n.visited(writer, source, node, entering)
	if entering {
		n.listDepth++
	} else {
		n.listDepth--
	}
	return ast.WalkContinue, nil
}

func (n *nodeRender) renderListItem(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n.visited(writer, source, node, entering)
	item := node.(*ast.ListItem)
	if entering {
		if item.HasBlankPreviousLines() {
			io.WriteString(writer, newline)
		}
		io.WriteString(writer, "* ")
	} else {
		io.WriteString(writer, newline)
	}
	return ast.WalkContinue, nil
}

func (n *nodeRender) renderText(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	text := node.(*ast.Text)
	n.visited(writer, source, node, entering)
	if entering {
		text := node.(*ast.Text)
		style := lipgloss.NewStyle()
		if n.inCodeSpan {
			style = style.Bold(true).Foreground(lipgloss.Color("#f81ce5"))
		}
		rendered := style.Render(string(text.Segment.Value(source)))
		io.WriteString(writer, rendered)
	} else {
		if text.HardLineBreak() {
			io.WriteString(writer, newline)
			io.WriteString(writer, newline)
		}
		if text.SoftLineBreak() {
			io.WriteString(writer, newline)
		}
	}
	return ast.WalkContinue, nil
}

func (n *nodeRender) renderParagraph(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n.visited(writer, source, node, entering)
	if entering {
		n.startBlock(writer, node)
	}
	return ast.WalkContinue, nil
}

func (n *nodeRender) renderCodeSpan(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n.visited(writer, source, node, entering)
	n.inCodeSpan = entering
	io.WriteString(writer, backtick)
	return ast.WalkContinue, nil
}

func (n *nodeRender) renderString(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n.visited(writer, source, node, entering)
	str := node.(*ast.String)
	if entering {
		writer.Write(str.Value)
	}
	return ast.WalkContinue, nil
}

func (n nodeRender) startBlock(writer util.BufWriter, node ast.Node) {
	if !n.inList() {
		io.WriteString(writer, newline)
		if node.HasBlankPreviousLines() {
			io.WriteString(writer, newline)
		}
	}
}

func (n *nodeRender) inList() bool {
	return n.listDepth > 0
}

func (n *nodeRender) visited(writer util.BufWriter, source []byte, node ast.Node, entering bool) {
	n.debugNode(node, entering, source)
}

func (n *nodeRender) debugNode(node ast.Node, entering bool, source []byte) {
	if !n.debug {
		return
	}
	depth := n.depth
	if !entering {
		depth = depth - 1
	}
	indent := strings.Repeat(" ", depth)
	typ := fmt.Sprintf("%T", node)
	direction := "->"
	if entering {
		block := node.Type() == ast.TypeBlock
		if block {
			switch node := node.(type) {
			case *ast.Paragraph:
				buf := new(bytes.Buffer)
				for i := 0; i < node.Lines().Len(); i++ {
					seg := node.Lines().At(i)
					bs := seg.Value(source)
					buf.Write(bs)
					buf.Write([]byte(newline))
				}
				fmt.Printf("%s%s BLOCK %s %q %v\n", indent, direction, typ, buf.String(), node.HasBlankPreviousLines())
			default:
				fmt.Printf("%s%s BLOCK %s %v\n", indent, direction, typ, node.HasBlankPreviousLines())
			}
		} else {
			switch node := node.(type) {
			case *ast.Text:
				val := node.Segment.Value(source)
				fmt.Printf("%s%s %s val=%q hard=%v soft=%v\n", indent, direction, typ, string(val), node.HardLineBreak(), node.SoftLineBreak())
			default:
				fmt.Printf("%s%s %s\n", indent, direction, typ)
			}
		}
	}
	if entering {
		n.depth++
	} else {
		n.depth--
	}
}
