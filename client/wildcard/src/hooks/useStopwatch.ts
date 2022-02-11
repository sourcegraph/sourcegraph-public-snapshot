import { useCallback, useState } from 'react'

import { useInterval } from './useInterval'

export interface Time {
    milliseconds: number
    seconds: number
    minutes: number
    hours: number
}

const getTime = (milliseconds: number): Time => {
    const seconds = Math.ceil(milliseconds / 1000)
    const minutes = Math.floor(seconds / 60)
    const hours = Math.floor(minutes / 60)

    return { milliseconds, seconds, minutes, hours }
}

export interface UseStopwatchControls {
    time: Time
    start: () => void
    stop: () => void
    isRunning: boolean
}

/**
 * Custom hook which provides stopwatch functionality to track elapsed time.
 *
 * @param autostart whether or not the stopwatch should start automatically
 */
export const useStopwatch = (autostart: boolean = true): UseStopwatchControls => {
    const [startTime, setStartTime] = useState(new Date().getTime())
    const [latestTime, setLatestTime] = useState(new Date().getTime())
    const [isRunning, setIsRunning] = useState(autostart)

    const updateLatestTime = useCallback(() => setLatestTime(new Date().getTime()), [])
    useInterval(updateLatestTime, isRunning ? 1000 : -1)

    const start = useCallback((): void => {
        setStartTime(new Date().getTime())
        setLatestTime(new Date().getTime())
        setIsRunning(true)
    }, [])

    const stop = useCallback((): void => {
        setIsRunning(false)
    }, [])

    return { time: getTime(latestTime - startTime), start, stop, isRunning }
}
