package main

const gqlLSIFIndexesQuery = `
	query LsifIndexes($state: LSIFIndexState, $first: Int, $after: String, $query: String) {
		lsifIndexes(query: $query, state: $state, first: $first, after: $after) {
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
            }
            commit {
                url
                oid
                abbreviatedOID
            }
        }
        steps {
            ...LsifIndexStepsFields
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
    fragment LsifIndexStepsFields on IndexSteps {
        setup {
            ...ExecutionLogEntryFields
        }
        preIndex {
            root
            image
            commands
            logEntry {
                ...ExecutionLogEntryFields
            }
        }
        index {
            indexerArgs
            outfile
            logEntry {
                ...ExecutionLogEntryFields
            }
        }
        upload {
            ...ExecutionLogEntryFields
        }
        teardown {
            ...ExecutionLogEntryFields
        }
    }
    fragment ExecutionLogEntryFields on ExecutionLogEntry {
        key
        command
        startTime
        exitCode
        out
        durationMilliseconds
    }
`

type gqlLSIFIndexesVars struct {
	State *string `json:"state"`
	First *int    `json:"first"`
	After *string `json:"after"`
	Query *string `json:"query"`
}

type gqlLSIFIndexesResponse struct {
	Data struct {
		LsifIndexes struct {
			Nodes []interface{}
		}
		TotalCount uint64
		PageInfo   struct {
			EndCursor   string
			HasNextPage bool
		}
	}
	Errors []interface{}
}
