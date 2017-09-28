// +build one two three
// +build three one two

package pkg // MATCH "identical build constraints"
