import { combineLatest, merge, ReplaySubject, throwError } from 'rxjs'
import { map, mergeMap, publishReplay, refCount, switchMap, take } from 'rxjs/operators'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PlatformContext } from '../../../../shared/src/platform/context'
import { mutateSettings, updateSettings } from '../../../../shared/src/settings/edit'
import { EMPTY_SETTINGS_CASCADE, gqlToCascade } from '../../../../shared/src/settings/settings'
import { LocalStorageSubject } from '../../../../shared/src/util/LocalStorageSubject'
import { toPrettyBlobURL } from '../../../../shared/src/util/url'
import * as runtime from '../browser/runtime'
import storage from '../browser/storage'
import { CodeHost } from '../libs/code_intelligence'
import { getContext } from '../shared/backend/context'
import { requestGraphQL } from '../shared/backend/graphql'
import { sendLSPHTTPRequests } from '../shared/backend/lsp'
import { canFetchForURL, sourcegraphUrl } from '../shared/util/context'
import { createExtensionHost } from './extensionHost'
import { editClientSettings, fetchViewerSettings, mergeCascades, storageSettingsCascade } from './settings'

/**
 * Creates the {@link PlatformContext} for the browser extension.
 */
export function createPlatformContext({ urlToFile }: Pick<CodeHost, 'urlToFile'>): PlatformContext {
    // TODO: support listening for changes to sourcegraphUrl
    const sourcegraphLanguageServerURL = new URL(sourcegraphUrl)
    sourcegraphLanguageServerURL.pathname = '.api/xlang'

    const updatedViewerSettings = new ReplaySubject<Pick<GQL.ISettingsCascade, 'subjects' | 'final'>>(1)

    const context: PlatformContext = {
        /**
         * The active settings cascade.
         *
         * - For unauthenticated users, this is the GraphQL settings plus client settings (which are stored locally
         *   in the browser extension.
         * - For authenticated users, this is just the GraphQL settings (client settings are ignored to simplify
         *   the UX).
         */
        settings: combineLatest(
            merge(
                storage.observeSync('sourcegraphURL').pipe(switchMap(() => fetchViewerSettings())),
                updatedViewerSettings
            ).pipe(
                publishReplay(1),
                refCount()
            ),
            storageSettingsCascade
        ).pipe(
            map(([gqlCascade, storageCascade]) =>
                mergeCascades(
                    gqlToCascade(gqlCascade),
                    gqlCascade.subjects.some(subject => subject.__typename === 'User')
                        ? EMPTY_SETTINGS_CASCADE
                        : storageCascade
                )
            )
        ),

        updateSettings: async (subject, edit) => {
            await updateSettings(
                context,
                subject,
                edit,
                // Support storing settings on the client (in the browser extension) so that unauthenticated
                // Sourcegraph viewers can update settings.
                subject === 'Client' ? () => editClientSettings(edit) : mutateSettings
            )
            if (subject !== 'Client') {
                updatedViewerSettings.next(await fetchViewerSettings().toPromise())
            }
        },
        queryGraphQL: (request, variables, requestMightContainPrivateInfo) =>
            storage.observeSync('sourcegraphURL').pipe(
                take(1),
                mergeMap(url =>
                    requestGraphQL({
                        ctx: getContext({ repoKey: '', isRepoSpecific: false }),
                        request,
                        variables,
                        url,
                        requestMightContainPrivateInfo,
                    })
                )
            ),
        backcompatQueryLSP: canFetchForURL(sourcegraphUrl)
            ? requests => sendLSPHTTPRequests(requests)
            : () =>
                  throwError(
                      'The queryLSP command is unavailable because the current repository does not exist on the Sourcegraph instance.'
                  ),
        forceUpdateTooltip: () => {
            // TODO(sqs): implement tooltips on the browser extension
        },
        createExtensionHost,
        getScriptURLForExtension: async bundleURL => {
            // We need to import the extension's JavaScript file (in importScripts in the Web Worker) from a blob:
            // URI, not its original http:/https: URL, because Chrome extensions are not allowed to be published
            // with a CSP that allowlists https://* in script-src (see
            // https://developer.chrome.com/extensions/contentSecurityPolicy#relaxing-remote-script). (Firefox
            // add-ons have an even stricter restriction.)
            const blobURL = await new Promise<string>(resolve =>
                runtime.sendMessage(
                    {
                        type: 'createBlobURL',
                        payload: bundleURL,
                    },
                    resolve
                )
            )
            return blobURL
        },
        urlToFile: location => {
            if (urlToFile) {
                // Construct URL to file on code host, if possible.
                return urlToFile(location)
            }
            // Otherwise fall back to linking to Sourcegraph (with an absolute URL).
            return `${sourcegraphUrl}${toPrettyBlobURL(location)}`
        },
        sourcegraphURL: sourcegraphUrl,
        clientApplication: 'other',
        traceExtensionHostCommunication: new LocalStorageSubject<boolean>('traceExtensionHostCommunication', false),
    }
    return context
}
