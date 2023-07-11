import { isErrorLike } from '$lib/common'
import { GitRefType } from '$lib/graphql-operations'
import { queryGitReferences } from '$lib/loader/repo'

import type { PageLoad } from './$types'

export const load: PageLoad = ({ parent }) => ({
    branches: {
        deferred: parent().then(({ resolvedRevision }) =>
            isErrorLike(resolvedRevision)
                ? null
                : queryGitReferences({
                      repo: resolvedRevision.repo.id,
                      type: GitRefType.GIT_BRANCH,
                      first: 20,
                  }).toPromise()
        ),
    },
})
