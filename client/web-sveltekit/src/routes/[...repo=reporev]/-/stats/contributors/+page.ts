import { getGraphQLClient, mapOrThrow } from '$lib/graphql'
import { getPaginationParams } from '$lib/Paginator'
import { parseRepoRevision } from '$lib/shared'

import type { PageLoad } from './$types'
import { ContributorsPage_ContributorsQuery } from './page.gql'

const pageSize = 20

export const load: PageLoad = ({ url, params }) => {
    const afterDate = url.searchParams.get('after') ?? null
    const { first, last, before, after } = getPaginationParams(url.searchParams, pageSize)
    const client = getGraphQLClient()
    const { repoName } = parseRepoRevision(params.repo)

    const contributors = client
        .query(ContributorsPage_ContributorsQuery, {
            afterDate,
            repoName,
            revisionRange: '',
            path: '',
            first,
            last,
            after,
            before,
        })
        .then(mapOrThrow(result => result.data?.repository?.contributors ?? null))
    return {
        after: afterDate,
        contributors,
    }
}
