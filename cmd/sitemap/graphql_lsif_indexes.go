package main

const gqlLSIFIndexesQuery = `
	query LsifIndexes($state: LSIFIndexState, $first: Int, $after: String, $query: String) {
		lsifIndexes(state: $state, first: $first, after: $after, query: $query) {
			nodes {
				...LsifIndexFields
			}
			totalCount
			pageInfo {
				endCursor
				hasNextPage
			}
		}
	}

	fragment LsifIndexFields on LSIFIndex {
        __typename
        id
        inputCommit
        inputRoot
        inputIndexer
        projectRoot {
            url
            path
            repository {
                url
                name
                stars
            }
            commit {
                url
                oid
                abbreviatedOID
            }
        }
        state
        failure
        queuedAt
        startedAt
        finishedAt
        placeInQueue
        associatedUpload {
            id
            state
            uploadedAt
            startedAt
            finishedAt
            placeInQueue
        }
    }
`

type gqlLSIFIndexesVars struct {
	State *string `json:"state"`
	First *int    `json:"first"`
	After *string `json:"after"`
	Query *string `json:"query"`
}

type gqlLSIFIndex struct {
	InputIndexer string
	ProjectRoot  struct {
		URL        string
		Repository struct {
			URL   string
			Name  string
			Stars uint64
		}
	}
}

type gqlLSIFIndexesResponse struct {
	Data struct {
		LsifIndexes struct {
			Nodes      []gqlLSIFIndex
			TotalCount uint64
			PageInfo   struct {
				EndCursor   *string
				HasNextPage bool
			}
		}
	}
	Errors []any
}
