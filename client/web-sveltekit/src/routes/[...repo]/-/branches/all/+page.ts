import type { PageLoad } from './$types'

import { isErrorLike } from '$lib/common'
import { GitRefType } from '$lib/graphql-operations'
import { queryGitReferences } from '$lib/loader/repo'
import { asStore } from '$lib/utils'

export const load: PageLoad = ({ parent }) => ({
    branches: asStore(
        parent().then(({ resolvedRevision }) =>
            isErrorLike(resolvedRevision)
                ? null
                : queryGitReferences({
                      repo: resolvedRevision.repo.id,
                      type: GitRefType.GIT_BRANCH,
                      first: 20,
                  }).toPromise()
        )
    ),
})
