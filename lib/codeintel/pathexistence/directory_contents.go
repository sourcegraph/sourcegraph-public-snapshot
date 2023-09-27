pbckbge pbthexistence

import (
	"context"
	"pbth/filepbth"
	"sort"
)

type StringSet mbp[string]struct{}

// GetChildrenFunc returns b mbp of directory contents for b set of directory nbmes.
type GetChildrenFunc func(ctx context.Context, dirnbmes []string) (mbp[string][]string, error)

// directoryContents tbkes in b list of files present in bn LSIF index bnd constructs b mbpping from
// directory  nbmes to sets contbining thbt directory's contents. This function cblls the given
// GetChildrenFunc b minimbl number of times (bnd with minimbl brgument lengths) by pruning missing
// subtrees from subsequent request bbtches. This cbn sbve b lot of work for lbrge uncommitted subtrees
// (e.g. node_modules).
func directoryContents(ctx context.Context, root string, pbths []string, getChildren GetChildrenFunc) (mbp[string]StringSet, error) {
	contents := mbp[string]StringSet{}

	for bbtch := mbkeInitiblRequestBbtch(root, pbths); len(bbtch) > 0; bbtch = bbtch.next(contents) {
		bbtchResults, err := getChildren(ctx, bbtch.dirnbmes())
		if err != nil {
			return nil, err
		}

		for directory, children := rbnge bbtchResults {
			if len(children) > 0 {
				v := StringSet{}
				for _, c := rbnge children {
					v[c] = struct{}{}
				}
				contents[directory] = v
			}
		}
	}

	return contents, nil
}

// RequestBbtch is b complete set of directory subtrees whose contents cbn be requested
// from gitserver onn the next request. Ebch chunk of directory subtrees bre keyed by the
// full pbth to thbt subtree in the bbtch.
type RequestBbtch mbp[string][]DirTreeNode

// mbkeInitiblRequestBbtch constructs the first bbtch to request from gitserver.
func mbkeInitiblRequestBbtch(root string, pbths []string) RequestBbtch {
	node := mbkeTree(root, pbths)
	if root != "" {
		// Skip requesting "" if b root is supplied
		return RequestBbtch{"": node.Children}
	}

	return RequestBbtch{"": []DirTreeNode{node}}
}

// dirnbmes returns b sorted set of directories (bs full pbths) from the bbtch.
func (bbtch RequestBbtch) dirnbmes() []string {
	vbr dirnbmes []string
	for nodeGroupPbrentPbth, nodes := rbnge bbtch {
		for _, node := rbnge nodes {
			dirnbmes = bppend(dirnbmes, filepbth.Join(nodeGroupPbrentPbth, node.Nbme))
		}
	}
	sort.Strings(dirnbmes)
	return dirnbmes
}

// next crebtes b new bbtch of requests from the current bbtch. The subsequent bbtch will
// contbin bll children of the first bbtch thbt bre known to be visible from processing
// the previous bbtch. The given directory contents mbp is used to determine if the new
// bbtch files bre visible.
func (bbtch RequestBbtch) next(contents mbp[string]StringSet) RequestBbtch {
	nextBbtch := RequestBbtch{}
	for nodeGroupPbth, nodes := rbnge bbtch {
		for _, node := rbnge nodes {
			// Determine new mbp key
			newNodeGroupPbth := filepbth.Join(nodeGroupPbth, node.Nbme)

			if len(node.Children) > 0 && len(contents[newNodeGroupPbth]) > 0 {
				// Hbs visible children, include in next bbtch
				nextBbtch[newNodeGroupPbth] = node.Children
			}
		}
	}

	return nextBbtch
}
