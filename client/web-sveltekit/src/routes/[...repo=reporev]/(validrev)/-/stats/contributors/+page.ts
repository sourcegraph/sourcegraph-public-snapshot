import { getGraphQLClient } from '$lib/graphql'
import { getPaginationParams } from '$lib/Paginator'
import { parseRepoRevision } from '$lib/shared'

import type { PageLoad } from './$types'
import { ContributorsPage_ContributorsQuery } from './page.gql'

const pageSize = 20

export const load: PageLoad = async ({ url, params }) => {
    const afterDate = url.searchParams.get('after') ?? ''
    const { first, last, before, after } = getPaginationParams(url.searchParams, pageSize)
    const client = await getGraphQLClient()
    const { repoName } = parseRepoRevision(params.repo)

    const contributors = client
        .query({
            query: ContributorsPage_ContributorsQuery,
            variables: {
                afterDate,
                repoName,
                revisionRange: '',
                path: '',
                first,
                last,
                after,
                before,
            },
        })
        .then(result => {
            return result.data.repository?.contributors ?? null
        })
    return {
        after: afterDate,
        contributors,
    }
}
