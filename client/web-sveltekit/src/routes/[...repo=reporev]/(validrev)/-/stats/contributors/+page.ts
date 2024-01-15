import { getPaginationParams } from '$lib/Paginator'

import type { PageLoad } from './$types'
import { PagedRepositoryContributors } from './page.gql'

const pageSize = 20

export const load: PageLoad = async ({ url, parent }) => {
    const afterDate = url.searchParams.get('after') ?? ''
    const { first, last, before, after } = getPaginationParams(url.searchParams, pageSize)
    const { resolvedRevision, graphqlClient } = await parent()

    const contributors = graphqlClient
        .query({
            query: PagedRepositoryContributors,
            variables: {
                afterDate,
                repo: resolvedRevision.repo.id,
                revisionRange: '',
                path: '',
                first,
                last,
                after,
                before,
            },
        })
        .then(result => {
            if (result.data.node?.__typename !== 'Repository') {
                return null
            }
            return result.data.node.contributors
        })
    return {
        after: afterDate,
        deferred: {
            contributors,
        },
    }
}
