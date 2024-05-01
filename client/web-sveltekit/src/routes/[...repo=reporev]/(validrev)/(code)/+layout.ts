import { dirname } from 'path'

import { from } from 'rxjs'

import type { LineOrPositionOrRange } from '$lib/common'
import { getGraphQLClient, infinityQuery, mapOrThrow } from '$lib/graphql'
import { GitRefType } from '$lib/graphql-types'
import { fetchSidebarFileTree } from '$lib/repo/api/tree'
import { resolveRevision } from '$lib/repo/utils'
import { parseRepoRevision } from '$lib/shared'

import type { LayoutLoad } from './$types'
import {
    GitHistoryQuery,
    LastCommitQuery,
    RepositoryGitCommits,
    RepositoryGitRefs,
    RepoPage_PreciseCodeIntel,
} from './layout.gql'

const HISTORY_COMMITS_PER_PAGE = 20
const REFERENCES_PER_PAGE = 20

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

        // We are not extracting the selected position from the URL because that creates a dependency
        // on the full URL, which causes this loader to be re-executed on every URL change.
        getReferenceStore: (lineOrPosition: LineOrPositionOrRange & { line: number }) =>
            infinityQuery({
                client,
                query: RepoPage_PreciseCodeIntel,
                variables: from(
                    resolvedRevision.then(revspec => ({
                        repoName,
                        revspec,
                        filePath: params.path ?? '',
                        first: REFERENCES_PER_PAGE,
                        // Line and character are 1-indexed, but the API expects 0-indexed
                        line: lineOrPosition.line - 1,
                        character: lineOrPosition.character! - 1,
                        afterCursor: null as string | null,
                    }))
                ),
                nextVariables: previousResult => {
                    if (previousResult?.data?.repository?.commit?.blob?.lsif?.references.pageInfo.hasNextPage) {
                        return {
                            afterCursor: previousResult.data.repository.commit.blob.lsif.references.pageInfo.endCursor,
                        }
                    }
                    return undefined
                },
                combine: (previousResult, nextResult) => {
                    if (!nextResult.data?.repository?.commit?.blob?.lsif) {
                        return nextResult
                    }

                    const previousNodes = previousResult.data?.repository?.commit?.blob?.lsif?.references?.nodes ?? []
                    const nextNodes = nextResult.data?.repository?.commit?.blob?.lsif?.references?.nodes ?? []

                    return {
                        ...nextResult,
                        data: {
                            repository: {
                                ...nextResult.data.repository,
                                commit: {
                                    ...nextResult.data.repository.commit,
                                    blob: {
                                        ...nextResult.data.repository.commit.blob,
                                        lsif: {
                                            ...nextResult.data.repository.commit.blob.lsif,
                                            references: {
                                                ...nextResult.data.repository.commit.blob.lsif.references,
                                                nodes: [...previousNodes, ...nextNodes],
                                            },
                                        },
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
                        mapOrThrow(({ data }) => {
                            let nodes = data?.repository?.ancestorCommits?.ancestors.nodes ?? []

                            // If we got a match for the OID, add it to the list if it doesn't already exist.
                            // We double check that the OID contains the search term because we cannot search
                            // specifically by OID, and an empty string resolves to HEAD.
                            const commitByHash = data?.repository?.commitByHash
                            if (
                                commitByHash &&
                                commitByHash.oid.includes(searchTerm) &&
                                !nodes.some(node => node.oid === commitByHash.oid)
                            ) {
                                nodes = [commitByHash, ...nodes]
                            }
                            return { nodes }
                        })
                    )
            ),
    }
}
