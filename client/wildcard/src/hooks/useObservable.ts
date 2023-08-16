import { useLayoutEffect, useEffect, useState, useMemo, useCallback } from 'react'

import { type Observable, type Observer, Subject } from 'rxjs'

import type { ObservableStatus } from '../types'

/**
 * React hook to get the latest value of an Observable.
 * Will return `undefined` if the Observable didn't emit yet.
 * If the Observable errors, will throw an error that can be caught with `try`/`catch` or with a React error boundary.
 * The Observable is subscribed on the first render and unsubscribed on unmount or whenever it changes (wrap it in `useMemo()` to prevent this).
 *
 * @param observable The Observable to subscribe to. If this is the return value of a function, you should use `useMemo()` to make sure it is not resubscribed on every render.
 * @throws If the Observable pipeline errors.
 */
export function useObservable<T>(observable: Observable<T>): T | undefined {
    const [error, setError] = useState<any>()
    const [currentValue, setCurrentValue] = useState<T>()

    // We use a layout effect to avoid UI tearing when the observable is updated because otherwise
    // the page will be rendered with the old value after the first render pass.
    useLayoutEffect(() => {
        setCurrentValue(undefined)
        const subscription = observable.subscribe({ next: setCurrentValue, error: setError })
        return () => subscription.unsubscribe()
    }, [observable])

    if (error) {
        throw error
    }

    return currentValue
}

/**
 * React hook to get the latest value of an Observable.
 *
 * @description This is a slightly modified version of `useObservable` that returns a `LoadStatus` instead of `undefined` when the Observable hasn't emitted yet.
 * @param observable The Observable to subscribe to.
 * @returns [T | undefined, undefined, Error | undefined]
 */
export function useObservableWithStatus<T>(observable: Observable<T>): [T | undefined, ObservableStatus, any] {
    const [error, setError] = useState<any>()
    const [currentValue, setCurrentValue] = useState<T>()
    const [status, setStatus] = useState<ObservableStatus>('initial')

    const handleNext = useCallback((value: T) => {
        setCurrentValue(value)
        setStatus('next')
    }, [])

    const handleError = useCallback((error: unknown) => {
        setError(error)
        setStatus('error')
    }, [])

    const handleComplete = useCallback(() => {
        setStatus('completed')
    }, [])

    useEffect(() => {
        setCurrentValue(undefined)
        const subscription = observable.subscribe({ next: handleNext, error: handleError, complete: handleComplete })
        return () => {
            setStatus('initial')
            subscription.unsubscribe()
        }
    }, [handleComplete, handleError, handleNext, observable])

    return [currentValue, status, error]
}

/**
 * A React hook to handle a React event with an RxJS pipeline.
 *
 * @template T The event type.
 * @template R The result type.
 * @param transform An RxJS operator pipeline that is passed an Observable of
 * the events triggered on the next function and can transform it to values of
 * which the latest one is returned. **You need to wrap this callback with `useCallback()`**.
 * @returns A next function to be used in JSX as an event handler, and the latest result of the Observable pipeline.
 * @throws If the Observable pipeline errors.
 */
export function useEventObservable<T, R>(
    transform: (events: Observable<T>) => Observable<R>
): [Observer<T>['next'], R | undefined]
export function useEventObservable<R>(
    transform: (events: Observable<void>) => Observable<R>
): [() => void, R | undefined]
export function useEventObservable<T, R>(
    transform: (events: Observable<T>) => Observable<R>
): [Observer<T>['next'], R | undefined] {
    const events = useMemo(() => new Subject<T>(), [])
    const observable = useMemo(() => events.pipe(transform), [events, transform])
    const nextEvent = useMemo(() => events.next.bind(events), [events])
    const value = useObservable(observable)
    return [nextEvent, value]
}
