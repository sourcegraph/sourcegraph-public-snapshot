import { getGraphQLClient, infinityQuery } from '$lib/graphql'
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
            nextVariables: previousResult => {
                if (previousResult?.data?.repository?.branches?.pageInfo?.hasNextPage) {
                    return {
                        first: previousResult.data.repository.branches.nodes.length + PAGE_SIZE,
                    }
                }
                return undefined
            },
            combine: (_previousResult, nextResult) => {
                return nextResult
            },
        }),
    }
}
