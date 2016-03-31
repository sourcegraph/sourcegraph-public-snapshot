package syntaxhighlight

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestTrie(t *testing.T) {

	actual := newTrie()

	expected := &trie{
		Children: []branch{},
		End:      true,
	}

	compareTries(t, "initial", actual, expected)

	err := actual.insert("ab")
	if err != nil {
		t.Errorf("%s Unexpected error %s", "ab", err.Error())
	}

	expected = &trie{
		Children: []branch{
			{
				V: 'a',
				T: &trie{
					Children: []branch{
						{
							V: 'b',
							T: &trie{
								Children: []branch{},
								End:      true,
							},
						},
					},
					End: false,
				},
			},
		},
		End: true,
	}

	compareTries(t, "ab", actual, expected)

	err = actual.insert("ac")
	if err != nil {
		t.Errorf("%s Unexpected error %s", "ac", err.Error())
	}

	expected = &trie{
		Children: []branch{
			{
				V: 'a',
				T: &trie{
					Children: []branch{
						{
							V: 'b',
							T: &trie{
								Children: []branch{},
								End:      true,
							},
						},
						{
							V: 'c',
							T: &trie{
								Children: []branch{},
								End:      true,
							},
						},
					},
					End: false,
				},
			},
		},
		End: true,
	}

	compareTries(t, "ac", actual, expected)

	err = actual.insert("b")
	if err != nil {
		t.Errorf("%s Unexpected error %s", "ac", err.Error())
	}

	expected = &trie{
		Children: []branch{
			{
				V: 'a',
				T: &trie{
					Children: []branch{
						{
							V: 'b',
							T: &trie{
								Children: []branch{},
								End:      true,
							},
						},
						{
							V: 'c',
							T: &trie{
								Children: []branch{},
								End:      true,
							},
						},
					},
					End: false,
				},
			},
			{
				V: 'b',
				T: &trie{
					Children: []branch{},
					End:      true,
				},
			},
		},
		End: true,
	}

	compareTries(t, "ac", actual, expected)

	err = actual.insert("abc")
	if err == nil {
		// shorter ab already exists
		t.Errorf("%s expected error, got nothing", "abc")
	}

	err = actual.insert("a")
	if err == nil {
		// longer ab already exists
		t.Errorf("%s expected error, got nothing", "a")
	}

	l := actual.lookup([]byte("abcd"), func(len int) bool {
		return true
	})

	if l != 1 {
		t.Errorf("lookup %s: expected %d, got %d", "abcd", 1, l)
	}

	l = actual.lookup([]byte("abcd"), func(len int) bool {
		return false
	})

	if l != -1 {
		t.Errorf("lookup %s: expected %d, got %d", "abcd(2)", -1, l)
	}

	l = actual.lookup([]byte("xyz"), func(len int) bool {
		return true
	})

	if l != -1 {
		t.Errorf("lookup %s: expected %d, got %d", "abcd(2)", -1, l)
	}

}

func compareTries(t *testing.T, label string, actual *trie, expected *trie) {
	if !reflect.DeepEqual(actual, expected) {
		aJson, _ := json.MarshalIndent(actual, "", "  ")
		eJson, _ := json.MarshalIndent(expected, "", "  ")
		t.Errorf("%s got != want\ngot:\n%s\nwant:\n%s",
			"ab",
			aJson,
			eJson)
	}
}
