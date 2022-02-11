import { useState, useEffect, useRef, useCallback } from 'react'

interface UseIntervalControls {
    hasInterval: boolean
    startExecution: () => void
    stopExecution: () => void
}

/**
 * Custom hook for controlling the repeated execution of a function on an interval.
 *
 * @param callback The function to call on the execution interval
 * @param interval The interval at which the function should be re-invoked, in ms
 */
export const useInterval = (
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    callback: () => any,
    interval: number
): UseIntervalControls => {
    const [internalInterval, setInternalInterval] = useState(interval)
    // The function will continue to re-execute so long as there is a positive interval value.
    const hasInterval = internalInterval > 0
    const isMounted = useRef(true)

    const savedCallback = useRef(callback)
    // Remember the latest callback.
    useEffect(() => {
        savedCallback.current = callback
    }, [callback])

    // Set up the interval to actually run the callback.
    useEffect(() => {
        if (isMounted.current && hasInterval) {
            const timeout = setInterval(() => {
                savedCallback.current()
            }, interval)
            return () => clearInterval(timeout)
        }
        return
    }, [interval, hasInterval])

    // Callback to stop the execution of the function.
    const stopExecution = useCallback(() => {
        if (isMounted.current && hasInterval) {
            setInternalInterval(-1)
        }
    }, [hasInterval])

    // Callback to start the execution of the function.
    const startExecution = useCallback(() => {
        if (isMounted.current && !hasInterval) {
            setInternalInterval(interval)
        }
    }, [interval, hasInterval])

    useEffect(() => {
        isMounted.current = true
        startExecution()
        return () => {
            isMounted.current = false
            stopExecution()
        }
    }, [startExecution, stopExecution])

    return { hasInterval, startExecution, stopExecution }
}
