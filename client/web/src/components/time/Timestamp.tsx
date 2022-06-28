import React, { useEffect, useMemo, useState } from 'react'

import { parseISO, format } from 'date-fns'
import formatDistance from 'date-fns/formatDistance'
import formatDistanceStrict from 'date-fns/formatDistanceStrict'

interface Props {
    /** The date (if string, in ISO 8601 format). */
    date: string | Date | number

    /** Omit the "about". */
    noAbout?: boolean

    /** Function that returns the current time (for stability in visual tests). */
    now?: () => Date

    /** Whether to use exact timestamps (i.e. omit "less than", "about", etc.) */
    strict?: boolean

    /** Whether to show absolute timestamp and show relative one in tooltip */
    preferAbsolute?: boolean
}

const RERENDER_INTERVAL_MSEC = 7000

/**
 * Displays a date's relative time ("... ago") and shows the full date on hover. Re-renders
 * periodically to ensure the relative time string is up-to-date.
 */
export const Timestamp: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    date,
    noAbout = false,
    strict = false,
    now = Date.now,
    preferAbsolute = false,
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

    const tooltip = useMemo(() => {
        const parsedDate = typeof date === 'string' ? parseISO(date) : new Date(date)
        const dateHasTime = date.toString().includes('T')
        return format(parsedDate, `yyyy-MM-dd${dateHasTime ? ' pp' : ''}`)
    }, [date])

    return (
        <span className="timestamp" data-tooltip={preferAbsolute ? label : tooltip}>
            {preferAbsolute ? tooltip : label}
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
