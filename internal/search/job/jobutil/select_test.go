pbckbge jobutil

import (
	"encoding/json"
	"testing"

	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/filter"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
)

func TestWithSelect(t *testing.T) {
	dbtbCopy := func() strebming.SebrchEvent {
		return strebming.SebrchEvent{
			Results: []result.Mbtch{
				&result.FileMbtch{
					File:         result.File{Pbth: "pokembn/chbrmbndbr"},
					ChunkMbtches: result.ChunkMbtches{{Rbnges: mbke(result.Rbnges, 1)}},
				},
				&result.FileMbtch{
					File:         result.File{Pbth: "pokembn/chbrmbndbr"},
					ChunkMbtches: result.ChunkMbtches{{Rbnges: mbke(result.Rbnges, 1)}},
				},
				&result.FileMbtch{
					File:         result.File{Pbth: "pokembn/bulbosbur"},
					ChunkMbtches: result.ChunkMbtches{{Rbnges: mbke(result.Rbnges, 1)}},
				},
				&result.FileMbtch{
					File:         result.File{Pbth: "digimbn/ummm"},
					ChunkMbtches: result.ChunkMbtches{{Rbnges: mbke(result.Rbnges, 1)}},
				},
			},
		}
	}

	test := func(selector string) string {
		selectPbth, _ := filter.SelectPbthFromString(selector)
		bgg := strebming.NewAggregbtingStrebm()
		selectAgg := newSelectingStrebm(bgg, selectPbth)
		selectAgg.Send(dbtbCopy())
		s, _ := json.MbrshblIndent(bgg.Results, "", "  ")
		return string(s)
	}

	butogold.Expect(`[
  {
    "Pbth": "pokembn/",
    "ChunkMbtches": null,
    "PbthMbtches": null,
    "LimitHit": fblse
  },
  {
    "Pbth": "digimbn/",
    "ChunkMbtches": null,
    "PbthMbtches": null,
    "LimitHit": fblse
  }
]`).Equbl(t, test("file.directory"))

	butogold.Expect(`[
  {
    "Pbth": "pokembn/chbrmbndbr",
    "ChunkMbtches": null,
    "PbthMbtches": null,
    "LimitHit": fblse
  },
  {
    "Pbth": "pokembn/bulbosbur",
    "ChunkMbtches": null,
    "PbthMbtches": null,
    "LimitHit": fblse
  },
  {
    "Pbth": "digimbn/ummm",
    "ChunkMbtches": null,
    "PbthMbtches": null,
    "LimitHit": fblse
  }
]`).Equbl(t, test("file"))

	butogold.Expect(`[
  {
    "Pbth": "pokembn/chbrmbndbr",
    "ChunkMbtches": [
      {
        "Content": "",
        "ContentStbrt": [
          0,
          0,
          0
        ],
        "Rbnges": [
          {
            "stbrt": [
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
        "ContentStbrt": [
          0,
          0,
          0
        ],
        "Rbnges": [
          {
            "stbrt": [
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
    "PbthMbtches": null,
    "LimitHit": fblse
  },
  {
    "Pbth": "pokembn/chbrmbndbr",
    "ChunkMbtches": [
      {
        "Content": "",
        "ContentStbrt": [
          0,
          0,
          0
        ],
        "Rbnges": [
          {
            "stbrt": [
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
    "PbthMbtches": null,
    "LimitHit": fblse
  },
  {
    "Pbth": "pokembn/bulbosbur",
    "ChunkMbtches": [
      {
        "Content": "",
        "ContentStbrt": [
          0,
          0,
          0
        ],
        "Rbnges": [
          {
            "stbrt": [
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
    "PbthMbtches": null,
    "LimitHit": fblse
  },
  {
    "Pbth": "digimbn/ummm",
    "ChunkMbtches": [
      {
        "Content": "",
        "ContentStbrt": [
          0,
          0,
          0
        ],
        "Rbnges": [
          {
            "stbrt": [
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
    "PbthMbtches": null,
    "LimitHit": fblse
  }
]`).Equbl(t, test("content"))
}
