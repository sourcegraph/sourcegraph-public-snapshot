import { fromEvent } from 'rxjs/observable/fromEvent'
import { of } from 'rxjs/observable/of'
import { catchError } from 'rxjs/operators/catchError'
import { mapTo } from 'rxjs/operators/mapTo'
import { publishReplay } from 'rxjs/operators/publishReplay'
import { refCount } from 'rxjs/operators/refCount'
import { take } from 'rxjs/operators/take'
import { timeout } from 'rxjs/operators/timeout'
import { eventLogger } from './eventLogger'

export interface EventQueryParameters {
    utm_campaign?: string
    utm_source?: string
    utm_product_name?: string
    utm_product_version?: string
}

/**
 * Indicates if the webapp ever receives a message from the user's Sourcegraph browser extension,
 * either in the form of a DOM marker element, or from a CustomEvent.
 *
 * You should likely use browserExtensionInstalled, rather than _browserExtensionMessageReceived,
 * which may never emit or complete.
 */
export const browserExtensionMessageReceived = (document.getElementById('sourcegraph-app-background')
    ? // If the marker exists, the extension is installed
      of(true)
    : // If not, listen for a registration event
      fromEvent<CustomEvent>(document, 'sourcegraph:browser-extension-registration').pipe(take(1), mapTo(true))
).pipe(
    // Replay the same latest value for every subscriber
    publishReplay(1),
    refCount()
)

/**
 * Indicates if the current user has the browser extension installed. It waits 500ms for the browser
 * extension to fire a registration event, and if it doesn't, emits false
 */
export const browserExtensionInstalled = browserExtensionMessageReceived.pipe(
    timeout(500),
    catchError(err => {
        if (err.name === 'TimeoutError') {
            return [false]
        }
        throw err
    }),
    // Replay the same latest value for every subscriber
    publishReplay(1),
    refCount()
)

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
    window.history.replaceState(window.history.state, window.document.title, parsedUrl.href)
}

function camelCaseToUnderscore(input: string): string {
    if (input.charAt(0) === '_') {
        input = input.substring(1)
    }
    return input.replace(/([A-Z])/g, $1 => `_${$1.toLowerCase()}`)
}
