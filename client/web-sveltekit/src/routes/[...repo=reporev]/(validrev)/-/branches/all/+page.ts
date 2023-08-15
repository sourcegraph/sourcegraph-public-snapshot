import { GitRefType } from '$lib/graphql-operations'
import { queryGitReferences } from '$lib/repo/api/refs'

import type { PageLoad } from './$types'

export const load: PageLoad = async ({ parent }) => {
    const { resolvedRevision } = await parent()
    return {
        deferred: {
            branches: queryGitReferences({
                repo: resolvedRevision.repo.id,
                type: GitRefType.GIT_BRANCH,
                first: 20,
            }),
        },
    }
}
