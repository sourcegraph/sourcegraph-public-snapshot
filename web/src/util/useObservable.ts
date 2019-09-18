import { useEffect, useState, useMemo } from 'react'
import { Observable, Observer, Subject } from 'rxjs'

/**
 * React hook to get the latest value of an Observable.
 * Will return `undefined` if the Observable didn't emit yet.
 * If the Observable errors, will throw an error that can be caught with `try`/`catch` or with a React error boundary.
 * The Observable is subscribed on the first render and unsubscribed on unmount or whenever `deps` change.
 *
 * @param observable The Observable to subscribe to.
 * @param deps The dependencies that will trigger a resubscribe when they changed. Help on how to decide what to include here:
 *             - If the observable is the **return value of a function** (e.g. `queryFooByID`), you should generally include the **parameters** to that function here.
 *             - If the observable is a **pipeline** that makes use of **closed-over variables**, include those too.
 *             - If the observable is an observable passed from **props** or from a **service**, include the **observable itself** here.
 *             - If none of the above apply, pass an **empty array**.
 * @throws If the Observable pipeline errors.
 */
export function useObservable<T>(observable: Observable<T>, deps: readonly unknown[]): T | undefined {
    const [error, setError] = useState<any>()
    const [currentValue, setCurrentValue] = useState<T>()

    useEffect(() => {
        setCurrentValue(undefined)
        const subscription = observable.subscribe({ next: setCurrentValue, error: setError })
        return () => subscription.unsubscribe()
    }, deps) // eslint-disable-line react-hooks/exhaustive-deps

    if (error) {
        throw error
    }

    return currentValue
}

/**
 * A React hook to handle a React event with an RxJS pipeline.
 *
 * @param transform An RxJS operator pipeline that is passed an Observable of
 * the events triggered on the next function and can transform it to values of
 * which the latest one is returned.
 * @returns A next function to be used in JSX as an event handler, and the latest result of the Observable pipeline.
 * @throws If the Observable pipeline errors.
 */
export function useEventObservable<T, R>(
    transform: (events: Observable<T>) => Observable<R>,
    deps: readonly unknown[]
): [Observer<T>['next'], R | undefined] {
    const events = useMemo(() => new Subject<T>(), [])
    const observable = useMemo(() => events.pipe(transform), deps) // eslint-disable-line react-hooks/exhaustive-deps
    const value = useObservable(observable, deps)
    return [events.next.bind(events), value]
}
