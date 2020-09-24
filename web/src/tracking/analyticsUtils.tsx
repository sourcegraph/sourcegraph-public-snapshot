import { fromEvent, of } from 'rxjs'
import { catchError, map, publishReplay, refCount, take, timeout } from 'rxjs/operators'
import { eventLogger } from './eventLogger'
import { asError } from '../../../shared/src/util/errors'

interface EventQueryParameters {
    utm_campaign?: string
    utm_source?: string
    utm_medium?: string
}

const extensionMarker = document.querySelector<HTMLDivElement>('#sourcegraph-app-background')

/**
 * Indicates if the webapp ever receives a message from the user's Sourcegraph browser extension,
 * either in the form of a DOM marker element, or from a CustomEvent.
 *
 * You should likely use browserExtensionInstalled, rather than _browserExtensionMessageReceived,
 * which may never emit or complete.
 */
export const browserExtensionMessageReceived = (extensionMarker
    ? // If the marker exists, the extension is installed
      of({ platform: extensionMarker.dataset?.platform })
    : // If not, listen for a registration event
      fromEvent<CustomEvent>(document, 'sourcegraph:browser-extension-registration').pipe(
          take(1),
          map(({ detail }) => ({
              platform: detail?.platform,
          }))
      )
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
    catchError(error => {
        if (asError(error).name === 'TimeoutError') {
            return [false]
        }
        throw error
    }),
    catchError(() => [false]),
    // Replay the same latest value for every subscriber
    publishReplay(1),
    refCount()
)

/**
 * Get pageview-specific event properties from URL query string parameters
 */
export function pageViewQueryParameters(url: string): EventQueryParameters {
    const parsedUrl = new URL(url)

    const utmSource = parsedUrl.searchParams.get('utm_source')
    if (utmSource === 'saved-search-email') {
        eventLogger.log('SavedSearchEmailClicked')
    } else if (utmSource === 'saved-search-slack') {
        eventLogger.log('SavedSearchSlackClicked')
    }

    return {
        utm_campaign: parsedUrl.searchParams.get('utm_campaign') || undefined,
        utm_source: parsedUrl.searchParams.get('utm_source') || undefined,
        utm_medium: parsedUrl.searchParams.get('utm_medium') || undefined,
    }
}

/**
 * Log events associated with URL query string parameters, and remove those parameters as necessary
 * Note that this is a destructive operation (it changes the page URL and replaces browser state) by
 * calling stripURLParameters
 */
export function handleQueryEvents(url: string): void {
    const parsedUrl = new URL(url)
    const isBadgeRedirect = !!parsedUrl.searchParams.get('badge')
    if (isBadgeRedirect) {
        eventLogger.log('RepoBadgeRedirected')
    }

    stripURLParameters(url, ['utm_campaign', 'utm_source', 'utm_medium', 'badge'])
}

/**
 * Strip provided URL parameters and update window history
 */
function stripURLParameters(url: string, parametersToRemove: string[] = []): void {
    const parsedUrl = new URL(url)
    for (const key of parametersToRemove) {
        if (parsedUrl.searchParams.has(key)) {
            parsedUrl.searchParams.delete(key)
        }
    }
    window.history.replaceState(window.history.state, window.document.title, parsedUrl.href)
}
