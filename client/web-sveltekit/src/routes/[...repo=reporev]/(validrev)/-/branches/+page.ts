import { queryGitBranchesOverview } from '$lib/repo/api/refs'

import type { PageLoad } from './$types'

export const load: PageLoad = async ({ parent }) => {
    const { resolvedRevision } = await parent()
    return {
        deferred: {
            branches: queryGitBranchesOverview({ repo: resolvedRevision.repo.id, first: 10 }),
        },
    }
}
