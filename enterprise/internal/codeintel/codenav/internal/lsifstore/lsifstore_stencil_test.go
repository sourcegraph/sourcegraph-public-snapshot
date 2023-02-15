package lsifstore

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestStencil(t *testing.T) {
	testCases := []struct {
		name           string
		uploadID       int
		path           string
		expectedRanges []string
	}{
		{
			name:     "scip",
			uploadID: testSCIPUploadID,
			path:     "template/src/telemetry.ts",
			expectedRanges: []string{
				"0:0-0:0",
				"0:12-0:23",
				"0:29-0:42",
				"10:12-10:19",
				"11:12-11:19",
				"12:12-12:19",
				"12:26-12:29",
				"23:16-23:26",
				"23:36-23:42",
				"23:52-23:59",
				"24:13-24:23",
				"24:26-24:36",
				"25:13-25:20",
				"25:23-25:27",
				"25:28-25:31",
				"26:13-26:19",
				"26:22-26:28",
				"27:13-27:20",
				"27:23-27:30",
				"35:11-35:19",
				"35:20-35:26",
				"35:36-35:40",
				"36:17-36:24",
				"36:25-36:28",
				"36:29-36:35",
				"40:13-40:20",
				"40:21-40:24",
				"40:25-40:31",
				"41:13-41:17",
				"41:18-41:24",
				"41:26-41:30",
				"41:32-41:37",
				"41:38-41:43",
				"41:47-41:54",
				"41:55-41:60",
				"41:61-41:66",
				"48:17-48:21",
				"48:22-48:28",
				"48:38-48:42",
				"48:58-48:65",
				"49:18-49:25",
				"54:18-54:29",
				"54:30-54:38",
				"54:39-54:53",
				"54:88-54:94",
				"55:19-55:23",
				"56:16-56:26",
				"56:33-56:40",
				"57:16-57:26",
				"57:33-57:43",
				"58:16-58:28",
				"58:35-58:41",
				"67:12-67:19",
				"68:15-68:19",
				"68:20-68:23",
				"68:33-68:40",
				"7:13-7:29",
				"8:12-8:22",
				"9:12-9:18",
			},
		},
	}

	store := populateTestStore(t)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ranges, err := store.GetStencil(context.Background(), testCase.uploadID, testCase.path)
			if err != nil {
				t.Fatalf("unexpected error %s", err)
			}

			serializedRanges := make([]string, 0, len(ranges))
			for _, r := range ranges {
				serializedRanges = append(serializedRanges, fmt.Sprintf("%d:%d-%d:%d", r.Start.Line, r.Start.Character, r.End.Line, r.End.Character))
			}
			sort.Strings(serializedRanges)

			if diff := cmp.Diff(testCase.expectedRanges, serializedRanges); diff != "" {
				t.Errorf("unexpected ranges (-want +got):\n%s", diff)
			}
		})
	}
}
