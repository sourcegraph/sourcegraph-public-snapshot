import { dirname } from 'path'

import { readable, derived, type Readable } from 'svelte/store'

import { SourcegraphURL } from '@sourcegraph/common'

import { CodyContextFiltersSchema, getFiltersFromCodyContextFilters } from '$lib/cody/config'
import { getGraphQLClient, infinityQuery, type GraphQLClient, IncrementalRestoreStrategy } from '$lib/graphql'
import { ROOT_PATH, fetchSidebarFileTree } from '$lib/repo/api/tree'
import { resolveRevision } from '$lib/repo/utils'
import { parseRepoRevision } from '$lib/shared'

import type { LayoutLoad } from './$types'
import { CodyContextFiltersQuery, GitHistoryQuery, LastCommitQuery } from './layout.gql'

const HISTORY_COMMITS_PER_PAGE = 20

export const load: LayoutLoad = async ({ parent, params, url }) => {
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
        lineOrPosition: SourcegraphURL.from(url).lineRange,
        isCodyAvailable: createCodyAvailableStore(client, repoName),
        lastCommit: resolvedRevision
            .then(revspec =>
                client.query(LastCommitQuery, {
                    repoName,
                    revspec,
                    filePath,
                })
            )
            .then(result => result.data?.repository?.lastCommit?.ancestors.nodes[0]),
        // Fetches the most recent commits for current blob, tree or repo root
        commitHistory: infinityQuery({
            client,
            query: GitHistoryQuery,
            variables: resolvedRevision.then(revspec => ({
                repoName,
                revspec,
                filePath,
                first: HISTORY_COMMITS_PER_PAGE,
                afterCursor: null as string | null,
            })),
            map: result => {
                const anestors = result.data?.repository?.commit?.ancestors
                return {
                    nextVariables: anestors?.pageInfo.hasNextPage
                        ? { afterCursor: anestors.pageInfo.endCursor }
                        : undefined,
                    data: anestors?.nodes,
                    error: result.error,
                }
            },
            merge: (previous, next) => (previous ?? []).concat(next ?? []),
            createRestoreStrategy: api =>
                new IncrementalRestoreStrategy(
                    api,
                    n => n.length,
                    n => ({ first: n.length })
                ),
        }),
    }
}

/**
 * Returns a store that indicates whether Cody is available for the current user and repository.
 * If cody is not enabled on the instance or for the current user, the store will always return false.
 * If this is sourcegraph.com, the store will always return true.
 * Otherwise we'll check the site configuration to see if Cody is disabled for the current repository.
 * Initially the store will return false until the site configuration is loaded. If there is an
 * error loading the site configuration or processing it, the store will return false.
 */
function createCodyAvailableStore(client: GraphQLClient, repoName: string): Readable<boolean> {
    if (!window.context.codyEnabledOnInstance || !window.context.codyEnabledForCurrentUser) {
        return readable(false)
    }

    // Cody is always enabled on sourcegraph.com
    if (window.context.sourcegraphDotComMode) {
        return readable(true)
    }

    // On enterprise instances, we check whether the site config disables
    // cody for specific repos.

    const queryResult = readable(
        // First check the cache to see if Cody is disabled for the current repo.
        client.readQuery(CodyContextFiltersQuery, {}),
        set => {
            // Then update the store with the latest data.
            client.query(CodyContextFiltersQuery, {}, { requestPolicy: 'network-only' }).then(set)
        }
    )

    // NOTE: The way this is implemented won't trigger a GraphQL on data prefetching. This is intentional
    // (for now) because we don't want to refetch the data for every data preload.
    return derived(queryResult, ($codyContextFilters, set) => {
        if (!$codyContextFilters || $codyContextFilters.error) {
            // Cody context filters are not available, disable Cody
            set(false)
            return
        }
        const filters = $codyContextFilters.data?.site.codyContextFilters.raw
        if (!filters) {
            // Cody context filters are not defined, enable Cody
            set(true)
            return
        }

        CodyContextFiltersSchema.safeParseAsync(filters).then(result => {
            if (!result.success) {
                // codyContextFilters cannot be parsed properly, disable Cody
                // TODO: log error with sentry
                set(false)
                return
            }
            if (result.data) {
                set(getFiltersFromCodyContextFilters(result.data)(repoName))
            }
        })
    })
}
