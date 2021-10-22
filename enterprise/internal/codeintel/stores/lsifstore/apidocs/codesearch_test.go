package apidocs

import (
	"testing"

	"github.com/hexops/autogold"
	"github.com/keegancsmith/sqlf"
)

func TestLexemes(t *testing.T) {
	testCases := []struct {
		input string
		want  autogold.Value
	}{
		{"", autogold.Want("empty string", []string{})},
		{"f", autogold.Want("single alphabetical", []string{"f"})},
		{".", autogold.Want("single punctuation", []string{"."})},
		{"f.", autogold.Want("single alphabetical and punctuation", []string{"f", "."})},
		{"foo.bar.baz", autogold.Want("basic", []string{"foo", ".", "bar", ".", "baz"})},
		{"foo::bar'a new Baz().bar//efg", autogold.Want("complex", []string{
			"foo", ":", ":", "bar", "'", "a", "new", "Baz", "(",
			")",
			".",
			"bar",
			"/",
			"/",
			"efg",
		})},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			got := Lexemes(tc.input)
			tc.want.Equal(t, got)
		})
	}
}

func TestTextSearchVector(t *testing.T) {
	testCases := []struct {
		input string
		want  autogold.Value
	}{
		{"", autogold.Want("empty string", "")},
		{"hello world", autogold.Want("english", "hello:1 world:2")},
		{"http.Router", autogold.Want("basic", "http:1 .:2 Router:3")},
		{"go github.com/golang/go private struct http.Router", autogold.Want("complex", "go:1 github:2 .:3 com:4 /:5 golang:6 /:7 go:8 private:9 struct:10 http:11 .:12 Router:13")},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			got := TextSearchVector(tc.input)
			tc.want.Equal(t, got)
		})
	}
}

func TestTextSearchRank(t *testing.T) {
	testCases := []struct {
		input            string
		subStringMatches bool
		want             autogold.Value
	}{
		{"", true, autogold.Want("empty string", [2]interface{}{"", []interface{}{}})},
		{"mux Router", true, autogold.Want("basic", [2]interface{}{
			"ts_rank_cd(column, $1, 2) + ts_rank_cd(column, $2, 2)",
			[]interface{}{
				"mux:*",
				"Router:*",
			},
		})},
		{"public struct github.com/gorilla/mux mux.Router", true, autogold.Want("complex", [2]interface{}{
			"ts_rank_cd(column, $1, 2) + ts_rank_cd(column, $2, 2) + ts_rank_cd(column, $3, 2) + ts_rank_cd(column, $4, 2)",
			[]interface{}{
				"public:*",
				"struct:*",
				"github:* <-> .:* <-> com:* <-> /:* <-> gorilla:* <-> /:* <-> mux:*",
				"mux:* <-> .:* <-> Router:*",
			},
		})},
		{"public struct github.com/gorilla/mux mux.Router", false, autogold.Want("complex no substring matching", [2]interface{}{
			"ts_rank_cd(column, $1, 2) + ts_rank_cd(column, $2, 2) + ts_rank_cd(column, $3, 2) + ts_rank_cd(column, $4, 2)",
			[]interface{}{
				"public",
				"struct",
				"github <-> . <-> com <-> / <-> gorilla <-> / <-> mux",
				"mux <-> . <-> Router",
			},
		})},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			q := TextSearchRank("column", tc.input, tc.subStringMatches)
			query := q.Query(sqlf.PostgresBindVar)
			got := [2]interface{}{query, q.Args()}
			tc.want.Equal(t, got)
		})
	}
}

func TestTextSearchQuery(t *testing.T) {
	testCases := []struct {
		input            string
		subStringMatches bool
		want             autogold.Value
	}{
		{"", true, autogold.Want("empty string", [2]interface{}{
			"column @@ $1 OR column @@ $2 OR column @@ $3 OR column @@ $4",
			[]interface{}{
				"",
				"",
				"",
				"",
			},
		})},
		{"mux Router", true, autogold.Want("basic", [2]interface{}{
			"column @@ $1 OR column @@ $2 OR column @@ $3 OR column @@ $4",
			[]interface{}{
				"mux:* <-> Router:*",
				"mux:* <2> Router:*",
				"mux:* <4> Router:*",
				"mux:* <5> Router:*",
			},
		})},
		{"github.com/gorilla/mux", true, autogold.Want("whole query terms are matched in exact sequence using <->", [2]interface{}{
			"column @@ $1 OR column @@ $2 OR column @@ $3 OR column @@ $4",
			[]interface{}{
				"github:* <-> .:* <-> com:* <-> /:* <-> gorilla:* <-> /:* <-> mux:*",
				"github:* <-> .:* <-> com:* <-> /:* <-> gorilla:* <-> /:* <-> mux:*",
				"github:* <-> .:* <-> com:* <-> /:* <-> gorilla:* <-> /:* <-> mux:*",
				"github:* <-> .:* <-> com:* <-> /:* <-> gorilla:* <-> /:* <-> mux:*",
			},
		})},
		{"github.com gorilla mux", true, autogold.Want("separate query terms are matched even if there is distance between using <N>", [2]interface{}{
			"column @@ $1 OR column @@ $2 OR column @@ $3 OR column @@ $4",
			[]interface{}{
				"github:* <-> .:* <-> com:* <-> gorilla:* <-> mux:*",
				"github:* <-> .:* <-> com:* <2> gorilla:* <2> mux:*",
				"github:* <-> .:* <-> com:* <4> gorilla:* <4> mux:*",
				"github:* <-> .:* <-> com:* <5> gorilla:* <5> mux:*",
			},
		})},
		{"public struct github.com/gorilla/mux mux.Router", true, autogold.Want("complex", [2]interface{}{
			"column @@ $1 OR column @@ $2 OR column @@ $3 OR column @@ $4",
			[]interface{}{
				"public:* <-> struct:* <-> github:* <-> .:* <-> com:* <-> /:* <-> gorilla:* <-> /:* <-> mux:* <-> mux:* <-> .:* <-> Router:*",
				"public:* <2> struct:* <2> github:* <-> .:* <-> com:* <-> /:* <-> gorilla:* <-> /:* <-> mux:* <2> mux:* <-> .:* <-> Router:*",
				"public:* <4> struct:* <4> github:* <-> .:* <-> com:* <-> /:* <-> gorilla:* <-> /:* <-> mux:* <4> mux:* <-> .:* <-> Router:*",
				"public:* <5> struct:* <5> github:* <-> .:* <-> com:* <-> /:* <-> gorilla:* <-> /:* <-> mux:* <5> mux:* <-> .:* <-> Router:*",
			},
		})},
		{"public struct github.com/gorilla/mux mux.Router", false, autogold.Want("complex no substring matching", [2]interface{}{
			"column @@ $1 OR column @@ $2 OR column @@ $3 OR column @@ $4",
			[]interface{}{
				"public <-> struct <-> github <-> . <-> com <-> / <-> gorilla <-> / <-> mux <-> mux <-> . <-> Router",
				"public <2> struct <2> github <-> . <-> com <-> / <-> gorilla <-> / <-> mux <2> mux <-> . <-> Router",
				"public <4> struct <4> github <-> . <-> com <-> / <-> gorilla <-> / <-> mux <4> mux <-> . <-> Router",
				"public <5> struct <5> github <-> . <-> com <-> / <-> gorilla <-> / <-> mux <5> mux <-> . <-> Router",
			},
		})},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			q := TextSearchQuery("column", tc.input, tc.subStringMatches)
			query := q.Query(sqlf.PostgresBindVar)
			got := [2]interface{}{query, q.Args()}
			tc.want.Equal(t, got)
		})
	}
}

func TestQuery(t *testing.T) {
	testCases := []struct {
		input string
		want  autogold.Value
	}{
		{"", autogold.Want("empty string", Query{SubStringMatches: true})},
		{"mux Router", autogold.Want("basic", Query{
			MetaTerms: "mux Router", MainTerms: "mux Router",
			SubStringMatches: true,
		})},
		{"github.com/gorilla/mux", autogold.Want("repository name", Query{
			MetaTerms: "github.com/gorilla/mux", MainTerms: "github.com/gorilla/mux",
			SubStringMatches: true,
		})},
		{"github.com gorilla mux", autogold.Want("repository name as separate terms", Query{
			MetaTerms: "github.com gorilla mux", MainTerms: "github.com gorilla mux",
			SubStringMatches: true,
		})},
		{"public struct github.com/gorilla/mux mux.Router", autogold.Want("complex query", Query{
			MetaTerms:        "public struct github.com/gorilla/mux mux.Router",
			MainTerms:        "public struct github.com/gorilla/mux mux.Router",
			SubStringMatches: true,
		})},
		{"public struct github.com/gorilla/mux: mux.Router", autogold.Want("metadata separated", Query{
			MetaTerms:        "public struct github.com/gorilla/mux",
			MainTerms:        "mux.Router",
			SubStringMatches: true,
		})},
		{"public struct github.com/gorilla/mux: mux::Router", autogold.Want("metadata separated, colon later", Query{
			MetaTerms:        "public struct github.com/gorilla/mux",
			MainTerms:        "mux::Router",
			SubStringMatches: true,
		})},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			got := ParseQuery(tc.input)
			tc.want.Equal(t, got)
		})
	}
}
