import React, { useState, useEffect } from 'react'

import { useLocation } from 'react-router'
import { fromEvent, concat, Observable, of, merge, EMPTY, OperatorFunction } from 'rxjs'
import { fromFetch } from 'rxjs/fetch'
import { catchError, filter, map, mapTo, mergeMap, publishReplay, refCount, take } from 'rxjs/operators'

import { isErrorLike, isFirefox } from '@sourcegraph/common'
import { useLocalStorage, useObservable } from '@sourcegraph/wildcard'

import { IS_CHROME } from '../marketing/util'
import { observeQuerySelector } from '../util/dom'

const BROWSER_EXTENSION_UTM_SOURCES = new Set(['safari-extension', 'firefox-extension', 'chrome-extension'])
const EXTENSION_MARKER_ID = '#sourcegraph-app-background'
const BROWSER_EXTENSION_LAST_DETECTION_KEY = 'integrations.browser.lastDetectionTimestamp'
const WEEK = 1000 * 60 * 60 * 24 * 7
const CHROME_EXTENSION_ID = 'dgjhfomjieaadpoljlnidmbgkdffpack'

/**
 * Indicates if the webapp ever receives a message from the user's Sourcegraph browser extension,
 * either in the form of a DOM marker element, or from a CustomEvent.
 *
 * NOTE: You should likely use useIsBrowserExtensionActiveUser, rather than browserExtensionMessageReceived,
 * which may never emit or complete.
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
 * A better way to check if the Chrome extension is installed that doesn't depend on the extension having permissions to the page.
 * This is not possible on Firefox though.
 */
const checkChromeExtensionInstalled = (): Observable<boolean> => {
    if (!IS_CHROME) {
        return of(false)
    }
    return fromFetch(`chrome-extension://${CHROME_EXTENSION_ID}/img/icon-16.png`, {
        selector: response => of(response.ok),
    }).pipe(catchError(() => of(false)))
}

function pipeIf<T>(predicate: (v: T) => boolean, ...pipes: OperatorFunction<T, T>[]) {
    return function (source: Observable<T>) {
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        return source.pipe(mergeMap(value => (predicate(value) ? of(value).pipe(...pipes) : of(value))))
    }
}

/**
 * Returns whether code is running in jest unit tests or not
 */
const isUnitTestRunning = (): boolean => process.env.JEST_WORKER_ID !== undefined

/**
 * Indicates if the current user has the browser extension installed. It waits 1000ms for the browser
 * extension to inject a DOM marker element, and if it doesn't, emits false
 */
const browserExtensionInstalled: Observable<boolean> = concat(
    checkChromeExtensionInstalled().pipe(filter(isInstalled => isInstalled)),
    observeQuerySelector({ selector: EXTENSION_MARKER_ID, timeout: 1000 }).pipe(
        mapTo(true),
        catchError(() => [false])
    )
).pipe(
    take(1),
    // Replay the same latest value for every subscriber unless running in tests
    pipeIf(() => !isUnitTestRunning(), publishReplay(1), refCount())
)

/**
 * Returns whether user has currently installed browser extension or have used it for the past week on this particular browser.
 */
export function useIsBrowserExtensionActiveUser(): boolean | undefined {
    // NOTE: intentionally using localStorage to keep track per browser instead user settings across browsers
    const [lastBrowserExtensionDetection] = useLocalStorage(BROWSER_EXTENSION_LAST_DETECTION_KEY, 0)
    const isBrowserExtensionInstalled = useObservable(browserExtensionInstalled)
    const [now] = useState<number>(Date.now())

    if (lastBrowserExtensionDetection && now - lastBrowserExtensionDetection < WEEK) {
        return true
    }

    return isBrowserExtensionInstalled
}

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
