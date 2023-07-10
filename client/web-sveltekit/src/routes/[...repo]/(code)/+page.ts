import { fetchCommits } from '$lib/loader/commits'

import type { PageLoad } from './$types'

export const load: PageLoad = ({ parent }) => ({
    commits: {
        deferred: parent()
            .then(({ resolvedRevision }) => fetchCommits(resolvedRevision, true))
            .then(result => result?.nodes.slice(0, 5) ?? []),
    },
})
