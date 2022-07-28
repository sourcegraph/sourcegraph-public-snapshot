import { gql } from '@sourcegraph/http-client'

const LOCKFILE_INDEX_FIELDS = gql`
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

        updatedAt
        createdAt
    }
`

export const LOCKFILE_INDEXES_LIST = gql`
    ${LOCKFILE_INDEX_FIELDS}

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

export const LOCKFILE_INDEX = gql`
    ${LOCKFILE_INDEX_FIELDS}

    query LockfileIndex($id: ID!) {
        node(id: $id) {
            ... on LockfileIndex {
                ...LockfileIndexFields
            }
        }
    }
`

export const DELETE_LOCKFILE_INDEX = gql`
    mutation DeleteLockfileIndex($id: ID!) {
        deleteLockfileIndex(lockfileIndex: $id) {
            alwaysNil
        }
    }
`
