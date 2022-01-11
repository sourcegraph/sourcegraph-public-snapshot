import { isDefined } from '@sourcegraph/common'

import { ConfiguredExtension, getScriptURLFromExtensionManifest } from '../../extensions/extension'
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
