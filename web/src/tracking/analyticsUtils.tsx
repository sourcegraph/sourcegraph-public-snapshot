import { eventLogger } from './eventLogger'

export interface EventQueryParameters {
    utm_campaign?: string
    utm_source?: string
    utm_product_name?: string
    utm_product_version?: string
}

/**
 * the browser extension is detected when it creates a div with id `sourcegraph-app-background` on page.
 * for on-premise or testing instances of Sourcegraph, the browser extension never runs, so this will return false.
 * proceed with caution.
 */
export function hasBrowserExtensionInstalled(): boolean {
    return document.getElementById('sourcegraph-app-background') !== null
}

/**
 * Get pageview-specific event properties from URL query string parameters
 */
export function pageViewQueryParameters(url: string): EventQueryParameters {
    const parsedUrl = new URL(url)
    return {
        utm_campaign: parsedUrl.searchParams.get('utm_campaign') || undefined,
        utm_source: parsedUrl.searchParams.get('utm_source') || undefined,
        utm_product_name: parsedUrl.searchParams.get('utm_product_name') || undefined,
        utm_product_version: parsedUrl.searchParams.get('utm_product_version') || undefined,
    }
}

/**
 * Log events associated with URL query string parameters, and remove those parameters as necessary
 * Note that this is a destructive operation (it changes the page URL and replaces browser state) by
 * calling stripURLParameters
 */
export function handleQueryEvents(url: string): void {
    const parsedUrl = new URL(url)
    const eventParameters: { [key: string]: string } = {}
    for (const [key, val] of parsedUrl.searchParams.entries()) {
        eventParameters[camelCaseToUnderscore(key)] = val
    }
    const eventName = parsedUrl.searchParams.get('_event')
    const isBadgeRedirect = !!parsedUrl.searchParams.get('badge')
    if (eventName || isBadgeRedirect) {
        if (isBadgeRedirect) {
            eventLogger.log('RepoBadgeRedirected', eventParameters)
        } else if (eventName === 'CompletedAuth0SignIn') {
            eventLogger.log('CompletedAuth0SignIn', eventParameters)
        } else if (eventName === 'SignupCompleted') {
            eventLogger.log('SignupCompleted', eventParameters)
        } else if (eventName) {
            eventLogger.log(eventName, eventParameters)
        }
    }

    stripURLParameters(url, [
        '_event',
        '_source',
        'utm_campaign',
        'utm_source',
        'utm_product_name',
        'utm_product_version',
        'badge',
    ])
}

/**
 * Strip provided URL parameters and update window history
 */
function stripURLParameters(url: string, paramsToRemove: string[] = []): void {
    const parsedUrl = new URL(url)
    for (const key of paramsToRemove) {
        if (parsedUrl.searchParams.has(key)) {
            parsedUrl.searchParams.delete(key)
        }
    }
    window.history.replaceState({}, window.document.title, parsedUrl.href)
}

function camelCaseToUnderscore(input: string): string {
    if (input.charAt(0) === '_') {
        input = input.substring(1)
    }
    return input.replace(/([A-Z])/g, $1 => `_${$1.toLowerCase()}`)
}
