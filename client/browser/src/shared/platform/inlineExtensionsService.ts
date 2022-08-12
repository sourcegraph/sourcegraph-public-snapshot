import { Observable, from } from 'rxjs'
import { map } from 'rxjs/operators'

import { checkOk, isErrorGraphQLResult, gql } from '@sourcegraph/http-client'
import { ExecutableExtension } from '@sourcegraph/shared/src/api/extension/activation'
import { ExtensionManifest } from '@sourcegraph/shared/src/extensions/extensionManifest'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import * as GQL from '@sourcegraph/shared/src/schema'

import extensions from '../../../code-intel-extensions.json'
import { isExtension } from '../context'

const DEFAULT_ENABLE_LEGACY_EXTENSIONS = true // Should be changed to false after Sourcegraph 4.0 release

/**
 * Determine if inline extensions should be loaded.
 *
 * This requires the browser extension to be built with inline extensions enabled.
 * At build time this is determined by `shouldBuildWithInlineExtensions`.
 */
export const shouldUseInlineExtensions = (requestGraphQL: PlatformContext['requestGraphQL']): Observable<boolean> =>
    requestGraphQL<GQL.IQuery>({
        request: gql`
            query PublicConfiguration {
                site {
                    publicConfiguration {
                        effectiveContents
                    }
                }
            }
        `,
        variables: {},
        mightContainPrivateInfo: false,
    }).pipe(
        map(result => {
            if (isErrorGraphQLResult(result)) {
                // PublicConfiguration query resolver may not be implemented on older versions.
                // Return `true` by default.
                return true
            }

            try {
                // const parsedConfig = JSON.parse(result.data.site.publicConfiguration.effectiveContents)
                // return Boolean(parsedConfig?.experimentalFeatures?.enableLegacyExtensions)

                return DEFAULT_ENABLE_LEGACY_EXTENSIONS // Should be taken from site config after Sourcegraph 4.0 release (uncomment the above lines)
            } catch {
                return DEFAULT_ENABLE_LEGACY_EXTENSIONS
            }
        }),
        map(enableLegacyExtensions => !enableLegacyExtensions && isExtension)
    )

/**
 * Get the manifest URL and script URL for a Sourcegraph extension which is inline (bundled with the browser add-on).
 */
function getURLsForInlineExtension(extensionID: string): { manifestURL: string; scriptURL: string } {
    const kebabCaseExtensionID = extensionID.replace(/^sourcegraph\//, 'sourcegraph-')

    return {
        manifestURL: browser.extension.getURL(`extensions/${kebabCaseExtensionID}/package.json`),
        scriptURL: browser.extension.getURL(`extensions/${kebabCaseExtensionID}/extension.js`),
    }
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
