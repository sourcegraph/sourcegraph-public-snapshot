import { dirname } from 'path'

import { from } from 'rxjs'

import type { LineOrPositionOrRange } from '$lib/common'
import { getGraphQLClient, infinityQuery } from '$lib/graphql'
import { ROOT_PATH, fetchSidebarFileTree } from '$lib/repo/api/tree'
import { resolveRevision } from '$lib/repo/utils'
import { parseRepoRevision } from '$lib/shared'

import type { LayoutLoad } from './$types'
import { GitHistoryQuery, LastCommitQuery, RepoPage_PreciseCodeIntel } from './layout.gql'

const HISTORY_COMMITS_PER_PAGE = 20
const REFERENCES_PER_PAGE = 20

export const load: LayoutLoad = async ({ parent, params }) => {
    const client = getGraphQLClient()
    const { repoName, revision = '' } = parseRepoRevision(params.repo)
    const filePath = params.path ? decodeURIComponent(params.path) : ''
    const parentPath = filePath ? dirname(filePath) : ROOT_PATH
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
        filePath,
        parentPath,
        lastCommit: client.query(LastCommitQuery, {
            repoName,
            revspec: revision,
            filePath,
        }),
        // Fetches the most recent commits for current blob, tree or repo root
        commitHistory: infinityQuery({
            client,
            query: GitHistoryQuery,
            variables: from(
                resolvedRevision.then(revspec => ({
                    repoName,
                    revspec,
                    filePath,
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
                        filePath,
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
    }
}
