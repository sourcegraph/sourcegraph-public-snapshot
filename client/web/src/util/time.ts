import { pluralize } from '@sourcegraph/common'

const units = [
    { denominator: 1000 * 60 * 60 * 24, name: 'day' },
    { denominator: 1000 * 60 * 60, name: 'hour' },
    { denominator: 1000 * 60, name: 'minute' },
    { denominator: 1000, name: 'second' },
    { denominator: 1, name: 'millisecond' },
]

type unitName = typeof units[number]['name']

export interface StructuredDuration {
    amount: number
    unit: unitName
}

/**
 * This is essentially to date-fns/formatDistance with support for milliseconds.
 *
 * Examples for the output:
 * - "1 day"
 * - "2 days and 1 hour"
 * - "1 minute and 5 seconds"
 * - "5 seconds"
 * - "1 millisecond".
 *
 * The output has the following properties:
 *
 * - Consists of either one unit ("x days") or two units ("x days and y hours")
 * - If there are more than one unit, they are adjacent (never "x days and y minutes")
 * - If there is a greater unit, the value will not exceed the next threshold (e.g. `2 minutes and 5 seconds`, never `125 seconds`).
 *
 * @param millis The number of milliseconds elapsed.
 */
export function formatDurationLong(millis: number): string {
    const parts = formatDurationStructured(millis)

    const description = parts
        .slice(0, 2)
        .map(part => `${part.amount} ${pluralize(part.unit, part.amount)}`)
        .join(' and ')

    // If description is empty return a canned string
    return description || '0 milliseconds'
}

function formatDurationStructured(millis: number): StructuredDuration[] {
    const parts: { amount: number; unit: string }[] = []

    // Construct a list of parts like `1 day` or `7 hours` in descending
    // order. If the value is zero, an empty string is added to the list.`
    units.reduce((msRemaining, { denominator, name }) => {
        // Determine how many units can fit into the current value
        const part = Math.floor(msRemaining / denominator)
        // Format this part (pluralize if value is more than one)
        if (part > 0) {
            parts.push({ amount: part, unit: name })
        }
        // Remove this order's contribution to the current value
        return msRemaining - part * denominator
    }, millis)

    return parts
}
