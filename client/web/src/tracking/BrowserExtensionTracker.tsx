import React, { useEffect } from 'react'

import { useLocation } from 'react-router'
import { fromEvent, Observable, merge, EMPTY } from 'rxjs'
import { catchError, map, publishReplay, refCount, take } from 'rxjs/operators'

import { isErrorLike, isFirefox } from '@sourcegraph/common'
import { useLocalStorage, useObservable } from '@sourcegraph/wildcard'

import { observeQuerySelector } from '../util/dom'

const BROWSER_EXTENSION_UTM_SOURCES = new Set(['safari-extension', 'firefox-extension', 'chrome-extension'])
const EXTENSION_MARKER_ID = '#sourcegraph-app-background'
const BROWSER_EXTENSION_LAST_DETECTION_KEY = 'integrations.browser.lastDetectionTimestamp'

/**
 * Indicates if the webapp ever receives a message from the user's Sourcegraph browser extension,
 * either in the form of a DOM marker element, or from a CustomEvent.
 */
export const browserExtensionMessageReceived: Observable<{ platform?: string; version?: string }> = merge(
    // If the marker exists, the extension is installed
    observeQuerySelector({ selector: EXTENSION_MARKER_ID, timeout: 10000 }).pipe(
        map(extensionMarker => ({
            platform: (extensionMarker as HTMLElement)?.dataset?.platform,
            version: (extensionMarker as HTMLElement)?.dataset?.version,
        })),
        catchError(() => EMPTY)
    ),
    // If not, listen for a registration event
    fromEvent<CustomEvent<{ platform?: string; version?: string }>>(
        document,
        'sourcegraph:browser-extension-registration'
    ).pipe(
        take(1),
        map(({ detail }) => {
            try {
                return { platform: detail?.platform, version: detail?.version }
            } catch (error) {
                // Temporary to fix issues on Firefox (https://github.com/sourcegraph/sourcegraph/issues/25998)
                if (
                    isFirefox() &&
                    isErrorLike(error) &&
                    error.message.includes('Permission denied to access property "platform"')
                ) {
                    return {
                        platform: 'firefox-extension',
                        version: 'unknown due to <<Permission denied to access property "platform">>',
                    }
                }

                throw error
            }
        })
    )
).pipe(
    // Replay the same latest value for every subscriber
    publishReplay(1),
    refCount()
)

/**
 * This component uses extension marker DOM element or UTM parameters to detect incoming traffic from our browser extensions (Chrome, Safari, Firefox)
 * and updates a localStorage whenever these are found.
 */

export const BrowserExtensionTracker: React.FunctionComponent<React.PropsWithChildren<unknown>> = React.memo(() => {
    const [, setLastBrowserExtensionDetection] = useLocalStorage(BROWSER_EXTENSION_LAST_DETECTION_KEY, 0)

    // OPTION 1: Use initial page load query parameters to detect inbound link from browser extension
    const location = useLocation()

    useEffect(() => {
        const parameters = new URLSearchParams(location.search)
        const utmSource = parameters.get('utm_source') ?? ''

        if (BROWSER_EXTENSION_UTM_SOURCES.has(utmSource)) {
            setLastBrowserExtensionDetection(Date.now())
        }

        // We only want to capture the query parameters on the first page load. In order to avoid
        // rerunning the effect whenever location change, we skip it from the dependency array.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [setLastBrowserExtensionDetection])

    // OPTION 2: Use browser extension marker on the page
    const isBrowserExtensionMessageReceived = useObservable(browserExtensionMessageReceived)

    useEffect(() => {
        if (isBrowserExtensionMessageReceived) {
            setLastBrowserExtensionDetection(Date.now())
        }
    }, [isBrowserExtensionMessageReceived, setLastBrowserExtensionDetection])

    return null
})

BrowserExtensionTracker.displayName = 'BrowserExtensionTracker'
