import { dirname } from 'path'

import { from } from 'rxjs'

import { getGraphQLClient, infinityQuery } from '$lib/graphql'
import { fetchSidebarFileTree } from '$lib/repo/api/tree'
import { resolveRevision } from '$lib/repo/utils'
import { parseRepoRevision } from '$lib/shared'

import type { LayoutLoad } from './$types'
import { GitHistoryQuery, LastCommitQuery } from './layout.gql'

const HISTORY_COMMITS_PER_PAGE = 20

export const load: LayoutLoad = ({ parent, params }) => {
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
        parentPath,
        fileTree,
        lastCommit: client.query(LastCommitQuery, {
            repoName: repoName,
            revspec: revision,
            filePath: parentPath,
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
    }
}
