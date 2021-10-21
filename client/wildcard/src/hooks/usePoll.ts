import { useState, useEffect, useRef, useCallback } from 'react'

interface UsePollControls {
    isPolling: boolean
    startPolling: () => void
    stopPolling: () => void
}

/**
 * Custom hook for controlling the polling of a function.
 *
 * @param callback The function to call on the polling interval
 * @param interval The interval at which the polling should occur, in ms
 */
export const usePoll = (
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    callback: () => any,
    interval: number
): UsePollControls => {
    const [pollInterval, setPollInterval] = useState(interval)
    const isPolling = pollInterval > 0
    const isMounted = useRef(true)

    const savedCallback = useRef(callback)
    // Remember the latest callback.
    useEffect(() => {
        savedCallback.current = callback
    }, [callback])

    // Set up the interval to actually run the callback.
    useEffect(() => {
        if (isMounted.current && isPolling) {
            const timeout = setInterval(() => {
                savedCallback.current()
            }, interval)
            return () => clearInterval(timeout)
        }
        return
    }, [interval, isPolling])

    // Callback to stop the polling.
    const stopPolling = useCallback(() => {
        if (isMounted.current && isPolling) {
            setPollInterval(-1)
        }
    }, [isPolling])

    // Callback to start the polling.
    const startPolling = useCallback(() => {
        if (isMounted.current && !isPolling) {
            setPollInterval(interval)
        }
    }, [interval, isPolling])

    useEffect(() => {
        isMounted.current = true
        startPolling()
        return () => {
            isMounted.current = false
            stopPolling()
        }
    }, [startPolling, stopPolling])

    return { isPolling, startPolling, stopPolling }
}
