import { getPaginationParams } from '$lib/Paginator'
import { fetchContributors } from '$lib/repo/api/contributors'

import type { PageLoad } from './$types'

const pageSize = 20

export const load: PageLoad = async ({ url, parent }) => {
    const afterDate = url.searchParams.get('after') ?? ''
    const { first, last, before, after } = getPaginationParams(url.searchParams, pageSize)
    const { resolvedRevision } = await parent()

    const contributors = fetchContributors({
        afterDate,
        repo: resolvedRevision.repo.id,
        revisionRange: '',
        path: '',
        first,
        last,
        after,
        before,
    })
    return {
        after: afterDate,
        deferred: {
            contributors,
        },
    }
}
