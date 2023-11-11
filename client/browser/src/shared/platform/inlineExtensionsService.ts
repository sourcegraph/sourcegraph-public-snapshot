import { type Observable, from } from 'rxjs'

import { checkOk } from '@sourcegraph/http-client'
import type { ExecutableExtension } from '@sourcegraph/shared/src/api/extension/activation'
import type { ExtensionManifest } from '@sourcegraph/shared/src/extensions/extensionManifest'

import extensions from '../../../code-intel-extensions.json'

/**
 * Get the manifest URL and script URL for a Sourcegraph extension which is inline (bundled with the browser add-on).
 */
function getURLsForInlineExtension(extensionID: string, assetsURL: string): { manifestURL: string; scriptURL: string } {
    const kebabCaseExtensionID = extensionID.replace(/^sourcegraph\//, 'sourcegraph-')
    const manifestPath = `extensions/${kebabCaseExtensionID}/package.json`
    const scriptPath = `extensions/${kebabCaseExtensionID}/extension.js`

    /**
     * In a browser extension environment, we need to find the absolute URL in the browser extension assets directory.
     * We assume extension bundles exist (e.g. were built using [build-inline-extensions](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@a2190db3ee4b7e61f02f821cd2e5b58fa9814540/-/blob/client/browser/scripts/build-inline-extensions.js) script).
     */
    if (window.browser !== undefined) {
        return {
            manifestURL: browser.runtime.getURL(manifestPath),
            scriptURL: browser.runtime.getURL(scriptPath),
        }
    }

    /**
     * In a native integration environment we can use this URL as the root for bundled assets.
     * We assume that bundled extensions are available by that URL. For example,
     *
     * • GitLab native integration and its assets are hosted on a corresponding GitLab instance (see {@link https://sourcegraph.com/gitlab.com/gitlab-org/gitlab@a99a4b33543b01d05abd5fcd76b05298ae614fe4/-/blob/app/assets/javascripts/sourcegraph/index.js|sourcegraph/index.js} and {@link https://github.com/sourcegraph/sourcegraph/pull/42106|pull/42106})
     *
     * • Bitbucket Server integration and its assets are hosted on a connected Sourcegraph instance (see {@link https://sourcegraph.com/github.com/sourcegraph/bitbucket-server-plugin@97fc6d39f406b100359ee5e22cc3bee9d1453df8/-/blob/src/main/resources/js/sourcegraph-bitbucket.js|sourcegraph-bitbucket.js} and {@link copyIntegrationAssets})
     */
    if (assetsURL) {
        return {
            manifestURL: new URL(manifestPath, assetsURL).toString(),
            scriptURL: new URL(scriptPath, assetsURL).toString(),
        }
    }

    throw new Error('Can not construct extension URLs')
}

export function getInlineExtensions(assetsURL: string): Observable<ExecutableExtension[]> {
    const promises: Promise<ExecutableExtension>[] = []

    for (const extensionID of extensions) {
        const { manifestURL, scriptURL } = getURLsForInlineExtension(extensionID, assetsURL)
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
