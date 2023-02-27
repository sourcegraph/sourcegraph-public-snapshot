import { catchError } from 'rxjs/operators'

import type { PageLoad } from './$types'

import { asError, isErrorLike, type ErrorLike } from '$lib/common'
import { fetchTreeEntries } from '$lib/loader/repo'
import { asStore } from '$lib/utils'
import { requestGraphQL } from '$lib/web'

export const load: PageLoad = ({ params, parent }) => ({
    treeEntries: asStore(
        parent().then(({ resolvedRevision, revision, repoName }) =>
            !isErrorLike(resolvedRevision)
                ? fetchTreeEntries({
                      repoName,
                      commitID: resolvedRevision.commitID,
                      revision: revision ?? '',
                      filePath: params.path,
                      first: 2500,
                      requestGraphQL: options => requestGraphQL(options.request, options.variables),
                  })
                      .pipe(catchError((error): [ErrorLike] => [asError(error)]))
                      .toPromise()
                : null
        )
    ),
})
