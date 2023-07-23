// We want to limit the number of imported modules as much as possible
/* eslint-disable no-restricted-imports */

import { memoize } from 'lodash'
import type { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import type { Repo, ResolvedRevision } from '@sourcegraph/web/src/repo/backend'

import { browser } from '$app/environment'
import { isErrorLike, type ErrorLike } from '$lib/common'
import type {
    RepositoryCommitResult,
    Scalars,
    RepositoryComparisonDiffResult,
    RepositoryComparisonDiffVariables,
    GitCommitFields,
    HistoryResult,
    GitHistoryResult,
    GitHistoryVariables,
} from '$lib/graphql-operations'
import { dataOrThrowErrors, gql, type GraphQLResult } from '$lib/http-client'
import { requestGraphQL } from '$lib/web'

// Unfortunately it doesn't seem possible to share fragements across package
// boundaries
const gitCommitFragment = gql`
    fragment GitCommitFields on GitCommit {
        id
        oid
        abbreviatedOID
        subject
        body
        author {
            ...SignatureFields
        }
        committer {
            ...SignatureFields
        }
        parents {
            oid
            abbreviatedOID
            url
        }
        url
        canonicalURL
        externalURLs {
            ...ExternalLinkFields
        }
    }

    fragment SignatureFields on Signature {
        person {
            avatarURL
            name
            displayName
            user {
                id
                username
                url
                displayName
            }
        }
        date
    }

    fragment ExternalLinkFields on ExternalLink {
        url
        serviceKind
    }
`

const diffStatFields = gql`
    fragment DiffStatFields on DiffStat {
        __typename
        added
        deleted
    }
`

const fileDiffHunkFields = gql`
    fragment FileDiffHunkFields on FileDiffHunk {
        oldRange {
            startLine
            lines
        }
        oldNoNewlineAt
        newRange {
            startLine
            lines
        }
        section
        highlight(disableTimeout: false) {
            aborted
            lines {
                kind
                html
            }
        }
    }
`

export const fileDiffFields = gql`
    fragment FileDiffFields on FileDiff {
        oldPath
        oldFile {
            __typename
            binary
            byteSize
        }
        newFile {
            __typename
            binary
            byteSize
        }
        newPath
        mostRelevantFile {
            __typename
            url
        }
        hunks {
            ...FileDiffHunkFields
        }
        stat {
            added
            deleted
        }
        internalID
    }

    ${fileDiffHunkFields}
`

const HISTORY_COMMITS_PER_PAGE = 20

const HISTORY_QUERY = gql`
    query GitHistory($repo: ID!, $revspec: String!, $first: Int, $afterCursor: String, $filePath: String) {
        node(id: $repo) {
            __typename
            ... on Repository {
                commit(rev: $revspec) {
                    ancestors(first: $first, path: $filePath, afterCursor: $afterCursor) {
                        ...HistoryResult
                    }
                }
            }
        }
    }

    fragment HistoryResult on GitCommitConnection {
        nodes {
            ...GitCommitFields
        }
        pageInfo {
            hasNextPage
            endCursor
        }
    }

    ${gitCommitFragment}
`

const COMMIT_QUERY = gql`
    query RepositoryCommit($repo: ID!, $revspec: String!) {
        node(id: $repo) {
            __typename
            ... on Repository {
                commit(rev: $revspec) {
                    __typename # Necessary for error handling to check if commit exists
                    ...GitCommitFields
                }
            }
        }
    }
    ${gitCommitFragment}
`

interface FetchRepoCommitsArgs {
    repoID: string
    revision: string
    filePath: string | null
    first?: number
    pageInfo?: HistoryResult['pageInfo']
}

export const fetchRepoCommits = memoize(
    async ({
        repoID,
        revision,
        filePath,
        first = HISTORY_COMMITS_PER_PAGE,
        pageInfo,
    }: FetchRepoCommitsArgs): Promise<HistoryResult> => {
        const emptyResult: HistoryResult = { nodes: [], pageInfo: { hasNextPage: false, endCursor: null } }

        const result = await requestGraphQL<GitHistoryResult, GitHistoryVariables>(HISTORY_QUERY, {
            repo: repoID,
            revspec: revision,
            filePath: filePath ?? null,
            first,
            afterCursor: pageInfo?.endCursor ?? null,
        }).toPromise()
        const data = dataOrThrowErrors(result)
        if (data.node?.__typename === 'Repository') {
            return data.node.commit?.ancestors ?? emptyResult
        }
        return emptyResult
    },
    args => [args.repoID, args.revision, args.filePath, args.first, args.pageInfo?.endCursor].join('-')
)

export function fetchRepoCommit(repoId: string, revision: string): Observable<GraphQLResult<RepositoryCommitResult>> {
    return requestGraphQL(COMMIT_QUERY, { repo: repoId, revspec: revision })
}

export type RepositoryComparisonDiff = Extract<RepositoryComparisonDiffResult['node'], { __typename?: 'Repository' }>

export function queryRepositoryComparisonFileDiffs(args: {
    repo: Scalars['ID']
    base: string | null
    head: string | null
    first: number | null
    after: string | null
    paths: string[] | null
}): Observable<RepositoryComparisonDiff['comparison']['fileDiffs']> {
    return requestGraphQL<RepositoryComparisonDiffResult, RepositoryComparisonDiffVariables>(
        gql`
            query RepositoryComparisonDiff(
                $repo: ID!
                $base: String
                $head: String
                $first: Int
                $after: String
                $paths: [String!]
            ) {
                node(id: $repo) {
                    __typename
                    ... on Repository {
                        comparison(base: $base, head: $head) {
                            fileDiffs(first: $first, after: $after, paths: $paths) {
                                nodes {
                                    ...FileDiffFields
                                }
                                totalCount
                                pageInfo {
                                    endCursor
                                    hasNextPage
                                }
                                diffStat {
                                    ...DiffStatFields
                                }
                            }
                        }
                    }
                }
            }

            ${fileDiffFields}

            ${diffStatFields}
        `,
        args
    ).pipe(
        map(result => {
            const data = dataOrThrowErrors(result)

            const repo = data.node
            if (repo === null) {
                throw new Error('Repository not found')
            }
            if (repo.__typename !== 'Repository') {
                throw new Error('Not a repository')
            }
            return repo.comparison.fileDiffs
        })
    )
}

const clientCache: Map<string, { nodes: GitCommitFields[] }> = new Map()

function getCacheKey(resolvedRevision: ResolvedRevision & Repo, filePath: string | null): string {
    return [resolvedRevision.repo.id, resolvedRevision.commitID ?? '', filePath].join('/')
}

export async function fetchCommits(
    resolvedRevision: (ResolvedRevision & Repo) | ErrorLike,
    filePath: string | null = null,
    force: boolean = false
): Promise<{ nodes: GitCommitFields[] }> {
    if (!isErrorLike(resolvedRevision)) {
        const cacheKey = getCacheKey(resolvedRevision, filePath)
        if (browser && !force) {
            const fromCache = clientCache.get(cacheKey)
            if (fromCache) {
                return fromCache
            }
        }
        const commits = await fetchRepoCommits({
            repoID: resolvedRevision.repo.id,
            revision: resolvedRevision.commitID ?? '',
            filePath,
        })

        if (browser) {
            clientCache.set(cacheKey, commits)
        }
        return commits
    }
    return { nodes: [] }
}
