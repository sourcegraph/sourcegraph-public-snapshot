import type { PageLoad } from './$types'

import { isErrorLike } from '$lib/common'
import { queryGitBranchesOverview } from '$lib/loader/repo'
import { asStore } from '$lib/utils'

export const load: PageLoad = ({ parent }) => ({
    branches: asStore(
        parent().then(({ resolvedRevision }) =>
            isErrorLike(resolvedRevision)
                ? null
                : queryGitBranchesOverview({ repo: resolvedRevision.repo.id, first: 10 }).toPromise()
        )
    ),
})
