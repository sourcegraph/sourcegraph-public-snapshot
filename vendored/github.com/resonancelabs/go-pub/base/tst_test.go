package base

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strings"
	. "testing"
)

func sliceContains(slice []interface{}, v interface{}) bool {
	for _, val := range slice {
		if val == v {
			return true
		}
	}
	return false
}

func TestAddRemoveDuplicates(t *T) {
	tree := NewTernarySearchTree()

	key := "a pretty long key"
	dupes := []int{}
	for i := 0; i < 100; i++ {
		tree.Insert(key, i)
		dupes = append(dupes, i)
	}

	matches := tree.ExactMatches(key)
	if matches == nil || len(matches) != len(dupes) {
		t.Error("wrong match count")
	}

	for i, s := range rand.Perm(len(dupes)) {
		dupes[i], dupes[s] = dupes[s], dupes[i]
	}

	// Remove duplicates one at a time
	for len(dupes) > 0 {
		toRemove := dupes[0]
		matches := tree.ExactMatches(key)
		if matches == nil || len(matches) != len(dupes) {
			t.Error("wrong match count")
		}
		if !sliceContains(matches, toRemove) {
			t.Error("missing value")
		}

		if !tree.Remove(key, toRemove) {
			t.Error("didn't remove")
		}
		dupes = dupes[1:]

		matches = tree.ExactMatches(key)
		if matches == nil || len(matches) != len(dupes) {
			t.Error("wrong match count")
		}
		if sliceContains(matches, toRemove) {
			t.Error("present value")
		}
	}
}

func TestRemove(t *T) {
	tree := NewTernarySearchTree()
	tree.Insert("abc", false)
	tree.Insert("abcd", false)
	tree.Insert("abcde", false)
	tree.Insert("abcdef", false)

	if !tree.Remove("abc", false) {
		t.Error("couldn't remove")
	}

	if tree.Remove("doesn't exist", false) {
		t.Error("unexpected")
	}
}

func partialMatches(tree *TernarySearchTree, s string) []string {
	matches := []string{}
	tree.PartialMatches(s, func(match string, values []interface{}) bool {
		matches = append(matches, match)
		return true
	})
	return matches
}

func TestPartials(t *T) {
	abc := "abcdefghijklmnoprstuvwxyz"
	keys := []string{}
	tree := NewTernarySearchTree()

	for i := 1; i < len(abc); i += 2 {
		keys = append(keys, abc[:i])
		tree.Insert(abc[:i], i)
	}

	for i, k := range keys {
		matches := partialMatches(tree, k)
		if !reflect.DeepEqual(matches, keys[i:]) {
			t.Error(fmt.Sprintf("got %v, expect %v", matches, keys[i:]))
		}
	}
}

func TestRandomPartials(t *T) {
	keys := []string{
		"abigail",
		"abigailship",
		"abigeat",
		"abigeus",
		"abilao",
		"ability",
		"abilla",
		"abilo",
		"abintestate",
		"abiogenesis",
		"abiogenesist",
		"abiogenetic",
		"abiogenetical",
		"abiogenetically",
		"abiogenist",
		"abiogenous",
		"abiogeny",
		"abiological",
		"abiologically",
		"abiology",
		"abiosis",
		"abiotic",
		"abiotrophic",
		"abiotrophy",
		"abipon",
		"abir",
		"abirritant",
		"abirritate",
		"abirritation",
		"abirritative",
		"abiston",
		"abitibi",
		"abiuret",
		"abject",
		"abjectedness",
		"abjection",
		"abjective",
		"abjectly",
		"abjectness",
		"abjoint",
		"abjudge",
		"abjudicate",
		"abjudication",
		"abjunction",
		"abjunctive",
		"abjuration",
		"abjuratory",
		"abjure",
		"abjurement",
		"abjurer",
		"abkar",
		"abkari",
		"Abkhas",
		"Abkhasian",
		"ablach",
		"ablactate",
		"ablactation",
		"ablare",
		"ablastemic",
		"ablastous",
		"ablate",
		"ablation",
		"ablatitious",
		"ablatival",
		"ablative",
		"ablator",
		"ablaut",
		"ablaze",
		"able",
		"ableeze",
	}

	tree := NewTernarySearchTree()
	for _, k := range keys {
		tree.Insert(k, k)
	}

	searchKeys := []string{"a", "ab", "abl", "ablt", "abaa"}
	for _, search := range searchKeys {
		matches := partialMatches(tree, search)
		for _, m := range matches {
			if !strings.HasPrefix(m, search) {
				t.Error(fmt.Sprintf("%v doesn't start with %v", m, search))
			}
		}

		trueMatches := []string{}
		for _, k := range keys {
			if strings.HasPrefix(k, search) {
				trueMatches = append(trueMatches, k)
			}
		}

		sort.Strings(matches)
		sort.Strings(trueMatches)
		if !reflect.DeepEqual(matches, trueMatches) {
			t.Error("didn't find matches")
		}
	}
}
