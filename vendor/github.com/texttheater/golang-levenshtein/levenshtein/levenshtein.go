// This package implements the Levenshtein algorithm for computing the
// similarity between two strings. The central function is MatrixForStrings,
// which computes the Levenshtein matrix. The functions DistanceForMatrix,
// EditScriptForMatrix and RatioForMatrix read various interesting properties
// off the matrix. The package also provides the convenience functions
// DistanceForStrings, EditScriptForStrings and RatioForStrings for going
// directly from two strings to the property of interest.
package levenshtein

import (
	"fmt"
	"io"
	"os"
)

type EditOperation int

const (
	Ins = iota
	Del
	Sub
	Match
)

type EditScript []EditOperation

type MatchFunction func(rune, rune) bool

type Options struct {
	InsCost int
	DelCost int
	SubCost int
	Matches MatchFunction
}

// DefaultOptions is the default options: insertion cost is 1, deletion cost is
// 1, substitution cost is 2, and two runes match iff they are the same.
var DefaultOptions Options = Options{
	InsCost: 1,
	DelCost: 1,
	SubCost: 2,
	Matches: func(sourceCharacter rune, targetCharacter rune) bool {
		return sourceCharacter == targetCharacter
	},
}

func (operation EditOperation) String() string {
	if operation == Match {
		return "match"
	} else if operation == Ins {
		return "ins"
	} else if operation == Sub {
		return "sub"
	}
	return "del"
}

// DistanceForStrings returns the edit distance between source and target.
func DistanceForStrings(source []rune, target []rune, op Options) int {
	return DistanceForMatrix(MatrixForStrings(source, target, op))
}

// DistanceForMatrix reads the edit distance off the given Levenshtein matrix.
func DistanceForMatrix(matrix [][]int) int {
	return matrix[len(matrix)-1][len(matrix[0])-1]
}

// RatioForStrings returns the Levenshtein ratio for the given strings. The
// ratio is computed as follows:
//
//     (sourceLength + targetLength - distance) / (sourceLength + targetLength)
func RatioForStrings(source []rune, target []rune, op Options) float64 {
	matrix := MatrixForStrings(source, target, op)
	return RatioForMatrix(matrix)
}

// RatioForMatrix returns the Levenshtein ratio for the given matrix. The ratio
// is computed as follows:
//
//     (sourceLength + targetLength - distance) / (sourceLength + targetLength)
func RatioForMatrix(matrix [][]int) float64 {
	sourcelength := len(matrix) - 1
	targetlength := len(matrix[0]) - 1
	sum := sourcelength + targetlength

	if sum == 0 {
		return 0
	}

	dist := DistanceForMatrix(matrix)
	return float64(sum-dist) / float64(sum)
}

// MatrixForStrings generates a 2-D array representing the dynamic programming
// table used by the Levenshtein algorithm, as described e.g. here:
// http://www.let.rug.nl/kleiweg/lev/
// The reason for putting the creation of the table into a separate function is
// that it cannot only be used for reading of the edit distance between two
// strings, but also e.g. to backtrace an edit script that provides an
// alignment between the characters of both strings.
func MatrixForStrings(source []rune, target []rune, op Options) [][]int {
	// Make a 2-D matrix. Rows correspond to prefixes of source, columns to
	// prefixes of target. Cells will contain edit distances.
	// Cf. http://www.let.rug.nl/~kleiweg/lev/levenshtein.html
	height := len(source) + 1
	width := len(target) + 1
	matrix := make([][]int, height)

	// Initialize trivial distances (from/to empty string). That is, fill
	// the left column and the top row with row/column indices.
	for i := 0; i < height; i++ {
		matrix[i] = make([]int, width)
		matrix[i][0] = i
	}
	for j := 1; j < width; j++ {
		matrix[0][j] = j
	}

	// Fill in the remaining cells: for each prefix pair, choose the
	// (edit history, operation) pair with the lowest cost.
	for i := 1; i < height; i++ {
		for j := 1; j < width; j++ {
			delCost := matrix[i-1][j] + op.DelCost
			matchSubCost := matrix[i-1][j-1]
			if !op.Matches(source[i-1], target[j-1]) {
				matchSubCost += op.SubCost
			}
			insCost := matrix[i][j-1] + op.InsCost
			matrix[i][j] = min(delCost, min(matchSubCost,
				insCost))
		}
	}
	//LogMatrix(source, target, matrix)
	return matrix
}

// EditScriptForStrings returns an optimal edit script to turn source into
// target.
func EditScriptForStrings(source []rune, target []rune, op Options) EditScript {
	return backtrace(len(source), len(target),
		MatrixForStrings(source, target, op), op)
}

// EditScriptForMatrix returns an optimal edit script based on the given
// Levenshtein matrix.
func EditScriptForMatrix(matrix [][]int, op Options) EditScript {
	return backtrace(len(matrix)-1, len(matrix[0])-1, matrix, op)
}

// WriteMatrix writes a visual representation of the given matrix for the given
// strings to the given writer.
func WriteMatrix(source []rune, target []rune, matrix [][]int, writer io.Writer) {
	fmt.Fprintf(writer, "    ")
	for _, targetRune := range target {
		fmt.Fprintf(writer, "  %c", targetRune)
	}
	fmt.Fprintf(writer, "\n")
	fmt.Fprintf(writer, "  %2d", matrix[0][0])
	for j, _ := range target {
		fmt.Fprintf(writer, " %2d", matrix[0][j+1])
	}
	fmt.Fprintf(writer, "\n")
	for i, sourceRune := range source {
		fmt.Fprintf(writer, "%c %2d", sourceRune, matrix[i+1][0])
		for j, _ := range target {
			fmt.Fprintf(writer, " %2d", matrix[i+1][j+1])
		}
		fmt.Fprintf(writer, "\n")
	}
}

// LogMatrix writes a visual representation of the given matrix for the given
// strings to os.Stderr. This function is deprecated, use
// WriteMatrix(source, target, matrix, os.Stderr) instead.
func LogMatrix(source []rune, target []rune, matrix [][]int) {
	WriteMatrix(source, target, matrix, os.Stderr)
}

func backtrace(i int, j int, matrix [][]int, op Options) EditScript {
	if i > 0 && matrix[i-1][j]+op.DelCost == matrix[i][j] {
		return append(backtrace(i-1, j, matrix, op), Del)
	}
	if j > 0 && matrix[i][j-1]+op.InsCost == matrix[i][j] {
		return append(backtrace(i, j-1, matrix, op), Ins)
	}
	if i > 0 && j > 0 && matrix[i-1][j-1]+op.SubCost == matrix[i][j] {
		return append(backtrace(i-1, j-1, matrix, op), Sub)
	}
	if i > 0 && j > 0 && matrix[i-1][j-1] == matrix[i][j] {
		return append(backtrace(i-1, j-1, matrix, op), Match)
	}
	return []EditOperation{}
}

func min(a int, b int) int {
	if b < a {
		return b
	}
	return a
}

func max(a int, b int) int {
	if b > a {
		return b
	}
	return a
}
