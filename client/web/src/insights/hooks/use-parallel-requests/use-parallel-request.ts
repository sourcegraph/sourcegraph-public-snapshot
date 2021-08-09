import { useEffect, useState } from 'react'
import { of, from, Subject, ObservableInput, Observable, asyncScheduler, scheduled } from 'rxjs'
import { mergeMap, map, takeUntil, take, catchError, takeWhile, switchMap, publish, refCount } from 'rxjs/operators'

import { ErrorLike, asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'

interface Request<T> {
    /**
     * Request factory (return promise or subscription like request)
     */
    request: () => ObservableInput<T>

    /**
     * Calls whenever request resolves
     *
     * @param result - fetch result from the request
     */
    onComplete: (result: T | ErrorLike) => void

    /**
     * Cancel request stream in order to prevent execution of ongoing  stream
     */
    cancel: Observable<boolean>
}

export interface FetchResult<T> {
    data: T | undefined
    error: ErrorLike | undefined
    loading: boolean
}

const MAX_PARALLEL_QUERIES = 2

/**
 * Parallel requests Hook factory. Used for better testing approach.
 */
// eslint-disable-next-line @typescript-eslint/explicit-function-return-type,@typescript-eslint/explicit-module-boundary-types
export function createUseParallelRequestsHook<T>({ maxRequests } = { maxRequests: MAX_PARALLEL_QUERIES }) {
    const requests = new Subject<Request<T>>()
    const cancelledRequests = new Set<Request<T>>()

    // Global pipeline query/request manager
    requests
        .pipe(
            // Merge map is used here for the concurrent logic over gql request operations.
            // Only N (be default 2) requests can be run in parallel
            mergeMap(event => {
                const { request, onComplete, cancel } = event

                // Makes this stream async to be able to stop further execution in takeWhile
                // by add event to the cancelledRequests set.
                return scheduled([null], asyncScheduler).pipe(
                    takeWhile(() => {
                        if (cancelledRequests.has(event)) {
                            // Make sure we do not have a memory leak in cancelledRequests set.
                            cancelledRequests.delete(event)

                            return false
                        }

                        return true
                    }),
                    switchMap(() =>
                        from(request()).pipe(
                            // In order to be able to cancel this ongoing stream/request
                            takeUntil(cancel),
                            map(payload => ({ payload, onComplete })),
                            // In order to close observable and free up space for other queued requests
                            // in merge map queue. Consider to move this into consumers request calls
                            take(1),
                            catchError(error =>
                                of({
                                    payload: asError(error),
                                    onComplete,
                                })
                            )
                        )
                    )
                )
            }, maxRequests)
        )
        // eslint-disable-next-line rxjs/no-ignored-subscription
        .subscribe(event => {
            const { payload, onComplete } = event

            onComplete(payload)
        })

    /**
     * Runs your request in parallel with other useParallelRequests request calls.
     *
     * @param request - request factory (observer, promise, subscribable like)
     */
    return <D>(request: () => ObservableInput<D>): FetchResult<D> => {
        const [state, setState] = useState<FetchResult<D>>({
            data: undefined,
            error: undefined,
            loading: true,
        })

        useEffect(() => {
            const cancelStream = new Subject<boolean>()

            setState({ data: undefined, loading: true, error: undefined })

            const event: Request<D> = {
                request,
                // Makes cancel stream a hot observable
                cancel: cancelStream.pipe(publish(), refCount()),
                onComplete: result => {
                    if (isErrorLike(result)) {
                        return setState({ data: undefined, loading: false, error: result })
                    }

                    setState({ data: result, loading: false, error: undefined })
                },
            }

            requests.next((event as unknown) as Request<T>)

            return () => {
                // Cancel scheduled stream
                cancelledRequests.add((event as unknown) as Request<T>)

                // Stop/cancel ongoing/started request stream
                cancelStream.next(true)
            }
        }, [request])

        return state
    }
}

// Export useParallelRequests hook with global request manager for all
// consumers.
export const useParallelRequests = createUseParallelRequestsHook<unknown>()
