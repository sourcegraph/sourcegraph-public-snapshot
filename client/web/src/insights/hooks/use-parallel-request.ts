import { useEffect, useState } from 'react';
import { of, from, Subject, ObservableInput, Observable } from 'rxjs';
import {
    mergeMap,
    map,
    takeUntil,
    take,
    catchError,
    delay,
    takeWhile,
    switchMap,
    publish,
    refCount
} from 'rxjs/operators'

import { ErrorLike, asError, isErrorLike } from '@sourcegraph/shared/src/util/errors';

interface Request<T> {
    /**
     * Request factory (return promise or subscription like request)
     */
    request: () => ObservableInput<T>,

    /**
     * Calls whenever request resolves
     *
     * @param result - fetch result from the request
     */
    onComplete: (result: T | ErrorLike ) => void;

    /**
     * Cancel request stream in order to prevent execution of ongoing  stream
     */
    cancel: Observable<boolean>;
}

interface FetchResult<T> {
    data: T | undefined,
    error: ErrorLike | undefined
    loading: boolean
}

const MAX_PARALLEL_QUERIES = 2
const requests = new Subject<Request<unknown>>()
const cancelledRequests = new Set<Request<unknown>>()

// Global pipeline query/request manager
requests
    .pipe(
        // Merge map is used here for the concurrent logic over gql request operations.
        // Only N (be default 2) requests can be run in parallel
        mergeMap(
            event => {
                const { request, onComplete, cancel } = event;

                return of(null).pipe(
                    // Makes this stream async to be able stop further execution in takeWhile
                    // by add event to the cancelledRequests set.
                    delay(0),
                    takeWhile(() => {
                        if (cancelledRequests.has(event)) {

                            // Make sure we do not have a memory leak in cancelledRequests set.
                            cancelledRequests.delete(event)

                            return false
                        }

                        return true
                    }),
                    switchMap(() =>
                        from(request())
                            .pipe(
                                // In order to be able to cancel this ongoing stream/request
                                takeUntil(cancel),
                                map(payload => ({ payload, onComplete })),
                                // In order to close observable and free up space for other queued requests
                                // in merge map queue. Consider to move this into consumers request calls
                                take(1),
                                catchError(error => of({
                                    payload: asError(error),
                                    onComplete
                                }))
                            )
                    )
                )
            },
            MAX_PARALLEL_QUERIES
        )
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
export function useParallelRequests<T>(request: () => ObservableInput<T>): FetchResult<T> {
    const [state, setState] = useState<FetchResult<T>>({
        data: undefined,
        error: undefined,
        loading: true
    })

    useEffect(() => {
        const cancelStream = new Subject<boolean>()

        const event: Request<T>  = {
            request,
            // Makes cancel stream a hot observable
            cancel: cancelStream.pipe(publish(), refCount()),
            onComplete: result => {
                if (isErrorLike(result)) {
                    return setState({ data: undefined, loading: false, error: result })
                }

                setState({ data: result, loading: false, error: undefined })
            }
        }

        requests.next(event as Request<unknown>)

        return () => {
            // Cancel scheduled stream
            cancelledRequests.add(event as Request<unknown>)

            // Stop/cancel ongoing/started request stream
            cancelStream.next(true)
        }
    }, [request])

    return state;
}
