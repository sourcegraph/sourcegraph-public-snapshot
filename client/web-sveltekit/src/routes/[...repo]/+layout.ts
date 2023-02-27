import { NEVER, of } from 'rxjs'
import { catchError } from 'rxjs/operators'

import type { LayoutLoad } from './$types'

import { asError, encodeURIPathComponent, type ErrorLike } from '$lib/common'
import { resolveRepoRevision } from '$lib/loader/repo'
import { isCloneInProgressErrorLike, isRepoSeeOtherErrorLike, parseRepoRevision } from '$lib/shared'

export const load: LayoutLoad = ({ params }) => {
    const { repoName, revision } = parseRepoRevision(params.repo)

    // TODO: Consider awaiting the resolved revision here since and use
    // SvelteKit's error handling / error page rendering instead. Verify whether
    // returning a promise instead of an object containing a promises changes
    // load behavior.
    const resolvedRevision = resolveRepoRevision({ repoName, revision })
        .pipe(
            catchError(error => {
                const redirect = isRepoSeeOtherErrorLike(error)

                if (redirect) {
                    // redirectToExternalHost(redirect)
                    return NEVER
                }

                if (isCloneInProgressErrorLike(error)) {
                    return of<ErrorLike>(asError(error))
                }

                throw error
            }),
            catchError(error => of<ErrorLike>(asError(error)))
        )
        .toPromise()

    return {
        repoURL: '/' + encodeURIPathComponent(repoName),
        repoName,
        revision,
        resolvedRevision,
    }
}
