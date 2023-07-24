import { fetchCommits } from '$lib/repo/api/commits'

import type { PageLoad } from './$types'

export const load: PageLoad = ({ parent }) => ({
    deferred: {
        commits: parent()
            .then(({ resolvedRevision }) => fetchCommits(resolvedRevision))
            .then(result => result?.nodes ?? []),
    },
})
