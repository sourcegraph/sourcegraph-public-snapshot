import { Observable, from } from 'rxjs'
import { map } from 'rxjs/operators'

import { checkOk, isErrorGraphQLResult, gql } from '@sourcegraph/http-client'
import { ExecutableExtension } from '@sourcegraph/shared/src/api/extension/activation'
import { ExtensionManifest } from '@sourcegraph/shared/src/extensions/extensionManifest'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'

import extensions from '../../../code-intel-extensions.json'
import { EnableLegacyExtensionsResult } from '../../graphql-operations'

const DEFAULT_ENABLE_LEGACY_EXTENSIONS = false

/**
 * Determine which extensions should be loaded:
 * - inline (bundled with the browser extension)
 * - or from the extensions registry (if `enableLegacyExtensions` experimental feature value is set to `true`).
 */
export const shouldUseInlineExtensions = (requestGraphQL: PlatformContext['requestGraphQL']): Observable<boolean> =>
    requestGraphQL<EnableLegacyExtensionsResult>({
        request: gql`
            query EnableLegacyExtensions {
                site {
                    enableLegacyExtensions
                }
            }
        `,
        variables: {},
        mightContainPrivateInfo: false,
    }).pipe(
        map(result => {
            if (isErrorGraphQLResult(result)) {
                // EnableLegacyExtensions query resolver may not be implemented on older versions.
                // Return `true` by default.
                return true
            }

            try {
                const enableLegacyExtensions = result.data.site.enableLegacyExtensions
                return typeof enableLegacyExtensions === 'undefined'
                    ? DEFAULT_ENABLE_LEGACY_EXTENSIONS
                    : enableLegacyExtensions
            } catch {
                return DEFAULT_ENABLE_LEGACY_EXTENSIONS
            }
        }),
        map(enableLegacyExtensions => {
            // TODO: The Phabricator native extension is currently the only runtime that runs the
            // browser extension code but does not use bundled extensions yet. We will fix this
            // when we update the browser extensions to use the new code intel APIs (#42104).
            if (window.SOURCEGRAPH_PHABRICATOR_EXTENSION === true) {
                return false
            }

            return !enableLegacyExtensions
        })
    )

/**
 * Get the manifest URL and script URL for a Sourcegraph extension which is inline (bundled with the browser add-on).
 */
function getURLsForInlineExtension(extensionID: string): { manifestURL: string; scriptURL: string } {
    const kebabCaseExtensionID = extensionID.replace(/^sourcegraph\//, 'sourcegraph-')

    const manifestPath = `extensions/${kebabCaseExtensionID}/package.json`
    const scriptPath = `extensions/${kebabCaseExtensionID}/extension.js`

    // In a browser extension environment, we need to find the absolute ULR in the browser extension
    // asssets directory.
    if (typeof window.browser !== 'undefined') {
        return {
            manifestURL: browser.extension.getURL(manifestPath),
            scriptURL: browser.extension.getURL(scriptPath),
        }
    }
    // If SOURCEGRAPH_ASSETS_URL is defined (happening for e.g. GitLab native extenisons), we can
    // use this URL as the root for bundled assets.
    if (window.SOURCEGRAPH_ASSETS_URL) {
        return {
            manifestURL: new URL(manifestPath, window.SOURCEGRAPH_ASSETS_URL).toString(),
            scriptURL: new URL(scriptPath, window.SOURCEGRAPH_ASSETS_URL).toString(),
        }
    }

    throw new Error('Can not construct extension URLs')
}

export function getInlineExtensions(): Observable<ExecutableExtension[]> {
    const promises: Promise<ExecutableExtension>[] = []

    for (const extensionID of extensions) {
        const { manifestURL, scriptURL } = getURLsForInlineExtension(extensionID)
        promises.push(
            fetch(manifestURL)
                .then(response => checkOk(response).json())
                .then(
                    (manifest: ExtensionManifest): ExecutableExtension => ({
                        id: extensionID,
                        manifest,
                        scriptURL,
                    })
                )
        )
    }

    return from(Promise.all(promises))
}
