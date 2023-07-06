import { isErrorLike } from '$lib/common'
import { queryGitBranchesOverview } from '$lib/loader/repo'

import type { PageLoad } from './$types'

export const load: PageLoad = ({ parent }) => ({
    branches: {
        deferred: parent().then(({ resolvedRevision }) =>
            isErrorLike(resolvedRevision)
                ? null
                : queryGitBranchesOverview({ repo: resolvedRevision.repo.id, first: 10 }).toPromise()
        ),
    },
})
