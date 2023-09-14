import { fetchRepoCommits } from '$lib/repo/api/commits'

import type { PageLoad } from './$types'

export const load: PageLoad = async ({ parent }) => {
    const { resolvedRevision } = await parent()

    return {
        deferred: {
            commits: fetchRepoCommits({ repoID: resolvedRevision.repo.id, revision: resolvedRevision.commitID }).then(
                result => result?.nodes ?? []
            ),
        },
    }
}
