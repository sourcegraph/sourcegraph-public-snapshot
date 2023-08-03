import { fetchCommits } from '$lib/repo/api/commits'

import type { PageLoad } from './$types'

export const load: PageLoad = async ({ parent }) => {
    const { resolvedRevision } = await parent()

    return {
        deferred: {
            commits: fetchCommits(resolvedRevision).then(result => result?.nodes ?? []),
        },
    }
}
