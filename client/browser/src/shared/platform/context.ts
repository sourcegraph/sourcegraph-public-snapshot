import { combineLatest, ReplaySubject } from 'rxjs'
import { map } from 'rxjs/operators'

import { isHTTPAuthError } from '@sourcegraph/shared/src/backend/fetch'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { mutateSettings, updateSettings } from '@sourcegraph/shared/src/settings/edit'
import { EMPTY_SETTINGS_CASCADE, gqlToCascade } from '@sourcegraph/shared/src/settings/settings'
import { asError } from '@sourcegraph/shared/src/util/errors'
import { LocalStorageSubject } from '@sourcegraph/shared/src/util/LocalStorageSubject'
import { toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'

import { ExtensionStorageSubject } from '../../browser-extension/web-extension-api/ExtensionStorageSubject'
import { background } from '../../browser-extension/web-extension-api/runtime'
import { createGraphQLHelpers } from '../backend/requestGraphQl'
import { CodeHost } from '../code-hosts/shared/codeHost'
import { isInPage } from '../context'

import { createExtensionHost } from './extensionHost'
import { getInlineExtensions, shouldUseInlineExtensions } from './inlineExtensionsService'
import { editClientSettings, fetchViewerSettings, mergeCascades, storageSettingsCascade } from './settings'

export interface SourcegraphIntegrationURLs {
    /**
     * The URL of the configured Sourcegraph instance. Used for extensions, find-references, ...
     */
    sourcegraphURL: string

    /**
     * The base URL where assets will be fetched from (CSS, extension host
     * worker bundle, ...)
     *
     * This is the sourcegraph URL in most cases, but may be different for
     * native code hosts that self-host the integration bundle.
     */
    assetsURL: string
}

/**
 * The PlatformContext provided in the browser (for browser extensions and native integrations).
 */
export interface BrowserPlatformContext extends PlatformContext {
    /**
     * Re-fetches the settings cascade from the Sourcegraph instance.
     */
    refreshSettings(): Promise<void>
}

/**
 * Creates the {@link PlatformContext} for the browser (for browser extensions and native integrations)
 */
export function createPlatformContext(
    { urlToFile }: Pick<CodeHost, 'urlToFile'>,
    { sourcegraphURL, assetsURL }: SourcegraphIntegrationURLs,
    isExtension: boolean
): BrowserPlatformContext {
    const updatedViewerSettings = new ReplaySubject<Pick<GQL.ISettingsCascade, 'subjects' | 'final'>>(1)
    console.log('createPlatformContext', sourcegraphURL)
    const { requestGraphQL, getBrowserGraphQLClient } = createGraphQLHelpers(sourcegraphURL, isExtension)

    const context: BrowserPlatformContext = {
        /**
         * The active settings cascade.
         *
         * - For unauthenticated users, this is the GraphQL settings plus client settings (which are stored locally in the browser extension).
         * - For authenticated users, this is just the GraphQL settings (client settings are ignored to simplify the UX).
         */
        settings: combineLatest([updatedViewerSettings, storageSettingsCascade]).pipe(
            map(([gqlCascade, storageCascade]) =>
                gqlCascade
                    ? mergeCascades(
                          gqlToCascade(gqlCascade),
                          gqlCascade.subjects.some(subject => subject.__typename === 'User')
                              ? EMPTY_SETTINGS_CASCADE
                              : storageCascade
                      )
                    : EMPTY_SETTINGS_CASCADE
            )
        ),
        refreshSettings: async () => {
            try {
                const settings = await fetchViewerSettings(requestGraphQL).toPromise()
                updatedViewerSettings.next(settings)
            } catch (error) {
                if (isHTTPAuthError(error)) {
                    // User is not signed in
                    console.warn(
                        `Could not fetch Sourcegraph settings from ${sourcegraphURL} because user is not signed into Sourcegraph`
                    )
                    updatedViewerSettings.next({ final: '{}', subjects: [] })
                } else {
                    throw error
                }
            }
        },
        updateSettings: async (subject, edit) => {
            if (subject === 'Client') {
                // Support storing settings on the client (in the browser extension) so that unauthenticated
                // Sourcegraph viewers can update settings.
                await updateSettings(context, subject, edit, () => editClientSettings(edit))
                return
            }

            try {
                await updateSettings(context, subject, edit, mutateSettings)
            } catch (error) {
                if (asError(error).message.includes('version mismatch')) {
                    // The user probably edited the settings in another tab, so
                    // try once more.
                    await context.refreshSettings()
                    await updateSettings(context, subject, edit, mutateSettings)
                } else {
                    throw error
                }
            }
            // TODO: We shouldn't need to make another HTTP request to get the latest state
            await context.refreshSettings()
        },
        requestGraphQL,
        getGraphQLClient: getBrowserGraphQLClient,
        forceUpdateTooltip: () => {
            // TODO(sqs): implement tooltips on the browser extension
        },
        createExtensionHost: () => createExtensionHost({ assetsURL }),
        getScriptURLForExtension: () => {
            if (isInPage || shouldUseInlineExtensions()) {
                // inline extensions have fixed scriptURLs
                return undefined
            }
            // We need to import the extension's JavaScript file (in importScripts in the Web Worker) from a blob:
            // URI, not its original http:/https: URL, because Chrome extensions are not allowed to be published
            // with a CSP that allowlists https://* in script-src (see
            // https://developer.chrome.com/extensions/contentSecurityPolicy#relaxing-remote-script). (Firefox
            // add-ons have an even stricter restriction.)
            return bundleURLs =>
                Promise.allSettled(bundleURLs.map(bundleURL => background.createBlobURL(bundleURL))).then(results =>
                    results.map(result => (result.status === 'rejected' ? asError(result.reason) : result.value))
                )
        },
        urlToFile: ({ rawRepoName, ...target }, context) => {
            // We don't always resolve the rawRepoName, e.g. if there are multiple definitions.
            // Construct URL to file on code host, if possible.
            if (rawRepoName && urlToFile) {
                return urlToFile(sourcegraphURL, { rawRepoName, ...target }, context)
            }
            // Otherwise fall back to linking to Sourcegraph (with an absolute URL).
            return `${sourcegraphURL}${toPrettyBlobURL(target)}`
        },
        sourcegraphURL,
        clientApplication: 'other',
        sideloadedExtensionURL: isInPage
            ? new LocalStorageSubject<string | null>('sideloadedExtensionURL', null)
            : new ExtensionStorageSubject('sideloadedExtensionURL', null),
        getStaticExtensions: () => {
            if (shouldUseInlineExtensions()) {
                return getInlineExtensions()
            }

            return undefined
        },
    }
    return context
}
