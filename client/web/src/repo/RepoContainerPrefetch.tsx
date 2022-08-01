import React, { Suspense, useMemo } from 'react'

import { NEVER, ObservableInput, of } from 'rxjs'
import { catchError, switchMap } from 'rxjs/operators'

import { asError, ErrorLike, repeatUntil } from '@sourcegraph/common'
import { isCloneInProgressErrorLike, isRepoSeeOtherErrorLike } from '@sourcegraph/shared/src/backend/errors'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { useObservable } from '@sourcegraph/wildcard'

import { parseBrowserRepoURL } from '../util/url'

import { fetchRepository, resolveRevision } from './backend'
import { RepoContainerProps } from './RepoContainer'

import { redirectToExternalHost } from '.'

interface RepoContainerPrefetchProps extends RepoContainerProps {}

const RepoContainer = lazyComponent(() => import('./RepoContainer'), 'RepoContainer')

/**
 * For Perf
 */
export const RepoContainerPrefetch: React.FunctionComponent<
    React.PropsWithChildren<RepoContainerPrefetchProps>
> = props => {
    const { repoName, revision } = parseBrowserRepoURL(location.pathname + location.search + location.hash)

    // Fetch repository upon mounting the component.
    const repoOrError = useObservable(
        useMemo(
            () =>
                fetchRepository({ repoName }).pipe(
                    catchError(
                        (error): ObservableInput<ErrorLike> => {
                            const redirect = isRepoSeeOtherErrorLike(error)
                            if (redirect) {
                                redirectToExternalHost(redirect)
                                return NEVER
                            }
                            return of(asError(error))
                        }
                    )
                ),
            [repoName]
        )
    )

    const resolvedRevisionOrError = useObservable(
        useMemo(
            () =>
                of(undefined)
                    .pipe(
                        // Wrap in switchMap so we don't break the observable chain when
                        // catchError returns a new observable, so repeatUntil will
                        // properly resubscribe to the outer observable and re-fetch.
                        switchMap(() =>
                            resolveRevision({ repoName, revision }).pipe(
                                catchError(error => {
                                    if (isCloneInProgressErrorLike(error)) {
                                        return of<ErrorLike>(asError(error))
                                    }
                                    throw error
                                })
                            )
                        )
                    )
                    .pipe(
                        repeatUntil(value => !isCloneInProgressErrorLike(value), { delay: 1000 }),
                        catchError(error => of<ErrorLike>(asError(error)))
                    ),
            [repoName, revision]
        )
    )

    return (
        <Suspense>
            <RepoContainer repoOrError={repoOrError} resolvedRevisionOrError={resolvedRevisionOrError} {...props} />
        </Suspense>
    )
}
