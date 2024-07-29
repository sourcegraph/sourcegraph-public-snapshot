import React, { useEffect, useReducer } from 'react'

import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'
import { parseISO } from 'date-fns'

import styles from './Duration.module.scss'

export interface DurationProps extends React.HTMLAttributes<HTMLDivElement> {
    /** The start time. */
    start: Date | string
    /** The end time. If not set, new Date() is used. Leave unset for timers. */
    end?: Date | string
    /**
     * If true, will ensure the duration is rendered at a stable width, even if the time
     * is changing (e.g. for a timer). Default is true.
     */
    stableWidth?: boolean
    /** String to precede screen-reader readout of the duration. */
    labelPrefix?: string
    /** String to follow screen-reader readout of the duration. */
    labelSuffix?: string
}

/**
 * Prints a duration between two given timestamps or one given one and now.
 * Formats as hh:mm:ss.
 */
export const Duration: React.FunctionComponent<React.PropsWithChildren<DurationProps>> = React.memo(function Duration({
    start,
    end,
    className,
    stableWidth = true,
    labelPrefix,
    labelSuffix,
    ...props
}) {
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

    const label = `${labelPrefix || ''} ${hours} hours, ${minutes} minutes, and ${seconds} seconds ${
        labelSuffix || ''
    }`.trim()

    return (
        <div
            className={classNames({ [styles.stableWidth]: stableWidth }, className)}
            {...props}
            role={end === undefined ? 'timer' : undefined}
        >
            {stableWidth && (
                // Set the width of the parent with a filler block of full-width digits,
                // to prevent layout shift if the time changes.
                // NOTE: This would not be a problem if we used a monospace font instead.
                <span className={styles.filler} aria-hidden={true}>
                    00:00:00
                </span>
            )}
            <VisuallyHidden>{label}</VisuallyHidden>
            <span className={styles.duration} aria-hidden={true}>
                {leading0(hours)}:{leading0(minutes)}:{leading0(seconds)}
            </span>
        </div>
    )
})

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
