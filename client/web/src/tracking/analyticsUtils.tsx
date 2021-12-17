import { fromEvent, concat, Observable, of } from 'rxjs'
import { fromFetch } from 'rxjs/fetch'
import { catchError, filter, map, mapTo, publishReplay, refCount, take } from 'rxjs/operators'

import { isFirefox } from '@sourcegraph/shared/src/util/browserDetection'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { IS_CHROME } from '../marketing/util'
import { observeQuerySelector } from '../util/dom'

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
          map(({ detail }) => {
              try {
                  return { platform: detail?.platform }
              } catch (error) {
                  // Temporary to fix issues on Firefox (https://github.com/sourcegraph/sourcegraph/issues/25998)
                  if (
                      isFirefox() &&
                      isErrorLike(error) &&
                      error.message.includes('Permission denied to access property "platform"')
                  ) {
                      return { platform: 'firefox-extension' }
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

const CHROME_EXTENSION_ID = 'dgjhfomjieaadpoljlnidmbgkdffpack'

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

/**
 * Indicates if the current user has the browser extension installed. It waits 1000ms for the browser
 * extension to inject a DOM marker element, and if it doesn't, emits false
 */
export const browserExtensionInstalled: Observable<boolean> = concat(
    checkChromeExtensionInstalled().pipe(filter(isInstalled => isInstalled)),
    observeQuerySelector({ selector: '#sourcegraph-app-background', timeoutMs: 1000 }).pipe(
        mapTo(true),
        catchError(() => [false])
    )
).pipe(
    take(1),
    // Replay the same latest value for every subscriber
    publishReplay(1),
    refCount()
)
