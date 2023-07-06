import { isErrorLike } from '$lib/common'
import { getPaginationParams } from '$lib/Paginator.svelte'
import { fetchContributors } from '$lib/repo/api/contributors'

import type { PageLoad } from './$types'

const pageSize = 20

export const load: PageLoad = ({ url, parent }) => {
    const afterDate = url.searchParams.get('after') ?? ''
    const { first, last, before, after } = getPaginationParams(url.searchParams, pageSize)

    const contributors = parent().then(({ resolvedRevision, graphqlClient }) =>
        !isErrorLike(resolvedRevision)
            ? fetchContributors(graphqlClient, {
                  afterDate,
                  repo: resolvedRevision.repo.id,
                  revisionRange: '',
                  path: '',
                  first,
                  last,
                  after,
                  before,
              })
            : null
    )
    return {
        after: afterDate,
        deferred: {
            contributors,
        },
    }
}
