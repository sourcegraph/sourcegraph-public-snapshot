import { map } from 'rxjs/operators'

import { fetchHighlight, fetchBlobPlaintext } from '$lib/loader/blob'
import { parseRepoRevision } from '$lib/shared'

import type { PageLoad } from './$types'

export const load: PageLoad = ({ params }) => {
    const { repoName, revision } = parseRepoRevision(params.repo)

    return {
        blob: {
            deferred: fetchBlobPlaintext({
                filePath: params.path,
                repoName,
                revision: revision ?? '',
            }).toPromise(),
        },
        highlights: {
            deferred: fetchHighlight({ filePath: params.path, repoName, revision: revision ?? '' })
                .pipe(map(highlight => highlight?.lsif))
                .toPromise(),
        },
    }
}
