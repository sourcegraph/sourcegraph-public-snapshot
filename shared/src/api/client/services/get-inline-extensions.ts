import { Subscribable } from 'rxjs'
import { ExtensionManifest } from '../../../schema/extensionSchema'
import { ExecutableExtension } from './extensionsService'
import { map } from 'rxjs/operators'
import { fromFetch } from '../../../graphql/fromFetch'
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
    return fromFetch(manifestURL, undefined, response => checkOk(response).json() as Promise<ExtensionManifest>).pipe(
        map(
            manifest =>
                [
                    {
                        id: `sourcegraph/${extensionName}`,
                        manifest,
                        scriptURL,
                    },
                ] as ExecutableExtension[]
        )
    )
}
