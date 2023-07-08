import { fetchCommits } from '$lib/loader/commits'

import type { PageLoad } from './$types'

export const load: PageLoad = ({ parent }) => ({
    commits: {
        deferred: parent()
            .then(({ resolvedRevision }) => fetchCommits(resolvedRevision))
            .then(result => result?.nodes ?? []),
    },
})
