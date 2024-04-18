import { dirname } from 'path'

import { from } from 'rxjs'

import { getGraphQLClient, infinityQuery, mapOrThrow } from '$lib/graphql'
import { GitRefType } from '$lib/graphql-types'
import { fetchSidebarFileTree } from '$lib/repo/api/tree'
import { resolveRevision } from '$lib/repo/utils'
import { parseRepoRevision } from '$lib/shared'

import type { LayoutLoad } from './$types'
import { GitHistoryQuery, LastCommitQuery, RepositoryGitCommits, RepositoryGitRefs } from './layout.gql'

const HISTORY_COMMITS_PER_PAGE = 20

export const load: LayoutLoad = async ({ parent, params }) => {
    const client = getGraphQLClient()
    const { repoName, revision = '' } = parseRepoRevision(params.repo)
    const parentPath = params.path ? dirname(params.path) : ''
    const resolvedRevision = resolveRevision(parent, revision)

    // Prefetch the sidebar file tree for the parent path.
    // (we don't want to wait for the file tree to execute the query)
    // This also used by the page to find the readme file
    const fileTree = resolvedRevision
        .then(revision =>
            fetchSidebarFileTree({
                repoName,
                revision,
                filePath: parentPath,
            })
        )
        .catch(() => null)

    return {
        fileTree,
        parentPath,
        lastCommit: client.query(LastCommitQuery, {
            repoName,
            revspec: revision,
            filePath: params.path ?? '',
        }),
        // Fetches the most recent commits for current blob, tree or repo root
        commitHistory: infinityQuery({
            client,
            query: GitHistoryQuery,
            variables: from(
                resolvedRevision.then(revspec => ({
                    repoName,
                    revspec,
                    filePath: params.path ?? '',
                    first: HISTORY_COMMITS_PER_PAGE,
                    afterCursor: null as string | null,
                }))
            ),
            nextVariables: previousResult => {
                if (previousResult?.data?.repository?.commit?.ancestors?.pageInfo?.hasNextPage) {
                    return {
                        afterCursor: previousResult.data.repository.commit.ancestors.pageInfo.endCursor,
                    }
                }
                return undefined
            },
            combine: (previousResult, nextResult) => {
                if (!nextResult.data?.repository?.commit) {
                    return nextResult
                }
                const previousNodes = previousResult.data?.repository?.commit?.ancestors?.nodes ?? []
                const nextNodes = nextResult.data.repository?.commit?.ancestors.nodes ?? []
                return {
                    ...nextResult,
                    data: {
                        repository: {
                            ...nextResult.data.repository,
                            commit: {
                                ...nextResult.data.repository.commit,
                                ancestors: {
                                    ...nextResult.data.repository.commit.ancestors,
                                    nodes: [...previousNodes, ...nextNodes],
                                },
                            },
                        },
                    },
                }
            },
        }),

        // Repository pickers queries (branch, tags and commits)
        getRepoBranches: (searchTerm: string) =>
            getGraphQLClient()
                .query(RepositoryGitRefs, {
                    repoName,
                    query: searchTerm,
                    type: GitRefType.GIT_BRANCH,
                })
                .then(
                    mapOrThrow(({ data, error }) => {
                        if (!data?.repository?.gitRefs) {
                            throw new Error(error?.message)
                        }

                        return data.repository.gitRefs
                    })
                ),
        getRepoTags: (searchTerm: string) =>
            getGraphQLClient()
                .query(RepositoryGitRefs, {
                    repoName,
                    query: searchTerm,
                    type: GitRefType.GIT_TAG,
                })
                .then(
                    mapOrThrow(({ data, error }) => {
                        if (!data?.repository?.gitRefs) {
                            throw new Error(error?.message)
                        }

                        return data.repository.gitRefs
                    })
                ),
        getRepoCommits: (searchTerm: string) =>
            parent().then(({ resolvedRevision }) =>
                getGraphQLClient()
                    .query(RepositoryGitCommits, {
                        repoName,
                        query: searchTerm,
                        revision: resolvedRevision.commitID,
                    })
                    .then(
                        mapOrThrow(({ data, error }) => {
                            if (!data?.repository?.commit) {
                                throw new Error(error?.message)
                            }

                            return data.repository.commit.ancestors
                        })
                    )
            ),
    }
}
