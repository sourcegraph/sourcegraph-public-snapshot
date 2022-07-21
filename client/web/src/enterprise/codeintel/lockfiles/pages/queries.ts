import { gql } from '@sourcegraph/http-client'

export const LOCKFILE_INDEXES_LIST = gql`
    fragment LockfileIndexFields on LockfileIndex {
        id
        lockfile
        fidelity

        repository {
            id
            name
            url
        }

        commit {
            url
            abbreviatedOID
            oid
        }
    }

    fragment LockfileIndexConnectionFields on LockfileIndexConnection {
        nodes {
            ...LockfileIndexFields
        }
        totalCount
        pageInfo {
            endCursor
            hasNextPage
        }
    }

    query LockfileIndexes($first: Int, $after: String) {
        lockfileIndexes(first: $first, after: $after) {
            ...LockfileIndexConnectionFields
        }
    }
`
