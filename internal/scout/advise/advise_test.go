package advise

import (
	"bufio"
	"context"
	"os"
	"testing"
	"time"

	"github.com/sourcegraph/src-cli/internal/scout"
)

func TestCheckUsage(t *testing.T) {
	cases := []struct {
		name         string
		usage        float64
		resourceType string
		container    string
		want         string
	}{
		{
			name:         "should return correct message for usage over 100",
			usage:        110,
			resourceType: "cpu",
			container:    "gitserver-0",
			want:         "\t🚨 gitserver-0: Your cpu usage is over 100% (110.00%). Add more cpu.",
		},
		{
			name:         "should return correct message for usage over 80 and under 100",
			usage:        87,
			resourceType: "memory",
			container:    "gitserver-0",
			want:         "\t⚠️  gitserver-0: Your memory usage is over 80% (87.00%). Consider raising limits.",
		},
		{
			name:         "should return correct message for usage over 40 and under 80",
			usage:        63.4,
			resourceType: "memory",
			container:    "gitserver-0",
			want:         "\t✅ gitserver-0: Your memory usage is under 80% (63.40%). Keep memory allocation the same.",
		},
		{
			name:         "should return correct message for usage under 40",
			usage:        22.33,
			resourceType: "memory",
			container:    "gitserver-0",
			want:         "\t⚠️  gitserver-0: Your memory usage is under 40% (22.33%). Consider lowering limits.",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := CheckUsage(tc.usage, tc.resourceType, tc.container)

			if got != tc.want {
				t.Errorf("got: '%s' want '%s'", got, tc.want)
			}
		})
	}
}

func TestOutputToFile(t *testing.T) {
	cfg := &scout.Config{
		Output: os.TempDir() + string(os.PathSeparator) + "test.txt",
	}
	name := "gitserver-0"
	advice := []string{
		"Add more CPU",
		"Add more memory",
	}

	err := OutputToFile(context.Background(), cfg, name, advice)
	if err != nil {
		t.Fatal(err)
	}

	lines := readOutputFile(t, cfg)

	cases := []struct {
		lineNum int
		want    string
	}{
		{1, "- gitserver-0"},
		{2, "Add more CPU"},
		{3, "Add more memory"},
	}

	for _, tc := range cases {
		tc := tc
		if lines[tc.lineNum-1] != tc.want {
			t.Errorf("Expected %q, got %q", tc.want, lines[tc.lineNum-1])
		}
	}

	if len(lines) > 3 {
		t.Error("Expected only 3 lines, got more")
	}

	if err != nil {
		t.Fatal(err)
	}
}

func readOutputFile(t *testing.T, cfg *scout.Config) []string {
	file, err := os.Open(cfg.Output)
	if err != nil {
		t.Fatal(err)
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	file.Close()
	err = os.Remove(cfg.Output)
	if err != nil {
		// try again after waiting a bit
		time.Sleep(100 * time.Millisecond)
		err = os.Remove(cfg.Output)
		if err != nil {
			t.Fatal(err)
		}
	}
	return lines
}
