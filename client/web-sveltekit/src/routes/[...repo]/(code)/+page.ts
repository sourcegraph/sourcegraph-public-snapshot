import { fetchCommits } from '$lib/loader/commits'
import { asStore } from '$lib/utils'

import type { PageLoad } from './$types'

export const load: PageLoad = ({ parent }) => ({
    commits: asStore(
        parent()
            .then(({ resolvedRevision }) => fetchCommits(resolvedRevision, true))
            .then(result => result?.nodes.slice(0, 5) ?? [])
    ),
})
