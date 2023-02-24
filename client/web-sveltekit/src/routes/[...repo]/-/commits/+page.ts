import type { PageLoad } from './$types'

import { fetchCommits } from '$lib/loader/commits'
import { asStore } from '$lib/utils'

export const load: PageLoad = ({ parent }) => ({
    commits: asStore(
        parent()
            .then(({ resolvedRevision }) => fetchCommits(resolvedRevision))
            .then(result => result?.nodes ?? [])
    ),
})
