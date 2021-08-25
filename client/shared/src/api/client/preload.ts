import { Observable } from 'rxjs'
import { distinctUntilChanged, map, startWith } from 'rxjs/operators'

import { ConfiguredExtension, getScriptURLFromExtensionManifest } from '../../extensions/extension'
import { getModeFromPath } from '../../languages'
import { isDefined } from '../../util/types'
import { ExecutableExtension, extensionsWithMatchedActivationEvent } from '../extension/activation'

/** Ensure that we only add <link> tags once for each scriptURL */
const loadedScriptURLs = new Set<string>()

/**
 * Attaches <link> tags to preload extension bundles in parallel.
 * If we were to download extensions with `importScripts` in the extension host Web Worker,
 * extension bundles would be downloaded sequentially. Ensure that `preloadExtensions`
 * is run in the main thread before the extension host requests extension bundles;
 * that way, the browser will share the request if it's inflight, or respond with the
 * cached bundle if it has been fulfilled.
 *
 * Only run this logic in the web app since the HTTP cache is not shared
 * between the browser extension's content script (which runs in the code host's context)
 * and background page. Don't waste resources by loading extensions twice.
 */
export function preloadExtensions({
    extensions,
    languages,
}: {
    extensions: ConfiguredExtension[]
    languages: Set<string>
}): void {
    try {
        const executableExtensions = extensions
            .map(extension => {
                try {
                    // No need to call PlatformContext#getScriptURL: that's a bext only workaround, and we won't run this code in bext.
                    // There also will not be inline extensions in the web app. Just look in the manifest for script URL.
                    const scriptURL = getScriptURLFromExtensionManifest(extension)

                    const executableExtension: ExecutableExtension = {
                        id: extension.id,
                        manifest: extension.manifest,
                        scriptURL,
                    }

                    return executableExtension
                } catch {
                    // Couldn't find scriptURL, skip this extension
                    return null
                }
            })
            .filter(isDefined)

        for (const activeExtension of extensionsWithMatchedActivationEvent(executableExtensions, languages)) {
            if (loadedScriptURLs.has(activeExtension.scriptURL)) {
                continue
            }
            const preloadLink = document.createElement('link')
            preloadLink.href = activeExtension.scriptURL
            preloadLink.rel = 'preload'
            preloadLink.as = 'script'
            document.head.append(preloadLink)

            loadedScriptURLs.add(activeExtension.scriptURL)
        }
    } catch (error) {
        console.error('Error preloading Sourcegraph extensions:', error)
    }
}

/**
 * Returns an Observable of language ID based on the URL.
 * To be used only to guess which language extension to preload.
 *
 * It's ok if this isn't perfect/results in false negatives, since even in the unlikely case that this fails,
 * the necessary extensions will be loaded in the extension host anyways.
 */
export function observeLanguage(): Observable<string> {
    const pathnames = new Observable<string>(function subscribe(observer) {
        const mutationObserver = new MutationObserver(() => {
            observer.next(location.pathname)
        })
        mutationObserver.observe(document, {
            childList: true,
            subtree: true,
        })

        return function unsubscribe() {
            mutationObserver.disconnect()
        }
    })

    // On each mutation, check if the language extension has changed
    return pathnames.pipe(
        startWith(location.pathname),
        distinctUntilChanged(),
        map(path => getModeFromPath(path)),
        distinctUntilChanged()
    )
}
