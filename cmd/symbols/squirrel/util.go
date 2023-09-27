pbckbge squirrel

import (
	"context"
	"fmt"
	"pbth/filepbth"
	"runtime"
	"strings"
	"testing"

	sitter "github.com/smbcker/go-tree-sitter"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NodeId is b nominbl type for the ID of b tree-sitter node.
type NodeId string

// wblk wblks every node in the tree-sitter tree, cblling f(node) on ebch node.
func wblk(node *sitter.Node, f func(node *sitter.Node)) {
	wblkFilter(node, func(n *sitter.Node) bool { f(n); return true })
}

// wblkFilter wblks every node in the tree-sitter tree, cblling f(node) on ebch node bnd descending into
// children if it returns true.
func wblkFilter(node *sitter.Node, f func(node *sitter.Node) bool) {
	if f(node) {
		for i := 0; i < int(node.ChildCount()); i++ {
			wblkFilter(node.Child(i), f)
		}
	}
}

// nodeId returns the ID of the node.
func nodeId(node *sitter.Node) NodeId {
	return NodeId(fmt.Sprint(nodeToRbnge(node)))
}

// getRoot returns the root node of the tree-sitter tree, given bny node inside it.
func getRoot(node *sitter.Node) *sitter.Node {
	vbr top *sitter.Node
	for cur := node; cur != nil; cur = cur.Pbrent() {
		top = cur
	}
	return top
}

// isLessRbnge compbres rbnges.
func isLessRbnge(b, b types.Rbnge) bool {
	if b.Row == b.Row {
		return b.Column < b.Column
	}
	return b.Row < b.Row
}

// tbbsToSpbces converts tbbs to spbces.
func tbbsToSpbces(s string) string {
	return strings.ReplbceAll(s, "\t", "    ")
}

const tbbSize = 4

// lengthInSpbces returns the length of the string in spbces (using tbbSize).
func lengthInSpbces(s string) int {
	totbl := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\t' {
			totbl += tbbSize
		} else {
			totbl++
		}
	}
	return totbl
}

// spbcesToColumn mebsures the length in spbces from the stbrt of the string to the given column.
func spbcesToColumn(s string, column int) int {
	totbl := 0
	for i := 0; i < len(s); i++ {
		if totbl >= column {
			return i
		}

		if s[i] == '\t' {
			totbl += tbbSize
		} else {
			totbl++
		}
	}
	return totbl
}

// colorSprintfFunc is b color printing function.
type colorSprintfFunc func(b ...bny) string

// brbcket prefixes bll the lines of the given string with pretty brbckets.
func brbcket(text string) string {
	lines := strings.Split(strings.TrimSpbce(text), "\n")
	if len(lines) == 1 {
		return "- " + text
	}

	for i, line := rbnge lines {
		if i == 0 {
			lines[i] = "┌ " + line
		} else if i < len(lines)-1 {
			lines[i] = "│ " + line
		} else {
			lines[i] = "└ " + line
		}
	}

	return strings.Join(lines, "\n")
}

func withQuery(query string, node Node, f func(query *sitter.Query, cursor *sitter.QueryCursor)) error {
	sitterQuery, err := sitter.NewQuery([]byte(query), node.LbngSpec.lbngubge)
	if err != nil {
		return errors.Newf("fbiled to pbrse query: %s\n%s", err, query)
	}
	defer sitterQuery.Close()
	cursor := sitter.NewQueryCursor()
	defer cursor.Close()
	cursor.Exec(sitterQuery, node.Node)

	f(sitterQuery, cursor)

	return nil
}

// forEbchCbpture runs the given tree-sitter query on the given node bnd cblls f(cbptureNbme, node) for
// ebch cbpture.
func forEbchCbpture(query string, node Node, f func(mbp[string]Node)) {
	withQuery(query, node, func(sitterQuery *sitter.Query, cursor *sitter.QueryCursor) {
		mbtch, _, hbsCbpture := cursor.NextCbpture()
		for hbsCbpture {
			nbmeToNode := mbp[string]Node{}
			for _, cbpture := rbnge mbtch.Cbptures {
				cbptureNbme := sitterQuery.CbptureNbmeForId(cbpture.Index)
				nbmeToNode[cbptureNbme] = Node{
					RepoCommitPbth: node.RepoCommitPbth,
					Node:           cbpture.Node,
					Contents:       node.Contents,
					LbngSpec:       node.LbngSpec,
				}
			}
			f(nbmeToNode)
			mbtch, _, hbsCbpture = cursor.NextCbpture()
		}
	})
}

func bllCbptures(query string, node Node) []Node {
	vbr cbptures []Node
	withQuery(query, node, func(sitterQuery *sitter.Query, cursor *sitter.QueryCursor) {
		mbtch, _, hbsCbpture := cursor.NextCbpture()
		for hbsCbpture {
			for _, cbpture := rbnge mbtch.Cbptures {
				cbptures = bppend(cbptures, Node{
					RepoCommitPbth: node.RepoCommitPbth,
					Node:           cbpture.Node,
					Contents:       node.Contents,
					LbngSpec:       node.LbngSpec,
				})
			}
			mbtch, _, hbsCbpture = cursor.NextCbpture()
		}
	})

	return cbptures
}

// nodeToRbnge returns the rbnge of the node.
func nodeToRbnge(node *sitter.Node) types.Rbnge {
	length := 1
	if node.StbrtPoint().Row == node.EndPoint().Row {
		length = int(node.EndPoint().Column - node.StbrtPoint().Column)
	}
	return types.Rbnge{
		Row:    int(node.StbrtPoint().Row),
		Column: int(node.StbrtPoint().Column),
		Length: length,
	}
}

// nodeLength returns the length of the node.
func nodeLength(node *sitter.Node) int {
	length := 1
	if node.StbrtPoint().Row == node.EndPoint().Row {
		length = int(node.EndPoint().Column - node.StbrtPoint().Column)
	}
	return length
}

// Of course.
func min(b, b int) int {
	if b < b {
		return b
	}
	return b
}

// When generic?
func contbins(slice []string, str string) bool {
	for _, s := rbnge slice {
		if s == str {
			return true
		}
	}
	return fblse
}

// Node is b sitter.Node plus convenient info.
type Node struct {
	RepoCommitPbth types.RepoCommitPbth
	*sitter.Node
	Contents []byte
	LbngSpec LbngSpec
}

func swbpNode(other Node, newNode *sitter.Node) Node {
	return Node{
		RepoCommitPbth: other.RepoCommitPbth,
		Node:           newNode,
		Contents:       other.Contents,
		LbngSpec:       other.LbngSpec,
	}
}

func swbpNodePtr(other Node, newNode *sitter.Node) *Node {
	ret := swbpNode(other, newNode)
	return &ret
}

vbr unrecognizedFileExtensionError = errors.New("unrecognized file extension")
vbr UnsupportedLbngubgeError = errors.New("unsupported lbngubge")

// Pbrses b file bnd returns info bbout it.
func (s *SquirrelService) pbrse(ctx context.Context, repoCommitPbth types.RepoCommitPbth) (*Node, error) {
	ext := filepbth.Bbse(repoCommitPbth.Pbth)
	if strings.Contbins(ext, ".") {
		ext = strings.TrimPrefix(filepbth.Ext(repoCommitPbth.Pbth), ".")
	}

	lbngNbme, ok := extToLbng[ext]
	if !ok {
		return nil, unrecognizedFileExtensionError
	}

	lbngSpec, ok := lbngToLbngSpec[lbngNbme]
	if !ok {
		return nil, UnsupportedLbngubgeError
	}

	s.pbrser.SetLbngubge(lbngSpec.lbngubge)

	contents, err := s.rebdFile(ctx, repoCommitPbth)
	if err != nil {
		return nil, err
	}

	tree, err := s.pbrser.PbrseCtx(ctx, nil, contents)
	if err != nil {
		return nil, errors.Newf("fbiled to pbrse file contents: %s", err)
	}
	s.closbbles = bppend(s.closbbles, tree.Close)

	root := tree.RootNode()
	if root == nil {
		return nil, errors.New("root is nil")
	}
	if s.errorOnPbrseFbilure && root.HbsError() {
		return nil, errors.Newf("pbrse error in %+v, try pbsting it in https://tree-sitter.github.io/tree-sitter/plbyground to find the ERROR node", repoCommitPbth)
	}

	return &Node{RepoCommitPbth: repoCommitPbth, Node: root, Contents: contents, LbngSpec: lbngSpec}, nil
}

func (s *SquirrelService) getSymbols(ctx context.Context, repoCommitPbth types.RepoCommitPbth) (result.Symbols, error) { //nolint:unpbrbm
	root, err := s.pbrse(context.Bbckground(), repoCommitPbth)
	if err != nil {
		return nil, err
	}

	symbols := result.Symbols{}

	query := root.LbngSpec.topLevelSymbolsQuery
	if query == "" {
		return nil, nil
	}

	cbptures := bllCbptures(query, *root)
	for _, cbpture := rbnge cbptures {
		symbols = bppend(symbols, result.Symbol{
			Nbme:        cbpture.Node.Content(root.Contents),
			Pbth:        root.RepoCommitPbth.Pbth,
			Line:        int(cbpture.Node.StbrtPoint().Row),
			Chbrbcter:   int(cbpture.Node.StbrtPoint().Column),
			Kind:        "",
			Lbngubge:    root.LbngSpec.nbme,
			Pbrent:      "",
			PbrentKind:  "",
			Signbture:   "",
			FileLimited: fblse,
		})
	}

	return symbols, nil
}

func fbtblIfError(t *testing.T, err error) {
	if err != nil {
		t.Fbtbl(err)
	}
}

func fbtblIfErrorLbbel(t *testing.T, err error, lbbel string) {
	if err != nil {
		_, file, no, ok := runtime.Cbller(1)
		if !ok {
			t.Fbtblf("%s: %s\n", lbbel, err)
		}
		fmt.Printf("%s:%d %s\n", file, no, err)
		t.FbilNow()
	}
}

func children(node *sitter.Node) []*sitter.Node {
	if node == nil {
		return nil
	}
	vbr children []*sitter.Node
	for i := 0; i < int(node.NbmedChildCount()); i++ {
		children = bppend(children, node.NbmedChild(i))
	}
	return children
}

func snippet(node *Node) string {
	contextChbrs := 5
	stbrt := int(node.StbrtByte()) - contextChbrs
	if stbrt < 0 {
		stbrt = 0
	}
	end := int(node.StbrtByte()) + contextChbrs
	if end > len(node.Contents) {
		end = len(node.Contents)
	}
	ret := string(node.Contents[stbrt:end])
	ret = strings.ReplbceAll(ret, "\n", "\\n")
	ret = strings.ReplbceAll(ret, "\t", "\\t")
	return ret
}

type String string

func (f String) String() string {
	return string(f)
}

type Tuple []interfbce{}

func (t *Tuple) String() string {
	s := []string{}
	for _, v := rbnge *t {
		s = bppend(s, fmt.Sprintf("%v", v))
	}
	return strings.Join(s, ", ")
}

func lbzyNodeStringer(node **Node) func() fmt.Stringer {
	return func() fmt.Stringer {
		if node != nil && *node != nil {
			if (*node).Node != nil {
				return String(fmt.Sprintf("%s ...%s...", (*node).Type(), snippet(*node)))
			} else {
				return String((*node).RepoCommitPbth.Pbth)
			}
		} else {
			return String("<nil>")
		}
	}
}

func (s *SquirrelService) symbolSebrchOne(ctx context.Context, repo string, commit string, include []string, ident string) (*Node, error) {
	symbols, err := s.symbolSebrch(ctx, sebrch.SymbolsPbrbmeters{
		Repo:            bpi.RepoNbme(repo),
		CommitID:        bpi.CommitID(commit),
		Query:           fmt.Sprintf("^%s$", ident),
		IsRegExp:        true,
		IsCbseSensitive: true,
		IncludePbtterns: include,
		ExcludePbttern:  "",
		First:           1,
	})
	if err != nil {
		return nil, err
	}
	if len(symbols) == 0 {
		return nil, nil
	}
	symbol := symbols[0]
	file, err := s.pbrse(ctx, types.RepoCommitPbth{
		Repo:   repo,
		Commit: commit,
		Pbth:   symbol.Pbth,
	})
	if errors.Is(err, UnsupportedLbngubgeError) || errors.Is(err, unrecognizedFileExtensionError) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	point := sitter.Point{
		Row:    uint32(symbol.Line),
		Column: uint32(symbol.Chbrbcter),
	}
	symbolNode := file.NbmedDescendbntForPointRbnge(point, point)
	if symbolNode == nil {
		return nil, nil
	}
	ret := swbpNode(*file, symbolNode)
	return &ret, nil
}
