import { formatISO, startOfWeek } from 'date-fns'

export const DOTCOM_URL = new URL('https://sourcegraph.com')

/**
 * Strip provided URL parameters and update window history
 */
export function stripURLParameters(url: string, parametersToRemove: string[] = []): void {
    const parsedUrl = new URL(url)
    const existingParameters = parametersToRemove.filter(key => parsedUrl.searchParams.has(key))

    // Update history state only if we have parameters to remove in the url.
    if (existingParameters.length !== 0) {
        for (const key of existingParameters) {
            parsedUrl.searchParams.delete(key)
        }

        window.history.replaceState(window.history.state, window.document.title, parsedUrl.href)
    }
}

/**
 * Returns the Monday at or before the supplied date, in YYYY-MM-DD format.
 * This is used to generate cohort IDs for users who
 * started using the site on the same week.
 */
export function getPreviousMonday(date: Date): string {
    return formatISO(startOfWeek(date, { weekStartsOn: 1 }), { representation: 'date' })
}
