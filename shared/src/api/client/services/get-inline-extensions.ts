import { Subscribable, from } from 'rxjs'
import { ExtensionManifest } from '../../../schema/extensionSchema'
import { ExecutableExtension } from './extensionsService'
import { checkOk } from '../../../backend/fetch'

/**
 * Get the manifest URL and script URL for a Sourcegraph extension which is inline (bundled with the browser addon).
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
    // return fromFetch(manifestURL, undefined, response => checkOk(response).json() as Promise<ExtensionManifest>).pipe(
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
