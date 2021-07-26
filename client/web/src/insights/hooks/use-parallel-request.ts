import { useEffect, useRef, useState } from 'react';
import { from, Subject, ObservableInput } from 'rxjs';
import { mergeMap, map, takeUntil, take } from 'rxjs/operators'

interface Request<T> {
    request: () => ObservableInput<T>,
    onComplete: (result: any) => void;
    cancel: Subject<any>;
}

const requests = new Subject<Request<unknown>>()

// Global pipeline query/request manager
requests
    .pipe(
        // Merge map is used here for the concurrent logic over gql request operations.
        // Only N (be default 2) requests can be run in parallel
        mergeMap(
            event => {
                const { request, onComplete, cancel } = event;

                return from(request())
                    .pipe(
                        map(payload => ({ payload, request, onComplete, cancel })),
                        take(1),
                        takeUntil(cancel),
                    )
            },
            2)
    )
    // eslint-disable-next-line rxjs/no-ignored-subscription
    .subscribe(event => {
        const { payload, onComplete } = event

        onComplete(payload)
    })

export function useParallelRequests(request: () => any) {
    const [state, setState] = useState<any>()
    const stopReference = useRef(new Subject())

    useEffect(() => {
        requests.next({
            request,
            cancel: stopReference.current,
            onComplete: res => setState(res)
        })

        return () => stopReference.current.next(true)
    }, [request])

    return state;
}
