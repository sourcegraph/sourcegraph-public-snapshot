import { parseISO } from 'date-fns'
import React, { useEffect, useReducer } from 'react'

export interface DurationProps {
    /** The start time. */
    start: Date | string
    /** The end time. If not set, new Date() is used. Leave unset for timers. */
    end?: Date | string
}

/**
 * Prints a duration between two given timestamps or one given one and now.
 * Formats as hh:mm:ss.
 */
export const Duration: React.FunctionComponent<DurationProps> = ({ start, end }) => {
    // Parse the start date.
    const startDate = typeof start === 'string' ? parseISO(start) : start
    // Parse the end date.
    const endDate = typeof end === 'string' ? parseISO(end) : end || new Date()
    // Grab the total duration in seconds. We floor each timestamp, so we get stable
    // timing steps when rerenders occur off the second interval.
    let duration = Math.floor(endDate.getTime() / 1000) - Math.floor(startDate.getTime() / 1000)
    const hours = Math.floor(duration / (60 * 60))
    duration -= hours * 60 * 60
    const minutes = Math.floor(duration / 60)
    duration -= minutes * 60
    const seconds = Math.floor(duration)

    // If the end timestamp is not set, we want to auto rerender every second, so
    // this component is always up to date.
    const [, forceUpdate] = useReducer((any: number) => any + 1, 0)
    useEffect(() => {
        if (end === undefined) {
            const timer = setInterval(() => {
                forceUpdate()
            }, 1000)
            return () => {
                clearInterval(timer)
            }
        }
        return undefined
    }, [end])

    return (
        <>
            {leading0(hours)}:{leading0(minutes)}:{leading0(seconds)}
        </>
    )
}

/**
 * Returns the number as a string, with a leading 0 if it has only 1 digit.
 *
 * @param index The number to format.
 * @returns A string version of the formatted number.
 */
function leading0(index: number): string {
    if (index < 10) {
        return '0' + String(index)
    }
    return String(index)
}
