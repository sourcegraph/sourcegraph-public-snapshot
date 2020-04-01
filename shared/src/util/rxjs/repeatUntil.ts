import { Observable } from 'rxjs'
import { repeatWhen, delay, takeWhile, repeat } from 'rxjs/operators'

/**
 * Mirrors values from the source observable and resubscribes to the source observable when it completes,
 * until an emission matches the provided condition. Resubscription is optionally delayed.
 */
export const repeatUntil = <T>(select: (value: T) => boolean, delayTime?: number) => (
    source: Observable<T>
): Observable<T> =>
    source.pipe(
        delayTime ? repeatWhen(completions => completions.pipe(delay(delayTime))) : repeat(),
        // Inclusive takeWhile so that the first value matching `select()` is emitted.
        takeWhile(value => !select(value), true)
    )
