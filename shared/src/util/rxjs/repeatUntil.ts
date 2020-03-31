import { Observable, merge, EMPTY, timer, of, defer } from 'rxjs'
import { share, last, switchMap, switchMapTo } from 'rxjs/operators'

/**
 * Mirrors values from the source observable and resubscribes to the source observable when it completes,
 * unless its last emitted value passes the provided condition. Resubscription is optionally delayed.
 */
export const repeatUntil = <T>(shouldComplete: (lastValue: T) => boolean, delay?: number) => (
    source: Observable<T>
): Observable<T> =>
    defer(() => {
        const sharedSource = source.pipe(share())
        const repeatedSource = source.pipe(repeatUntil(shouldComplete, delay))
        return merge(
            sharedSource,
            sharedSource.pipe(
                last(),
                switchMap(lastValue => (shouldComplete(lastValue) ? EMPTY : delay ? timer(delay) : of(null))),
                switchMapTo(repeatedSource)
            )
        )
    })
