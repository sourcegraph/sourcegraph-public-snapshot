// We want to limit the number of imported modules as much as possible
/* eslint-disable no-restricted-imports */

import type { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { createAggregateError, memoizeObservable } from '$lib/common'
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
import { gql } from '$lib/http-client'
import { requestGraphQL } from '$lib/web'

export { resolveRepoRevision } from '@sourcegraph/web/src/repo/backend'
export { fetchTreeEntries } from '@sourcegraph/shared/src/backend/repo'

// Copies of non-reusable queries from the main app
// TODO: Refactor queries to make them reusable

export const CONTRIBUTORS_QUERY = gql`
    query PagedRepositoryContributors(
        $repo: ID!
        $first: Int
        $last: Int
        $after: String
        $before: String
        $revisionRange: String
        $afterDate: String
        $path: String
    ) {
        node(id: $repo) {
            ... on Repository {
                __typename
                contributors(
                    first: $first
                    last: $last
                    before: $before
                    after: $after
                    revisionRange: $revisionRange
                    afterDate: $afterDate
                    path: $path
                ) {
                    ...PagedRepositoryContributorConnectionFields
                }
            }
        }
    }

    fragment PagedRepositoryContributorConnectionFields on RepositoryContributorConnection {
        totalCount
        pageInfo {
            hasNextPage
            hasPreviousPage
            endCursor
            startCursor
        }
        nodes {
            ...PagedRepositoryContributorNodeFields
        }
    }

    fragment PagedRepositoryContributorNodeFields on RepositoryContributor {
        __typename
        person {
            name
            displayName
            email
            avatarURL
            user {
                username
                url
                displayName
                avatarURL
            }
        }
        count
        commits(first: 1) {
            nodes {
                oid
                abbreviatedOID
                url
                subject
                author {
                    date
                }
            }
        }
    }
`

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
            __typename
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

export const queryGitReferences = memoizeObservable(
    (args: {
        repo: Scalars['ID']
        first?: number
        query?: string
        type: GitRefType
        withBehindAhead?: boolean
    }): Observable<GitRefConnectionFields> =>
        requestGraphQL<RepositoryGitRefsResult, RepositoryGitRefsVariables>(REPOSITORY_GIT_REFS, {
            query: args.query ?? null,
            first: args.first ?? null,
            repo: args.repo,
            type: args.type,
            withBehindAhead:
                args.withBehindAhead !== undefined ? args.withBehindAhead : args.type === GitRefType.GIT_BRANCH,
        }).pipe(
            map(({ data, errors }) => {
                if (data?.node?.__typename !== 'Repository' || !data.node.gitRefs) {
                    throw createAggregateError(errors)
                }
                return data.node.gitRefs
            })
        ),
    args => `${args.repo}:${String(args.first)}:${String(args.query)}:${args.type}`
)

interface Data {
    defaultBranch: GitRefFields | null
    activeBranches: GitRefFields[]
    hasMoreActiveBranches: boolean
}

export const queryGitBranchesOverview = memoizeObservable(
    (args: { repo: Scalars['ID']; first: number }): Observable<Data> =>
        requestGraphQL<RepositoryGitBranchesOverviewResult, RepositoryGitBranchesOverviewVariables>(
            gql`
                query RepositoryGitBranchesOverview($repo: ID!, $first: Int!, $withBehindAhead: Boolean!) {
                    node(id: $repo) {
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
        ).pipe(
            map(({ data, errors }) => {
                if (!data?.node) {
                    throw createAggregateError(errors)
                }
                const repo = data.node
                if (repo.__typename !== 'Repository') {
                    throw new Error('Not a GitRef')
                }
                if (!repo.gitRefs.nodes) {
                    throw createAggregateError(errors)
                }
                return {
                    defaultBranch: repo.defaultBranch,
                    activeBranches: repo.gitRefs.nodes.filter(
                        // Filter out default branch from activeBranches.
                        ({ id }) => !repo.defaultBranch || repo.defaultBranch.id !== id
                    ),
                    hasMoreActiveBranches: repo.gitRefs.pageInfo.hasNextPage,
                }
            })
        ),
    args => `${args.repo}:${args.first}`
)
