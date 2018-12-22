import { concat, Observable, ReplaySubject, throwError } from 'rxjs'
import { map, publishReplay, refCount } from 'rxjs/operators'
import ExtensionHostWorker from 'worker-loader!../../../shared/src/api/extension/main.worker.ts'
import { createWebWorkerMessageTransports } from '../../../shared/src/api/protocol/jsonrpc2/transports/webWorker'
import { gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { PlatformContext } from '../../../shared/src/platform/context'
import { mutateSettings, updateSettings } from '../../../shared/src/settings/edit'
import { gqlToCascade } from '../../../shared/src/settings/settings'
import { LocalStorageSubject } from '../../../shared/src/util/LocalStorageSubject'
import { toPrettyBlobURL } from '../../../shared/src/util/url'
import { requestGraphQL } from '../backend/graphql'
import { Tooltip } from '../components/tooltip/Tooltip'
import { fetchViewerSettings } from '../user/settings/backend'

/**
 * Creates the {@link PlatformContext} for the web app.
 */
export function createPlatformContext(): PlatformContext {
    const updatedSettings = new ReplaySubject<GQL.ISettingsCascade>(1)
    const context: PlatformContext = {
        settings: concat(fetchViewerSettings(), updatedSettings).pipe(
            map(gqlToCascade),
            publishReplay(1),
            refCount()
        ),
        updateSettings: async (subject, edit) => {
            // Unauthenticated users can't update settings. (In the browser extension, they can update client
            // settings even when not authenticated. The difference in behavior in the web app vs. browser
            // extension is why this logic lives here and not in shared/.)
            if (!window.context.isAuthenticatedUser) {
                const editDescription =
                    typeof edit === 'string' ? 'edit settings' : `update setting ` + '`' + edit.path.join('.') + '`'
                const u = new URL(window.context.externalURL)
                throw new Error(
                    `Unable to ${editDescription} because you are not signed in.` +
                        '\n\n' +
                        `[**Sign into Sourcegraph${
                            u.hostname === 'sourcegraph.com' ? '' : ` on ${u.host}`
                        }**](${`${u.href.replace(/\/$/, '')}/sign-in`})`
                )
            }

            await updateSettings(context, subject, edit, mutateSettings)
            updatedSettings.next(await fetchViewerSettings().toPromise())
        },
        queryGraphQL: (request, variables) =>
            requestGraphQL(
                gql`
                    ${request}
                `,
                variables
            ),
        backcompatQueryLSP: () =>
            throwError(
                'queryLSP is no longer implemented; extensions must manage their own connection to a language server'
            ),
        forceUpdateTooltip: () => Tooltip.forceUpdate(),
        createExtensionHost: () => {
            const worker = new ExtensionHostWorker()
            const messageTransports = createWebWorkerMessageTransports(worker)
            return new Observable(sub => {
                sub.next(messageTransports)
                return () => worker.terminate()
            })
        },
        urlToFile: toPrettyBlobURL,
        getScriptURLForExtension: bundleURL => bundleURL,
        sourcegraphURL: window.context.externalURL,
        clientApplication: 'sourcegraph',
        traceExtensionHostCommunication: new LocalStorageSubject<boolean>('traceExtensionHostCommunication', false),
    }
    return context
}
