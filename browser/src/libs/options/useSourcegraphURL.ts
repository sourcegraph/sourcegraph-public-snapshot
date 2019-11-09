import { useMemo } from 'react'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { Subject, merge, Observable, of } from 'rxjs'
import { ErrorLike } from '../../../../shared/src/util/errors'
import {
    map,
    sample,
    debounceTime,
    takeUntil,
    switchMap,
    filter,
    mapTo,
    startWith,
    catchError,
    concatMap,
    repeat,
    share,
} from 'rxjs/operators'
import { isDefined } from '../../../../shared/src/util/types'
import { InvalidSourcegraphURLError } from '../../shared/util/context'

/**
 * The connection status associated with a Sourcegraph URL.
 */
export type ConnectionStatus =
    | { type: 'connecting' }
    | { type: 'connected' }
    | { type: 'error'; error: ErrorLike; urlHasPermissions?: boolean }

interface SourcegraphURLAndStatus {
    sourcegraphURL: string
    connectionStatus: ConnectionStatus | undefined
}

const withValidSourcegraphURL = <T>(project: (sourcegraphURL: string) => Observable<T>) => (
    sourcegraphURL: string
): Observable<SourcegraphURLAndStatus | T> => {
    try {
        new URL(sourcegraphURL)
    } catch (err) {
        return of({
            sourcegraphURL,
            connectionStatus: {
                type: 'error' as const,
                error: new InvalidSourcegraphURLError(sourcegraphURL),
            },
        })
    }
    return project(sourcegraphURL)
}

export interface UseSourcegraphURLOptions {
    /**
     * Returns an Observable of sourcegraph URLs persisted to browser extension storage.
     */
    observeSourcegraphURL: () => Observable<string>
    /**
     * Attempt to connect to the Sourcegraph instance hosted at the given URL.
     */
    connectToSourcegraphInstance: (sourcegraphURL: string) => Observable<void>
    /**
     * Persists the given URL to browser extension storage.
     */
    persistSourcegraphURL: (sourcegraphURL: string) => Observable<void>
    /**
     * Checks whether the given URL has permissions.
     */
    urlHasPermissions: (url: string) => Observable<boolean>
}

/**
 * Returns an Observable of {@link SourcegraphURLAndStatus}.
 *
 * Exported for testing, callers should use {@link useSourcegraphURL}.
 */
export const observeSourcegraphURLEdition = ({
    changes,
    submits,
    persistSourcegraphURL,
    connectToSourcegraphInstance,
    observeSourcegraphURL,
    urlHasPermissions,
}: UseSourcegraphURLOptions & {
    changes: Observable<string>
    submits: Observable<void>
}): Observable<SourcegraphURLAndStatus> => {
    // Make sure changes and submits multicast.
    changes = changes.pipe(share())
    submits = submits.pipe(share())
    return merge(
        // Clear the connection status when editing.
        changes.pipe(map(sourcegraphURL => ({ sourcegraphURL, connectionStatus: undefined }))),

        // On submit, check if the URL is valid.
        // If it is, persist it to storage, without emitting a connection status.
        // If it isn't, emit an error status.
        merge(
            // Explicit submit.
            changes.pipe(sample(submits)),
            // Submit after 2s of inactivity.
            changes.pipe(debounceTime(2000), takeUntil(submits), repeat())
        ).pipe(
            concatMap(
                withValidSourcegraphURL(sourcegraphURL =>
                    persistSourcegraphURL(sourcegraphURL).pipe(
                        mapTo(null),
                        catchError(() => [
                            {
                                sourcegraphURL,
                                connectionStatus: {
                                    type: 'error' as const,
                                    error: new Error('Error setting Sourcegraph URL'),
                                },
                            },
                        ])
                    )
                )
            ),
            filter(isDefined)
        ),

        // Emit connection status from persisted sourcegraph URLs.
        observeSourcegraphURL().pipe(
            switchMap(
                // Check whether the URL is valid: while only valid URLs are persisted following user edits,
                // an invalid URL may have been persisted in a prior version of the extension, or through managed storage.
                // If the URL is valid, attempt to connect to the Sourcegraph instance.
                withValidSourcegraphURL(sourcegraphURL =>
                    connectToSourcegraphInstance(sourcegraphURL).pipe(
                        mapTo({ type: 'connected' as const }),
                        startWith({ type: 'connecting' as const }),
                        catchError(error =>
                            urlHasPermissions(sourcegraphURL).pipe(
                                catchError(err => {
                                    console.error('Error checking Sourcegraph URL permissions', err)
                                    return [undefined]
                                }),
                                map(urlHasPermissions => ({
                                    type: 'error' as const,
                                    error,
                                    urlHasPermissions,
                                }))
                            )
                        ),
                        map(connectionStatus => ({ connectionStatus, sourcegraphURL }))
                    )
                )
            )
        )
    )
}

/**
 * A React hook encapsulating a edition, validation and persistence of the browser extension's sourcgraph URL.
 *
 * @returns a tuple of three elements:
 * - A function to be called when the user edits the sourcegraph URL, with the edited URL as argument.
 * - A function to be called when the user requests validation of the edited URL.
 * - An object containing the current sourcegraph URL, and its connection status.
 */
export function useSourcegraphURL({
    observeSourcegraphURL,
    connectToSourcegraphInstance,
    persistSourcegraphURL,
    urlHasPermissions,
}: UseSourcegraphURLOptions): [(url: string) => void, () => void, SourcegraphURLAndStatus | undefined] {
    const changes = useMemo(() => new Subject<string>(), [])
    const submits = useMemo(() => new Subject<void>(), [])
    const sourcegraphURLAndStatus = useObservable<SourcegraphURLAndStatus>(
        useMemo(
            () =>
                observeSourcegraphURLEdition({
                    changes,
                    submits,
                    observeSourcegraphURL,
                    connectToSourcegraphInstance,
                    persistSourcegraphURL,
                    urlHasPermissions,
                }),
            [
                changes,
                connectToSourcegraphInstance,
                observeSourcegraphURL,
                persistSourcegraphURL,
                submits,
                urlHasPermissions,
            ]
        )
    )
    return [changes.next.bind(changes), submits.next.bind(submits), sourcegraphURLAndStatus]
}
