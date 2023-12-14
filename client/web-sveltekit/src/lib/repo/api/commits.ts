import { query, gql } from '$lib/graphql'
import type {
    RepositoryCommitResult,
    Scalars,
    RepositoryComparisonDiffResult,
    RepositoryComparisonDiffVariables,
    HistoryResult,
    GitHistoryResult,
    GitHistoryVariables,
    CommitDiffResult,
    CommitDiffVariables,
    FileDiffFields,
    RepositoryCommitVariables,
} from '$lib/graphql-operations'

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
            canonicalURL
        }
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
        oldNoNewlineAt
        oldRange {
            startLine
            lines
        }
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

const fileDiffFields = gql`
    fragment FileDiffFields on FileDiff {
        oldPath
        oldFile {
            __typename
            canonicalURL
            binary
            byteSize
        }
        newPath
        newFile {
            __typename
            canonicalURL
            binary
            byteSize
        }
        mostRelevantFile {
            __typename
            canonicalURL
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
            id
            ... on Repository {
                commit(rev: $revspec) {
                    id
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
            id
            ... on Repository {
                id
                commit(rev: $revspec) {
                    __typename # Necessary for error handling to check if commit exists
                    id
                    ...GitCommitFields
                }
            }
        }
    }
    ${gitCommitFragment}
`

interface FetchRepoCommitsArgs {
    repoID: Scalars['ID']['input']
    revision: string
    filePath?: string
    first?: number
    pageInfo?: HistoryResult['pageInfo']
}

export async function fetchRepoCommits({
    repoID,
    revision,
    filePath,
    first = HISTORY_COMMITS_PER_PAGE,
    pageInfo,
}: FetchRepoCommitsArgs): Promise<HistoryResult> {
    const emptyResult: HistoryResult = { nodes: [], pageInfo: { hasNextPage: false, endCursor: null } }

    const data = await query<GitHistoryResult, GitHistoryVariables>(HISTORY_QUERY, {
        repo: repoID,
        revspec: revision,
        filePath: filePath ?? null,
        first,
        afterCursor: pageInfo?.endCursor ?? null,
    })
    if (data.node?.__typename === 'Repository') {
        return data.node.commit?.ancestors ?? emptyResult
    }
    return emptyResult
}

export async function fetchRepoCommit(repoId: string, revision: string): Promise<RepositoryCommitResult> {
    return query<RepositoryCommitResult, RepositoryCommitVariables>(COMMIT_QUERY, {
        repo: repoId,
        revspec: revision,
    })
}

export type RepositoryComparisonDiff = Extract<RepositoryComparisonDiffResult['node'], { __typename?: 'Repository' }>

export async function queryRepositoryComparisonFileDiffs(args: {
    repo: Scalars['ID']['input']
    base: string | null
    head: string | null
    first: number | null
    after: string | null
    paths: string[] | null
}): Promise<RepositoryComparisonDiff['comparison']['fileDiffs']> {
    const data = await query<RepositoryComparisonDiffResult, RepositoryComparisonDiffVariables>(
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
                    id
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
    )

    const repo = data.node
    if (repo === null) {
        throw new Error('Repository not found')
    }
    if (repo.__typename !== 'Repository') {
        throw new Error('Not a repository')
    }
    return repo.comparison.fileDiffs
}

export async function fetchDiff(
    repoID: Scalars['ID']['input'],
    revspec: string,
    paths: string[] = []
): Promise<FileDiffFields[]> {
    const data = await query<CommitDiffResult, CommitDiffVariables>(
        gql`
            query CommitDiff($repoID: ID!, $revspec: String!, $paths: [String!], $first: Int) {
                node(id: $repoID) {
                    ... on Repository {
                        __typename
                        id
                        commit(rev: $revspec) {
                            id
                            diff {
                                fileDiffs(paths: $paths, first: $first) {
                                    nodes {
                                        ...FileDiffFields
                                    }
                                }
                            }
                        }
                    }
                }
            }
            ${fileDiffFields}
        `,
        { repoID, revspec, paths, first: paths.length }
    )
    if (data.node?.__typename !== 'Repository') {
        return []
    }
    return data.node.commit?.diff.fileDiffs.nodes ?? []
}
