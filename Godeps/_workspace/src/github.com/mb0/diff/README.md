diff
====

A difference algorithm package for go.

The algorithm is described by Eugene Myers in
["An O(ND) Difference Algorithm and its Variations"](http://www.xmailserver.org/diff2.pdf).

Example
-------
You can use diff.Ints, diff.Runes, diff.ByteStrings, and diff.Bytes

    diff.Runes([]rune("sögen"), []rune("mögen")) // returns []Changes{{0,0,1,1}}

or you can implement diff.Data

    type MixedInput struct {
    	A []int
    	B []string
    }
    func (m *MixedInput) Equal(i, j int) bool {
    	return m.A[i] == len(m.B[j])
    }

and call

    m := &MixedInput{..}
    diff.Diff(len(m.A), len(m.B), m)

Also has granularity functions to merge changes that are close by.

    diff.Granular(1, diff.ByteStrings("emtire", "umpire")) // returns []Changes{{0,0,3,3}}

Documentation at http://godoc.org/github.com/mb0/diff
