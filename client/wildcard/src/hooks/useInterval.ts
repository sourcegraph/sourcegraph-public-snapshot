import { useCallback, useEffect, useRef, useState } from 'react'

import { useIsMounted } from '.'

interface UseIntervalControls {
    hasInterval: boolean
    startExecution: () => void
    stopExecution: () => void
}

/**
 * Custom hook for controlling the repeated execution of a function on an interval.
 * @param callback The function to call on the execution interval
 * @param interval The interval at which the function should be re-invoked, in ms
 */
export const useInterval = (callback: () => any, interval: number): UseIntervalControls => {
    const [internalInterval, setInternalInterval] = useState(interval)
    // The function will continue to re-execute so long as there is a positive interval value.
    const hasInterval = internalInterval > 0
    const isMounted = useIsMounted()

    const savedCallback = useRef(callback)
    // Remember the latest callback.
    useEffect(() => {
        savedCallback.current = callback
    }, [callback])

    // Update the interval if it changes
    useEffect(() => {
        setInternalInterval(interval)
    }, [interval])

    // Set up the interval to actually run the callback.
    useEffect(() => {
        if (isMounted() && hasInterval) {
            const timeout = setInterval(() => {
                savedCallback.current()
            }, interval)
            return () => clearInterval(timeout)
        }
        return
    }, [interval, hasInterval, isMounted])

    // Callback to stop the execution of the function.
    const stopExecution = useCallback(() => {
        if (hasInterval) {
            setInternalInterval(-1)
        }
    }, [hasInterval])

    // Callback to start the execution of the function.
    const startExecution = useCallback(() => {
        if (!hasInterval) {
            setInternalInterval(interval)
        }
    }, [interval, hasInterval])

    return { hasInterval, startExecution, stopExecution }
}
