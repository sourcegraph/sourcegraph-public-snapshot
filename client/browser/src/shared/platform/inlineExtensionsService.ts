import { Subscribable, from } from 'rxjs'

import { isFirefox } from '@sourcegraph/common'
import { checkOk } from '@sourcegraph/http-client'
import { ExecutableExtension } from '@sourcegraph/shared/src/api/extension/activation'
import { ExtensionManifest } from '@sourcegraph/shared/src/extensions/extensionManifest'

import { isExtension } from '../context'

/**
 * Determine if inline extensions should be loaded.
 *
 * This requires the browser extension to be built with inline extensions enabled.
 * At build time this is determined by `shouldBuildWithInlineExtensions`.
 */
export const shouldUseInlineExtensions = (): boolean => isExtension && isFirefox()

/**
 * Get the manifest URL and script URL for a Sourcegraph extension which is inline (bundled with the browser add-on).
 */
function getURLsForInlineExtension(extensionName: string): { manifestURL: string; scriptURL: string } {
    return {
        manifestURL: browser.extension.getURL(`extensions/${extensionName}/package.json`),
        scriptURL: browser.extension.getURL(`extensions/${extensionName}/extension.js`),
    }
}

export function getInlineExtensions(): Subscribable<ExecutableExtension[]> {
    const extensionName = 'template'
    const { manifestURL, scriptURL } = getURLsForInlineExtension('template')
    const requestPromise = fetch(manifestURL)
        .then(response => checkOk(response).json())
        .then(
            (manifest: ExtensionManifest) =>
                [
                    {
                        id: `sourcegraph/${extensionName}`,
                        manifest,
                        scriptURL,
                    },
                ] as ExecutableExtension[]
        )

    return from(requestPromise)
}
