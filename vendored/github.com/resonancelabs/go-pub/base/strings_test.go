package base

import (
	"fmt"
	"testing"
)

// A simple test that doesn't require the gocheck fixture.
func TestPostSnippet(t *testing.T) {
	tests := []struct {
		prose, expected string
	}{
		{"t t t", "t t t"},
		{"test", "test"},
		{"verylongword", "verylongword"},
		{"test test test", "test..."},
		{"test a test", "test a..."},
		{"testing test", "testing..."},
	}

	for _, tc := range tests {
		result := MessageSnippetOfLength(tc.prose, 6)
		if result != tc.expected {
			t.Error(fmt.Sprintf("expected '%s', got '%s' from '%s'", tc.expected, result, tc.prose))
		}
	}
}

func TestMessageContainsPhrase(t *testing.T) {
	tests := []struct {
		prose, phrase string
		expected      bool
	}{
		{"abracadabra", "bra", false},
		{"kitty", "kitty", true},
		{" kitty", "kitty", true},
		{"kitty ", "kitty", true},
		{" kitty ", "kitty", true},
		{"go whitespace", "go whitespace", true},
		{" go whitespace", "go whitespace", true},
		{"go whitespace ", "go whitespace", true},
		{" go whitespace ", "go whitespace", true},
		{"no play makes jack a dull boy", "ack a", false},
		{"no play makes jack a dull boy", "jack a", true},
		{"no play makes jack a dull boy", "jack a dull", true},
		{"no play makes jack a dullZ boy", "jack a dull", false},
	}

	for _, tc := range tests {
		result := MessageContainsPhrase(tc.prose, tc.phrase)
		if result != tc.expected {
			t.Error(fmt.Sprintf("expected %v from MessageContainsPhrase('%s', '%s')", tc.expected, tc.prose, tc.phrase))
		}
	}
}

func TestToUnderscoreCase(t *testing.T) {
	tests := []struct {
		str, expected string
	}{
		{"", ""},
		{"a", "a"},
		{"A", "a"},
		{"AlphaBeta", "alpha_beta"},
		{"alpha_beta", "alpha_beta"},
		{"_Gamma_DeltaEpsilon", "_gamma_delta_epsilon"},
		{"AbCdE", "ab_cd_e"},
		{"A123", "a123"},
		{"b2_A", "b2_a"},
	}
	for _, test := range tests {
		if result := ToUnderscoreCase(test.str); result != test.expected {
			t.Error(fmt.Sprintf("expected %v, found %v", test.expected, result))
		}
	}
}

func TestMessageContainsPhoneNumber(t *testing.T) {
	tests := []struct {
		str      string
		expected bool
	}{
		{"2112223333", false},
		{"211-222-3333", true},
		{"(211) -222-3333", true},
		{"211 a222-3333", false},
		{"211 222-abcd", false},
		{"211.222.3333", true},
		{"211 222 3333", true},
		{"111 222 3333", false},
	}
	for _, test := range tests {
		if result := MessageContainsPhoneNumber(test.str); result != test.expected {
			t.Error(fmt.Sprintf("expected %v, found %v", test.expected, result))
		}
	}
}

func TestMessageContainsPotentialProperName(t *testing.T) {
	tests := []struct {
		str      string
		expected []string
	}{
		{"I personally think Kim Kardashian is the worst.", []string{"Kim Kardashian"}},
		{"I Hope this works!", []string{}},
		{"I can't wait for the new Bruce Lee movie. My Father loves him!", []string{"Bruce Lee"}},
		{"Ben, Ben, and Jenn are my favorites (they rhyme?)", []string{}},
		{"HELLO MY NAME IS CRAZY PERSON", []string{"CRAZY PERSON"}},
		{"My favorite movie is Top Gun with Tom Cruise.", []string{"Top Gun", "Tom Cruise"}},
	}
	for _, test := range tests {
		result := MessageContainsPotentialProperName(test.str)
		if len(result) != len(test.expected) {
			t.Error(fmt.Sprintf("expected %v, found %v", test.expected, result))
		}
		for index, name := range result {
			if name != test.expected[index] {
				t.Error(fmt.Sprintf("expected %v, found %v", test.expected, result))
			}
		}
	}
}

func TestJoinSliceFormatted(t *testing.T) {
	if JoinSliceFormatted([]string{""}, ",") != "" {
		t.Error("Error")
	}
	if JoinSliceFormatted([]string{"A", "B", "C"}, "") != "ABC" {
		t.Error("Error")
	}
	if JoinSliceFormatted([]int64{0, 1, 2}, ", ") != "0, 1, 2" {
		t.Error("Error")
	}
}
