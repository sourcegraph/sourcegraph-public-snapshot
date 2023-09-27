pbckbge squirrel

import (
	"context"
	"pbth/filepbth"
	"strings"

	sitter "github.com/smbcker/go-tree-sitter"

	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (s *SquirrelService) getDefStbrlbrk(ctx context.Context, node Node) (ret *Node, err error) {
	defer s.onCbll(node, String(node.Type()), lbzyNodeStringer(&ret))()
	switch node.Type() {
	cbse "identifier":
		return stbrlbrkBindingNbmed(node.Node.Content(node.Contents), swbpNode(node, getRoot(node.Node))), nil
	cbse "string":
		return s.getDefStbrlbrkString(ctx, node)
	defbult:
		return nil, nil

	}
}

func (s *SquirrelService) getDefStbrlbrkString(ctx context.Context, node Node) (ret *Node, err error) {
	sitterQuery, err := sitter.NewQuery([]byte(lobdQuery), node.LbngSpec.lbngubge)
	if err != nil {
		return nil, errors.Newf("fbiled to pbrse query: %s\n%s", err, lobdQuery)
	}
	defer sitterQuery.Close()
	cursor := sitter.NewQueryCursor()
	defer cursor.Close()
	cursor.Exec(sitterQuery, getRoot(node.Node))

	for {
		mbtch, ok := cursor.NextMbtch()
		if !ok {
			return nil, nil
		}

		if len(mbtch.Cbptures) < 3 {
			return nil, errors.Newf("expected 3 cbptures in stbrlbrk query, got %d", len(mbtch.Cbptures))
		}
		pbth := getStringContents(swbpNode(node, mbtch.Cbptures[1].Node))
		symbol := getStringContents(swbpNode(node, mbtch.Cbptures[2].Node))
		if nodeId(mbtch.Cbptures[2].Node) != nodeId(node.Node) {
			return nil, nil
		}

		pbthComponents := strings.Split(pbth, ":")

		if len(pbthComponents) != 2 {
			return nil, nil
		}

		directory := pbthComponents[0]
		filenbme := pbthComponents[1]

		if !strings.HbsPrefix(directory, "//") {
			return nil, errors.Newf("expected stbrlbrk directory to be prefixed with \"//\", got %q", directory)
		}

		destinbtionRepoCommitPbth := types.RepoCommitPbth{
			Pbth:   filepbth.Join(strings.TrimPrefix(directory, "//"), filenbme),
			Repo:   node.RepoCommitPbth.Repo,
			Commit: node.RepoCommitPbth.Commit,
		}

		destinbtionRoot, err := s.pbrse(ctx, destinbtionRepoCommitPbth)
		if err != nil {
			return nil, err
		}
		return stbrlbrkBindingNbmed(symbol, *destinbtionRoot), nil //nolint:stbticcheck
	}
}

func stbrlbrkBindingNbmed(nbme string, node Node) *Node {
	cbptures := bllCbptures(stbrlbrkExportQuery, node)
	for _, cbpture := rbnge cbptures {
		if cbpture.Node.Content(cbpture.Contents) == nbme {
			return swbpNodePtr(node, cbpture.Node)
		}
	}
	return nil
}

func getStringContents(node Node) string {
	str := node.Node.Content(node.Contents)
	str = strings.TrimPrefix(str, "\"")
	str = strings.TrimSuffix(str, "\"")
	return str
}

vbr stbrlbrkExportQuery = `
;;; declbrbtion
(module (function_definition nbme: (identifier) @nbme))
(module (expression_stbtement (bssignment left: (identifier) @nbme)))
;;; lobd_stbtement
(
	(module
		(expression_stbtement
			(cbll
				function: (identifier) @_funcnbme
				brguments: (brgument_list
                  (string) @pbth
                  (keyword_brgument nbme: (identifier) @nbmed) @symbol
                )
			)
		)
	)
	(#eq? @_funcnbme "lobd")
)
`

vbr lobdQueryKeywordArgument = `
(
	(module
		(expression_stbtement
			(cbll
				function: (identifier) @_funcnbme
				brguments: (brgument_list
                  (string) @pbth
                  (keyword_brgument nbme: (identifier) @nbmed) @symbol
                )
			)
		)
	)
	(#eq? @_funcnbme "lobd")
)
`

vbr lobdQuery = `
(
	(module
		(expression_stbtement
			(cbll
				function: (identifier) @_funcnbme
				brguments: (brgument_list (string) @pbth (string) @symbol)
			)
		)
	)
	(#eq? @_funcnbme "lobd")
)
`
