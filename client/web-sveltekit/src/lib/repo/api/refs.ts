import { gql, query } from '$lib/graphql'
import {
    GitRefType,
    type GitRefConnectionFields,
    type GitRefFields,
    type RepositoryGitBranchesOverviewResult,
    type RepositoryGitBranchesOverviewVariables,
    type RepositoryGitRefsResult,
    type RepositoryGitRefsVariables,
    type Scalars,
} from '$lib/graphql-operations'

export const gitReferenceFragments = gql`
    fragment GitRefFields on GitRef {
        __typename
        id
        displayName
        name
        abbrevName
        url
        target {
            commit {
                id
                author {
                    ...SignatureFieldsForReferences
                }
                committer {
                    ...SignatureFieldsForReferences
                }
                behindAhead(revspec: "HEAD") @include(if: $withBehindAhead) {
                    behind
                    ahead
                }
            }
        }
    }

    fragment SignatureFieldsForReferences on Signature {
        __typename
        person {
            displayName
            user {
                username
            }
        }
        date
    }
`

export const REPOSITORY_GIT_REFS = gql`
    query RepositoryGitRefs($repo: ID!, $first: Int, $query: String, $type: GitRefType!, $withBehindAhead: Boolean!) {
        node(id: $repo) {
            id
            ... on Repository {
                gitRefs(first: $first, query: $query, type: $type, orderBy: AUTHORED_OR_COMMITTED_AT) {
                    __typename
                    ...GitRefConnectionFields
                }
            }
        }
    }

    fragment GitRefConnectionFields on GitRefConnection {
        nodes {
            __typename
            ...GitRefFields
        }
        totalCount
        pageInfo {
            hasNextPage
        }
    }

    ${gitReferenceFragments}
`

export async function queryGitReferences(args: {
    repo: Scalars['ID']['input']
    first?: number
    query?: string
    type: GitRefType
    withBehindAhead?: boolean
}): Promise<GitRefConnectionFields> {
    const data = await query<RepositoryGitRefsResult, RepositoryGitRefsVariables>(REPOSITORY_GIT_REFS, {
        query: args.query ?? null,
        first: args.first ?? null,
        repo: args.repo,
        type: args.type,
        withBehindAhead:
            args.withBehindAhead !== undefined ? args.withBehindAhead : args.type === GitRefType.GIT_BRANCH,
    })

    if (data?.node?.__typename !== 'Repository' || !data.node.gitRefs) {
        throw new Error('Unable to fetch git information')
    }
    return data.node.gitRefs
}

interface Data {
    defaultBranch: GitRefFields | null
    activeBranches: GitRefFields[]
    hasMoreActiveBranches: boolean
}

export async function queryGitBranchesOverview(args: { repo: Scalars['ID']['input']; first: number }): Promise<Data> {
    const data = await query<RepositoryGitBranchesOverviewResult, RepositoryGitBranchesOverviewVariables>(
        gql`
            query RepositoryGitBranchesOverview($repo: ID!, $first: Int!, $withBehindAhead: Boolean!) {
                node(id: $repo) {
                    id
                    ...RepositoryGitBranchesOverviewRepository
                }
            }

            fragment RepositoryGitBranchesOverviewRepository on Repository {
                __typename
                defaultBranch {
                    ...GitRefFields
                }
                gitRefs(first: $first, type: GIT_BRANCH, orderBy: AUTHORED_OR_COMMITTED_AT) {
                    nodes {
                        ...GitRefFields
                    }
                    pageInfo {
                        hasNextPage
                    }
                }
            }
            ${gitReferenceFragments}
        `,
        { ...args, withBehindAhead: true }
    )
    if (!data?.node) {
        throw new Error('Unable to fetch repository information')
    }
    const repo = data.node
    if (repo.__typename !== 'Repository') {
        throw new Error('Not a GitRef')
    }
    if (!repo.gitRefs.nodes) {
        throw new Error('Unable to fetch repository information')
    }
    return {
        defaultBranch: repo.defaultBranch,
        activeBranches: repo.gitRefs.nodes.filter(
            // Filter out default branch from activeBranches.
            ({ id }) => !repo.defaultBranch || repo.defaultBranch.id !== id
        ),
        hasMoreActiveBranches: repo.gitRefs.pageInfo.hasNextPage,
    }
}
