package mafsa

import "os"

// New constructs a new, empty MA-FSA that can be filled with data.
func New() *BuildTree {
	t := new(BuildTree)
	t.register = make(map[string]*BuildTreeNode)
	t.Root = new(BuildTreeNode)
	t.Root.Edges = make(map[rune]*BuildTreeNode)
	return t
}

// Load loads an existing MA-FSA from a file specified by filename.
// It returns a read-only MA-FSA, or an error if loading failed.
func Load(filename string) (*MinTree, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	mtree, err := new(Decoder).ReadFrom(f)
	if err != nil {
		return nil, err
	}

	return mtree, nil
}
