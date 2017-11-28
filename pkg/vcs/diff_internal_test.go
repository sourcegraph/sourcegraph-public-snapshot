package vcs

import "testing"

func TestTruncateLongLines(t *testing.T) {
	const maxCharsPerLine = 5

	tests := map[string]string{
		"":       "",
		"1":      "1",
		"12345":  "12345",
		"123456": "12345",
		"一二三四五六七八九十":       "一二三四五",
		"一二三四五六七\n一二三四五六七": "一二三四五\n一二三四五",
		"😄😱👽😁😘😤😸":          "😄😱👽😁😘",
		"😄😱👽😁😘😤😸\n😄😱👽😁😘😤😸": "😄😱👽😁😘\n😄😱👽😁😘",
	}

	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			got := truncateLongLines([]byte(input), maxCharsPerLine)
			if string(got) != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}
