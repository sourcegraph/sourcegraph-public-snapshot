import { dirname } from 'path'

import { catchError } from 'rxjs/operators'

import type { LayoutLoad } from './$types'

import { asError, isErrorLike, type ErrorLike } from '$lib/common'
import { fetchTreeEntries } from '$lib/loader/repo'
import { asStore } from '$lib/utils'
import { requestGraphQL } from '$lib/web'

export const load: LayoutLoad = ({ parent, params }) => ({
    treeEntries: asStore(
        parent().then(({ resolvedRevision, repoName, revision }) =>
            !isErrorLike(resolvedRevision)
                ? fetchTreeEntries({
                      repoName,
                      commitID: resolvedRevision.commitID,
                      revision: revision ?? '',
                      filePath: params.path ? dirname(params.path) : '.',
                      first: 2500,
                      requestGraphQL: options => requestGraphQL(options.request, options.variables),
                  })
                      .pipe(catchError((error): [ErrorLike] => [asError(error)]))
                      .toPromise()
                : null
        )
    ),
})
