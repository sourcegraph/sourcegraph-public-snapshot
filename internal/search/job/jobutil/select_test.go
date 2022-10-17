package jobutil

import (
	"encoding/json"
	"testing"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

func TestWithSelect(t *testing.T) {
	dataCopy := func() streaming.SearchEvent {
		return streaming.SearchEvent{
			Results: []result.Match{
				&result.FileMatch{
					File:         result.File{Path: "pokeman/charmandar"},
					ChunkMatches: result.ChunkMatches{{Ranges: make(result.Ranges, 1)}},
				},
				&result.FileMatch{
					File:         result.File{Path: "pokeman/charmandar"},
					ChunkMatches: result.ChunkMatches{{Ranges: make(result.Ranges, 1)}},
				},
				&result.FileMatch{
					File:         result.File{Path: "pokeman/bulbosaur"},
					ChunkMatches: result.ChunkMatches{{Ranges: make(result.Ranges, 1)}},
				},
				&result.FileMatch{
					File:         result.File{Path: "digiman/ummm"},
					ChunkMatches: result.ChunkMatches{{Ranges: make(result.Ranges, 1)}},
				},
			},
		}
	}

	test := func(selector string) string {
		selectPath, _ := filter.SelectPathFromString(selector)
		agg := streaming.NewAggregatingStream()
		selectAgg := newSelectingStream(agg, selectPath)
		selectAgg.Send(dataCopy())
		s, _ := json.MarshalIndent(agg.Results, "", "  ")
		return string(s)
	}

	autogold.Want("dedupe paths for select:file.directory", `[
  {
    "Path": "pokeman/",
    "ChunkMatches": null,
    "PathMatches": null,
    "LimitHit": false
  },
  {
    "Path": "digiman/",
    "ChunkMatches": null,
    "PathMatches": null,
    "LimitHit": false
  }
]`).Equal(t, test("file.directory"))

	autogold.Want("dedupe paths select:file", `[
  {
    "Path": "pokeman/charmandar",
    "ChunkMatches": null,
    "PathMatches": null,
    "LimitHit": false
  },
  {
    "Path": "pokeman/bulbosaur",
    "ChunkMatches": null,
    "PathMatches": null,
    "LimitHit": false
  },
  {
    "Path": "digiman/ummm",
    "ChunkMatches": null,
    "PathMatches": null,
    "LimitHit": false
  }
]`).Equal(t, test("file"))

	autogold.Want("don't dedupe file matches for select:content", `[
  {
    "Path": "pokeman/charmandar",
    "ChunkMatches": [
      {
        "Content": "",
        "ContentStart": [
          0,
          0,
          0
        ],
        "Ranges": [
          {
            "start": [
              0,
              0,
              0
            ],
            "end": [
              0,
              0,
              0
            ]
          }
        ]
      },
      {
        "Content": "",
        "ContentStart": [
          0,
          0,
          0
        ],
        "Ranges": [
          {
            "start": [
              0,
              0,
              0
            ],
            "end": [
              0,
              0,
              0
            ]
          }
        ]
      }
    ],
    "PathMatches": null,
    "LimitHit": false
  },
  {
    "Path": "pokeman/charmandar",
    "ChunkMatches": [
      {
        "Content": "",
        "ContentStart": [
          0,
          0,
          0
        ],
        "Ranges": [
          {
            "start": [
              0,
              0,
              0
            ],
            "end": [
              0,
              0,
              0
            ]
          }
        ]
      }
    ],
    "PathMatches": null,
    "LimitHit": false
  },
  {
    "Path": "pokeman/bulbosaur",
    "ChunkMatches": [
      {
        "Content": "",
        "ContentStart": [
          0,
          0,
          0
        ],
        "Ranges": [
          {
            "start": [
              0,
              0,
              0
            ],
            "end": [
              0,
              0,
              0
            ]
          }
        ]
      }
    ],
    "PathMatches": null,
    "LimitHit": false
  },
  {
    "Path": "digiman/ummm",
    "ChunkMatches": [
      {
        "Content": "",
        "ContentStart": [
          0,
          0,
          0
        ],
        "Ranges": [
          {
            "start": [
              0,
              0,
              0
            ],
            "end": [
              0,
              0,
              0
            ]
          }
        ]
      }
    ],
    "PathMatches": null,
    "LimitHit": false
  }
]`).Equal(t, test("content"))
}
