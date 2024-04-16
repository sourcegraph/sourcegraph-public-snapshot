package bytesize

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input string
		want  Bytes
	}{
		// Happy paths
		{"1 B", 1},

		{"1 KB", 1000},
		{"1 KiB", 1024},

		{"1 MB", 1_000_000},
		{"1 MiB", 1024 * 1024},

		{"1 GB", 1_000_000_000},
		{"1 GiB", 1024 * 1024 * 1024},

		// Whitespace
		{"  1       B", 1},
		{"  100       KB", 100_000},

		// Various
		{"100 B", 100},
		{"1024 B", 1024},

		{"12 KB", 12_000},
		{"72 KiB", 73_728},

		{"3 MB", 3000000},
		{"3 MiB", 3145728},

		{"7 GB", 7_000_000_000},
		{"7 GiB", 7_516_192_768},
	}

	for _, tt := range tests {
		got, err := Parse(tt.input)
		if err != nil {
			t.Errorf("FromString(%q), got error: %s", tt.input, err)
		}

		if got != tt.want {
			t.Errorf("FromString(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}

	invalid := []string{
		"  ",
		"aaa",
		"1",
		"1 count",
		"1 megabyte",
		"10 horses",
		"1324 k b",
	}
	for _, tt := range invalid {
		_, err := Parse(tt)
		if err == nil {
			t.Errorf("Parse(%q) expected error but got nil", tt)
		}
	}
}

func TestParseOverflow(t *testing.T) {
	input := fmt.Sprintf("%d KB", (maxSize / 2))
	_, err := Parse(input)
	if err == nil {
		t.Errorf("Parse(%q) expected error, but got none", input)
	}
}

func TestReadNumber(t *testing.T) {
	tests := []struct {
		input string
		want  int
		len   int
	}{
		{"1234 foobar", 1234, 4},
		{"1234foobar", 1234, 4},
		{"12 34", 12, 2},
		{"1f", 1, 1},
		{"foobar", 0, 0},
	}

	for _, tt := range tests {
		num, numLen, err := readNumber(tt.input)
		require.NoError(t, err)

		if num != tt.want {
			t.Errorf("readNumber(%q) = %d, want %d", tt.input, num, tt.want)
		}
		if numLen != tt.len {
			t.Errorf("readNumber(%q) length = %d, want %d", tt.input, numLen, tt.len)
		}
	}
}

func TestMultiplication(t *testing.T) {
	tests := []struct {
		inputNumber int
		inputUnit   Bytes
		want        Bytes
	}{
		{10, B, 10},
		{25, KB, 25_000},
		{25_000, KB, 25_000_000},
	}

	for _, tt := range tests {
		got := Bytes(tt.inputNumber) * tt.inputUnit
		if got != tt.want {
			t.Errorf("FromUnit(%d, %v) = %d, want %d", tt.inputNumber, tt.inputUnit, got, tt.want)
		}
	}
}
