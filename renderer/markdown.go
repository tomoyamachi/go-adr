package renderer

import (
	"bytes"
	"fmt"
	"io"

	"github.com/tomoyamachi/go-adr/models"

	. "gopkg.in/russross/blackfriday.v2"
)

// MarkdownRenderer is a type that implements the Renderer interface for Markdown output.
type MarkdownRenderer struct {
	w                  bytes.Buffer
	newHistory         *models.History
	lastOutputLen      int
	tableHeadCount     int
	firstTableCell     bool
	currentHeaderTitle string
}

var (
	quoteTag          = []byte(">")
	codeTag           = []byte("```")
	imageTag          = []byte("!")
	strongTag         = []byte("*")
	strikethroughTag  = []byte("-")
	emTag             = []byte("_")
	linkTitleTag      = []byte("[")
	linkTitleCloseTag = []byte("]")
	linkTag           = []byte("(")
	linkCloseTag      = []byte(")")
	liTag             = []byte("*")
	olTag             = []byte("1.")
	nestLiTag         = []byte("  ")
	hrTag             = []byte("----")
	tableColumnDivide = []byte("|")
	tableRowDivide    = []byte("---")
	h1Tag             = []byte("#")
	h2Tag             = []byte("##")
	h3Tag             = []byte("###")
	h4Tag             = []byte("####")
	h5Tag             = []byte("#####")
	h6Tag             = []byte("######")
)

var (
	nlBytes    = []byte{'\n'}
	spaceBytes = []byte{' '}
)

var itemLevel = 0

func (r *MarkdownRenderer) cr(w io.Writer) {
	// Linkのあとに0byteのTextオブジェクトが入ることがあるので、必ず開業が入るように修正
	if r.lastOutputLen > -1 {
		r.out(w, nlBytes)
	}
}

func (r *MarkdownRenderer) out(w io.Writer, text []byte) {
	w.Write(text)
	r.lastOutputLen = len(text)
}

func headingTagFromLevel(level int) []byte {
	switch level {
	case 1:
		return h1Tag
	case 2:
		return h2Tag
	case 3:
		return h3Tag
	case 4:
		return h4Tag
	case 5:
		return h5Tag
	default:
		return h6Tag
	}
}

// Render prints out the whole document from the ast.
func (r *MarkdownRenderer) ChangeStatus(ast *Node) []byte {
	ast.Walk(func(node *Node, entering bool) WalkStatus {
		return r.RenderNode(&r.w, node, entering)
	})

	return r.w.Bytes()
}

func (r *MarkdownRenderer) shouldReplaceNewStatus(node *Node) bool {
	if node.Parent.Type != Heading && r.newHistory != nil && r.currentHeaderTitle == models.StatusHeader {
		return true
	}
	return false
}

func (r *MarkdownRenderer) addNewHistory(w io.Writer) {
	// History内のテーブルなら、rowにデータを追加
	if r.newHistory != nil && r.currentHeaderTitle == models.HistoryHeader {
		newRow := fmt.Sprintf(
			"%s | %s | %s",
			r.newHistory.Date,
			r.newHistory.Status,
			r.newHistory.Memo,
		)
		r.out(w, []byte(newRow))
		r.cr(w)
	}
}

// RenderNode is output a each single node
func (r *MarkdownRenderer) RenderNode(w io.Writer, node *Node, entering bool) WalkStatus {
	switch node.Type {
	case Text:
		if r.shouldReplaceNewStatus(node) {
			r.out(w, []byte(r.newHistory.Status))
		} else {
			r.out(w, node.Literal)
		}
		// そのときの直近のヘッダテキストを保存
		if node.Parent.Type == Heading {
			r.currentHeaderTitle = string(node.Literal)
		}
	case Softbreak:
		break
	case Hardbreak:
		break
	case BlockQuote:
		if entering {
			r.out(w, quoteTag)
			w.Write(spaceBytes)
		}
	case CodeBlock:
		r.out(w, []byte("```"))
		r.cr(w)
		w.Write(node.Literal)
		r.cr(w)
		r.out(w, []byte("```"))
		r.cr(w)
		r.cr(w)
	case Code:
		r.out(w, []byte("`"))
		r.out(w, node.Literal)
		r.out(w, []byte("`"))
	case Emph:
		r.out(w, emTag)
	case Heading:
		headingTag := headingTagFromLevel(node.Level)
		if entering {
			r.out(w, headingTag)
			w.Write(spaceBytes)
		} else {
			r.cr(w)
		}
	case Image:
		if entering {
			dest := node.LinkData.Destination
			r.out(w, imageTag)
			r.out(w, dest)
		} else {
			r.out(w, imageTag)
		}
	case Item:
		if entering {
			itemTag := liTag
			if node.ListFlags&ListTypeOrdered != 0 {
				itemTag = olTag
			}
			for i := 1; i <= itemLevel; i++ {
				if i == itemLevel {
					r.out(w, itemTag)
				} else {
					r.out(w, nestLiTag)
				}
			}

			w.Write(spaceBytes)
		}
	case Link:
		if entering {
			r.out(w, linkTitleTag)
		} else {
			r.out(w, linkTitleCloseTag)
			r.out(w, linkTag)
			r.out(w, node.LinkData.Destination)
			r.out(w, linkCloseTag)
		}
	case HorizontalRule:
		r.cr(w)
		r.out(w, hrTag)
		r.cr(w)
	case List:
		if entering {
			itemLevel++
		} else {
			itemLevel--
			if itemLevel == 0 {
				r.cr(w)
			}
		}
	case Document:
		break
	case Paragraph:
		if !entering {
			if node.Parent.Type != Item {
				r.cr(w)
			}
			r.cr(w)
		}
	case Strong:
		r.out(w, strongTag)
	case Del:
		r.out(w, strikethroughTag)
	case Table:
		r.firstTableCell = true
		if !entering {
			// ヘッダ数をリセット
			r.tableHeadCount = 0
			// 必要ならHistoryにRow追加
			r.addNewHistory(w)
			r.cr(w)
		}
	case TableCell:
		// ヘッダのCellが増えるごとにヘッダ数を増やす
		if node.IsHeader && entering {
			r.tableHeadCount++
		}
		if entering {
			if r.firstTableCell {
				r.firstTableCell = false
			} else {
				r.out(w, tableColumnDivide)
			}
		}
	case TableHead:
		if entering {
			r.tableHeadCount = 0
		} else {
			for i := 1; i <= r.tableHeadCount; i++ {
				r.out(w, tableRowDivide)
				if i != r.tableHeadCount {
					r.out(w, tableColumnDivide)
				}
			}
			r.cr(w)
		}
		break
	case TableBody:
		break
	case TableRow:
		if !entering {
			r.firstTableCell = true
			r.cr(w)
		}
	case HTMLSpan:
		r.out(w, node.Literal)
	default:
		panic("Unknown node type " + node.Type.String())
	}
	return GoToNext
}

// Render prints out the whole document from the ast.
func (r *MarkdownRenderer) Render(ast *Node) []byte {
	ast.Walk(func(node *Node, entering bool) WalkStatus {
		return r.RenderNode(&r.w, node, entering)
	})

	return r.w.Bytes()
}

// RenderHeader writes document header (unused).
func (r *MarkdownRenderer) RenderHeader(w io.Writer, ast *Node) {
}

// RenderFooter writes document footer (unused).
func (r *MarkdownRenderer) RenderFooter(w io.Writer, ast *Node) {
}

// MarkdownRun return markdown bytes
func MarkdownRun(input []byte, opts ...Option) []byte {
	r := &MarkdownRenderer{}
	optList := []Option{WithRenderer(r), WithExtensions(CommonExtensions)}
	optList = append(optList, opts...)
	parser := New(optList...)
	ast := parser.Parse([]byte(input))
	return r.Render(ast)
}

func StatusChange(input []byte, history *models.History, opts ...Option) []byte {
	r := &MarkdownRenderer{newHistory: history}
	optList := []Option{WithRenderer(r), WithExtensions(CommonExtensions)}
	optList = append(optList, opts...)
	parser := New(optList...)
	ast := parser.Parse([]byte(input))

	return r.Render(ast)
}
