import { Observable, from } from 'rxjs'

import { checkOk } from '@sourcegraph/http-client'
import { ExecutableExtension } from '@sourcegraph/shared/src/api/extension/activation'
import { ExtensionManifest } from '@sourcegraph/shared/src/extensions/extensionManifest'

import extensions from '../../../code-intel-extensions.json'

/**
 * Get the manifest URL and script URL for a Sourcegraph extension which is inline (bundled with the browser add-on).
 */
function getURLsForInlineExtension(
    extensionID: string,
    sourcegraphURL: string
): { manifestURL: string; scriptURL: string } {
    const kebabCaseExtensionID = extensionID.replace(/^sourcegraph\//, 'sourcegraph-')
    const manifestPath = `extensions/${kebabCaseExtensionID}/package.json`
    const scriptPath = `extensions/${kebabCaseExtensionID}/extension.js`

    /**
     * In a browser extension environment, we need to find the absolute ULR in the browser extension assets directory.
     * We assume extension bundles exist (e.g. were built using [build-inline-extensions](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@a2190db3ee4b7e61f02f821cd2e5b58fa9814540/-/blob/client/browser/scripts/build-inline-extensions.js) script).
     */
    if (typeof window.browser !== 'undefined') {
        return {
            manifestURL: browser.runtime.getURL(manifestPath),
            scriptURL: browser.runtime.getURL(scriptPath),
        }
    }

    /**
     * In a native integration environment if SOURCEGRAPH_ASSETS_URL is defined (e.g. GitLab, see [sourcegraph/index.js](https://sourcegraph.com/gitlab.com/gitlab-org/gitlab@a99a4b33543b01d05abd5fcd76b05298ae614fe4/-/blob/app/assets/javascripts/sourcegraph/index.js)),
     * we can use this URL as the root for bundled assets.
     */
    if (window.SOURCEGRAPH_ASSETS_URL) {
        return {
            manifestURL: new URL(manifestPath, window.SOURCEGRAPH_ASSETS_URL).toString(),
            scriptURL: new URL(scriptPath, window.SOURCEGRAPH_ASSETS_URL).toString(),
        }
    }

    /**
     * In a native integration environment if SOURCEGRAPH_ASSETS_URL is not defined,
     * we assume that native integration has code-intel extensions bundled during the {@link copyIntegrationAssets} step.
     * (e.g. BitBucket, see [sourcegraph-bitbucket.js](https://sourcegraph.com/github.com/sourcegraph/bitbucket-server-plugin@97fc6d39f406b100359ee5e22cc3bee9d1453df8/-/blob/src/main/resources/js/sourcegraph-bitbucket.js)).
     */
    return {
        manifestURL: new URL(`.assets/extension/${manifestPath}`, sourcegraphURL).toString(),
        scriptURL: new URL(`.assets/extension/${scriptPath}`, sourcegraphURL).toString(),
    }
}

export function getInlineExtensions(sourcegraphURL: string): Observable<ExecutableExtension[]> {
    const promises: Promise<ExecutableExtension>[] = []

    for (const extensionID of extensions) {
        const { manifestURL, scriptURL } = getURLsForInlineExtension(extensionID, sourcegraphURL)
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
