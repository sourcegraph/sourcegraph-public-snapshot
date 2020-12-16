import { parseISO } from 'date-fns'
import formatDistance from 'date-fns/formatDistance'
import formatDistanceStrict from 'date-fns/formatDistanceStrict'
import React, { useEffect, useState } from 'react'

interface Props {
    /** The date (if string, in ISO 8601 format). */
    date: string | Date | number

    /** Omit the "about". */
    noAbout?: boolean

    /** Function that returns the current time (for stability in visual tests). */
    now?: () => Date

    /** Whether to use exact timestamps (i.e. omit "less than", "about", etc.) */
    strict?: boolean
}

const RERENDER_INTERVAL_MSEC = 7000

/**
 * Displays a date's relative time ("... ago") and shows the full date on hover. Re-renders
 * periodically to ensure the relative time string is up-to-date.
 */
export const Timestamp: React.FunctionComponent<Props> = ({
    date,
    noAbout = false,
    strict = false,
    now = Date.now,
}) => {
    const [label, setLabel] = useState<string>(calculateLabel(date, now, strict, noAbout))
    useEffect(() => {
        const intervalHandle = window.setInterval(
            () => setLabel(calculateLabel(date, now, strict, noAbout)),
            RERENDER_INTERVAL_MSEC
        )
        return () => {
            window.clearInterval(intervalHandle)
        }
    }, [date, noAbout, now, strict])

    return (
        <span className="timestamp" data-tooltip={date}>
            {label}
        </span>
    )
}

function calculateLabel(
    date: string | Date | number,
    now: () => Date | number,
    strict: boolean,
    noAbout: boolean
): string {
    let label: string
    if (strict) {
        label = formatDistanceStrict(typeof date === 'string' ? parseISO(date) : date, now(), {
            addSuffix: true,
        })
    } else {
        label = formatDistance(typeof date === 'string' ? parseISO(date) : date, now(), {
            addSuffix: true,
            includeSeconds: true,
        })
    }
    if (noAbout) {
        label = label.replace('about ', '')
    }
    return label
}
