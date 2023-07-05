import { map } from 'rxjs/operators'

import { createAggregateError, memoizeObservable } from '$lib/common'
import type { FetchHistoryResult, FetchHistoryVariables } from '$lib/graphql-operations'
import { gql } from '$lib/http-client'
import { requestGraphQL } from '$lib/web'

const FETCH_HISTORY = gql`
    fragment GitCommit on GitCommit {
        url
        oid
        abbreviatedOID
        subject
        author {
            person {
                displayName
                avatarURL
            }
            date
        }
        committer {
            person {
                displayName
                avatarURL
            }
            date
        }
    }

    query FetchHistory($repo: ID!, $revision: String!, $currentPath: String!, $first: Int!) {
        node(id: $repo) {
            ... on Repository {
                __typename
                commit(rev: $revision) {
                    __typename
                    ancestors(first: $first, path: $currentPath) {
                        nodes {
                            ...GitCommit
                        }
                    }
                }
            }
        }
    }
`

interface FetchLastCommitObservableArgs {
    repoID: string
    revision: string
    path: string
}

const fetchLastCommitObservable = memoizeObservable(
    ({ repoID, revision, path }: FetchLastCommitObservableArgs) =>
        requestGraphQL<FetchHistoryResult, FetchHistoryVariables>(FETCH_HISTORY, {
            repo: repoID,
            revision,
            currentPath: path,
            first: 1,
        }),
    ({ repoID, revision, path }) => `${repoID}-${revision}-${path}`
)

export function fetchLastCommit(repoID: string, revision: string, path: string) {
    return fetchLastCommitObservable({ repoID, revision, path })
        .pipe(
            map(({ data, errors }) => {
                if (errors || data?.node?.__typename !== 'Repository' || !data.node.commit?.ancestors.nodes.length) {
                    throw createAggregateError(errors)
                }
                return data.node.commit.ancestors.nodes[0]
            })
        )
        .toPromise()
}

interface FetchHistoryObservableArgs extends FetchLastCommitObservableArgs {
    first: number
}

const fetchHistoryObservable = memoizeObservable(
    ({ repoID, revision, path, first }: FetchHistoryObservableArgs) =>
        requestGraphQL<FetchHistoryResult, FetchHistoryVariables>(FETCH_HISTORY, {
            repo: repoID,
            revision,
            currentPath: path,
            first,
        }),
    ({ repoID, revision, path, first }) => `${repoID}-${revision}-${path}-${first}`
)

export function fetchHistory(repoID: string, revision: string, path: string, first: number) {
    return fetchHistoryObservable({ repoID, revision, path, first })
        .pipe(
            map(({ data, errors }) => {
                if (errors || data?.node?.__typename !== 'Repository' || !data.node.commit?.ancestors.nodes.length) {
                    throw createAggregateError(errors)
                }
                return data.node.commit.ancestors.nodes
            })
        )
        .toPromise()
}
