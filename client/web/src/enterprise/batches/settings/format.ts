import { upperFirst, capitalize } from 'lodash'

/**
 * Formats the days of the week for a rollout window for display.
 *
 * If days are provided, join them with commas and capitalize each day's name.
 * Otherwise, returns 'every other day'.
 *
 * @param days The days of the week for the rollout window, e.g. ['monday', 'wednesday']
 * @returns The formatted days for display in the UI
 */
export const formatDays = (days: string[] | undefined): string => {
    if (days && days.length > 0) {
        return days.join(', ').replaceAll(/\w+/g, capitalize)
    }

    return 'every other day'
}

/**
 * Formats the rollout window rate for display.
 *
 * According to the schema, if the rate is a number then it can only be zero.
 * If the rate starts with '0/' then we revert to displaying None, since this is the same as 0.
 * Otherwise, we display the rate in a readable format, e.g. '2 changesets per minute'.
 *
 * https://sourcegraph.sourcegraph.com/github.com/sourcegraph/sourcegraph@3ee30bb/-/blob/schema/site.schema.json?L567-571
 *
 * @param rate The rollout window rate, either a number or a string like '1/minute'
 * @returns The formatted rate for display
 */
export const formatRate = (rate: string | number): string => {
    if (typeof rate === 'number' || rate.startsWith('0/')) {
        return 'None'
    }
    return upperFirst(rate.replace('/', ' changesets per '))
}
