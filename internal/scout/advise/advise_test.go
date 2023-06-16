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
		want         scout.Advice
	}{
		{
			name:         "should return correct message for usage over 100",
			usage:        110,
			resourceType: "cpu",
			container:    "gitserver-0",
			want: scout.Advice{
				Kind: scout.DANGER,
				Msg:  "🚨 gitserver-0: cpu is under-provisioned (110.00% usage). Add resources.",
			},
		},
		{
			name:         "should return correct message for usage over 80 and under 100",
			usage:        87,
			resourceType: "memory",
			container:    "gitserver-0",
			want: scout.Advice{
				Kind: scout.DANGER,
				Msg:  "🚨 gitserver-0: memory is under-provisioned (87.00% usage). Add resources.",
			},
		},
		{
			name:         "should return correct message for usage over 40 and under 80",
			usage:        63.4,
			resourceType: "memory",
			container:    "gitserver-0",
			want: scout.Advice{
				Kind: scout.HEALTHY,
				Msg:  "✅ gitserver-0: memory is well-provisioned (63.40% usage). No action needed.",
			},
		},
		{
			name:         "should return correct message for usage under 40",
			usage:        12.33,
			resourceType: "memory",
			container:    "gitserver-0",
			want: scout.Advice{
				Kind: scout.WARNING,
				Msg:  "⚠️  gitserver-0: memory is over-provisioned (12.33% usage). Trim resources.",
			},
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
	advice := []scout.Advice{
		{
			Kind: scout.WARNING,
			Msg:  "Add more CPU",
		},
		{
			Kind: scout.WARNING,
			Msg:  "Add more memory",
		},
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
		{2, advice[0].Msg},
		{3, advice[1].Msg},
	}

	for _, tc := range cases {
		tc := tc
		got := lines[tc.lineNum-1]
		if got != tc.want {
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
