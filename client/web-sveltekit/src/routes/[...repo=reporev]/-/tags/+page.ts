import { OverwriteRestoreStrategy, getGraphQLClient, infinityQuery } from '$lib/graphql'
import { parseRepoRevision } from '$lib/shared'

import type { PageLoad } from './$types'
import { TagsPage_TagsQuery } from './page.gql'

const PAGE_SIZE = 50

export const load: PageLoad = ({ params, url }) => {
    const client = getGraphQLClient()
    const { repoName } = parseRepoRevision(params.repo)
    const query = url.searchParams.get('query') ?? ''

    return {
        query,
        tagsQuery: infinityQuery({
            client,
            query: TagsPage_TagsQuery,
            variables: {
                repoName,
                first: PAGE_SIZE,
                withBehindAhead: false,
                query,
            },
            map: result => {
                const gitRefs = result.data?.repository?.gitRefs
                return {
                    nextVariables: gitRefs?.pageInfo.hasNextPage
                        ? { first: gitRefs.nodes.length + PAGE_SIZE }
                        : undefined,
                    data: gitRefs
                        ? {
                              nodes: gitRefs.nodes,
                              totalCount: gitRefs.totalCount,
                          }
                        : undefined,
                    error: result.error,
                }
            },
            createRestoreStrategy: api => new OverwriteRestoreStrategy(api, data => ({ first: data.nodes.length })),
        }),
    }
}
