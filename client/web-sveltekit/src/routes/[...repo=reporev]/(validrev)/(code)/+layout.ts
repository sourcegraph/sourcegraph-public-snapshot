import { dirname } from 'path'

import { browser } from '$app/environment'
import type { Scalars } from '$lib/graphql-types'
import { fetchSidebarFileTree } from '$lib/repo/api/tree'

import type { LayoutLoad } from './$types'
import { GitHistoryQuery, type GitHistory_HistoryConnection } from './layout.gql'

const HISTORY_COMMITS_PER_PAGE = 20

// Signifies the path of the repository root
const REPO_ROOT = '.'

let getRootPath = (_repo: string, path: string) => path

// We keep state in the browser to load the tree entries of the "highest" directory that was visited.
if (browser) {
    const topTreePath: Record<string, string> = {}

    getRootPath = (repo: string, path: string) => {
        const treePath = topTreePath[repo]
        if (treePath && (treePath === REPO_ROOT || path.startsWith(treePath))) {
            return topTreePath[repo]
        }
        return (topTreePath[repo] = path)
    }
}

interface FetchCommitHistoryArgs {
    repo: Scalars['ID']['input']
    revspec: string
    filePath: string
    afterCursor: string | null
}

export const load: LayoutLoad = async ({ parent, params }) => {
    const { resolvedRevision, repoName, graphqlClient } = await parent()
    const parentPath = getRootPath(repoName, params.path ? dirname(params.path) : REPO_ROOT)

    function fetchCommitHistory({
        repo,
        revspec,
        filePath,
        afterCursor,
    }: FetchCommitHistoryArgs): Promise<GitHistory_HistoryConnection | null> {
        return graphqlClient
            .query({
                query: GitHistoryQuery,
                variables: {
                    repo,
                    revspec,
                    filePath,
                    first: HISTORY_COMMITS_PER_PAGE,
                    afterCursor,
                },
            })
            .then(result => {
                if (result.data.node?.__typename !== 'Repository') {
                    throw new Error('Expected repository')
                }
                return result.data.node.commit?.ancestors ?? null
            })
    }

    return {
        parentPath,
        fetchCommitHistory,
        deferred: {
            // Fetches the most recent commits for current blob, tree or repo root
            commitHistory: fetchCommitHistory({
                repo: resolvedRevision.repo.id,
                revspec: resolvedRevision.commitID,
                filePath: params.path ?? '',
                afterCursor: null,
            }),
            fileTree: fetchSidebarFileTree({
                repoID: resolvedRevision.repo.id,
                commitID: resolvedRevision.commitID,
                filePath: parentPath,
            }),
        },
    }
}
