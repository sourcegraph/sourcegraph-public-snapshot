import { useRef, useEffect, useCallback } from 'react'

type CancelTimeout = () => void
type TimeoutSetter = (
    callback: (...args: unknown[]) => void,
    timeout: number | undefined,
    ...args: unknown[]
) => CancelTimeout

/**
 * React hook that returns a TimeoutSetter function that automates timeout management.
 * It clears previously queued timeouts when called and on component unmount.
 *
 * NOTE: Each instance of TimeoutSetter is meant to manage one timeout at a time.
 * Create a new TimeoutSetter instance when you want to set concurrent timeouts
 *
 * @returns A function with the same parameters as `window.setTimeout` that returns
 * a convenience `clearTimeout` function
 */
export function useTimeout(): TimeoutSetter {
    const timeoutIDReference = useRef<number | undefined>()

    // eslint-disable-next-line arrow-body-style
    useEffect(() => {
        return () => clearTimeout(timeoutIDReference.current)
    }, [])

    return useCallback(function timeoutSetter(callback, timeout, ...args) {
        if (timeoutIDReference.current) {
            clearTimeout(timeoutIDReference.current)
        }

        const timeoutID = window.setTimeout(callback, timeout, args)
        timeoutIDReference.current = timeoutID
        return function cancelTimeout() {
            clearTimeout(timeout)
            // don't void `timeoutIDReference` to prevent invocations of
            // old `cancelTimeout` functions from breaking cleanup
        }
    }, [])
}
