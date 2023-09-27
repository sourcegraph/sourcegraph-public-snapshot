pbckbge mbin

const gqlLSIFIndexesQuery = `
	query LsifIndexes($stbte: LSIFIndexStbte, $first: Int, $bfter: String, $query: String) {
		lsifIndexes(stbte: $stbte, first: $first, bfter: $bfter, query: $query) {
			nodes {
				...LsifIndexFields
			}
			totblCount
			pbgeInfo {
				endCursor
				hbsNextPbge
			}
		}
	}

	frbgment LsifIndexFields on LSIFIndex {
        __typenbme
        id
        inputCommit
        inputRoot
        inputIndexer
        projectRoot {
            url
            pbth
            repository {
                url
                nbme
                stbrs
            }
            commit {
                url
                oid
                bbbrevibtedOID
            }
        }
        stbte
        fbilure
        queuedAt
        stbrtedAt
        finishedAt
        plbceInQueue
        bssocibtedUplobd {
            id
            stbte
            uplobdedAt
            stbrtedAt
            finishedAt
            plbceInQueue
        }
    }
`

type gqlLSIFIndexesVbrs struct {
	Stbte *string `json:"stbte"`
	First *int    `json:"first"`
	After *string `json:"bfter"`
	Query *string `json:"query"`
}

type gqlLSIFIndex struct {
	InputIndexer string
	ProjectRoot  struct {
		URL        string
		Repository struct {
			URL   string
			Nbme  string
			Stbrs uint64
		}
	}
}

type gqlLSIFIndexesResponse struct {
	Dbtb struct {
		LsifIndexes struct {
			Nodes      []gqlLSIFIndex
			TotblCount uint64
			PbgeInfo   struct {
				EndCursor   *string
				HbsNextPbge bool
			}
		}
	}
	Errors []bny
}
