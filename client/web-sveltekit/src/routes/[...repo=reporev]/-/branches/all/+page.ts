import { getGraphQLClient, infinityQuery, OverwriteRestoreStrategy } from '$lib/graphql'
import { parseRepoRevision } from '$lib/shared'

import type { PageLoad } from './$types'
import { AllBranchesPage_BranchesQuery } from './page.gql'

const PAGE_SIZE = 50

export const load: PageLoad = ({ params, url }) => {
    const client = getGraphQLClient()
    const { repoName } = parseRepoRevision(params.repo)
    const query = url.searchParams.get('query') ?? ''

    return {
        query,
        branchesQuery: infinityQuery({
            client,
            query: AllBranchesPage_BranchesQuery,
            variables: {
                repoName,
                first: PAGE_SIZE,
                withBehindAhead: true,
                query,
            },
            map: result => {
                const branches = result.data?.repository?.branches
                return {
                    nextVariables: branches?.pageInfo.hasNextPage
                        ? { first: branches.nodes.length + PAGE_SIZE }
                        : undefined,
                    data: branches
                        ? {
                              nodes: branches.nodes,
                              totalCount: branches.totalCount,
                          }
                        : undefined,
                    error: result.error,
                }
            },
            createRestoreStrategy: api => new OverwriteRestoreStrategy(api, data => ({ first: data.nodes.length })),
        }),
    }
}
