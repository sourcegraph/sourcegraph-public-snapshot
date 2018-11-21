import { isEqual } from 'lodash'
import { distinctUntilChanged, map } from 'rxjs/operators'
import ExtensionHostWorker from 'worker-loader!./extensionHost.worker'
import { InitData } from '../../../shared/src/api/extension/extensionHost'
import { SettingsCascade } from '../../../shared/src/api/protocol'
import { MessageTransports } from '../../../shared/src/api/protocol/jsonrpc2/connection'
import { createWebWorkerMessageTransports } from '../../../shared/src/api/protocol/jsonrpc2/transports/webWorker'
import { ConfiguredExtension } from '../../../shared/src/extensions/extension'
import { gql } from '../../../shared/src/graphql/graphql'
import { PlatformContext } from '../../../shared/src/platform/context'
import { mutateSettings, updateSettings } from '../../../shared/src/settings/edit'
import { gqlToCascade } from '../../../shared/src/settings/settings'
import { isErrorLike } from '../../../shared/src/util/errors'
import { requestGraphQL } from '../backend/graphql'
import { sendLSPHTTPRequests } from '../backend/lsp'
import { Tooltip } from '../components/tooltip/Tooltip'
import { settingsCascade } from '../settings/configuration'
import { refreshSettings } from '../user/settings/backend'

/**
 * Creates the {@link PlatformContext} for the web app.
 */
export function createPlatformContext(): PlatformContext {
    const context: PlatformContext = {
        settingsCascade: settingsCascade.pipe(
            map(gqlToCascade),
            distinctUntilChanged((a, b) => isEqual(a, b))
        ),
        updateSettings: async (subject, args) => {
            // Unauthenticated users can't update settings. (In the browser extension, they can update client
            // settings even when not authenticated. The difference in behavior in the web app vs. browser
            // extension is why this logic lives here and not in shared/.)
            if (!window.context.isAuthenticatedUser) {
                let editDescription = 'edit settings' // default description
                if ('edit' in args && args.edit) {
                    editDescription = `update user setting ` + '`' + args.edit.path + '`'
                } else if ('extensionID' in args) {
                    editDescription =
                        `${typeof args.enabled === 'boolean' ? 'enable' : 'disable'} extension ` +
                        '`' +
                        args.extensionID +
                        '`'
                }
                const u = new URL(window.context.externalURL)
                throw new Error(
                    `Unable to ${editDescription} because you are not signed in.` +
                        '\n\n' +
                        `[**Sign into Sourcegraph${
                            u.hostname === 'sourcegraph.com' ? '' : ` on ${u.host}`
                        }**](${`${u.href.replace(/\/$/, '')}/sign-in`})`
                )
            }

            try {
                await updateSettings(context, subject, args, mutateSettings)
            } finally {
                await refreshSettings().toPromise()
            }
        },
        queryGraphQL: (request, variables) =>
            requestGraphQL(
                gql`
                    ${request}
                `,
                variables
            ),
        queryLSP: requests => sendLSPHTTPRequests(requests),
        forceUpdateTooltip: () => Tooltip.forceUpdate(),
        createMessageTransports: (extension, settingsCascade) =>
            Promise.resolve(createMessageTransports(extension, settingsCascade)),
    }
    return context
}

function createMessageTransports(
    extension: Pick<ConfiguredExtension, 'id' | 'manifest'>,
    settingsCascade: SettingsCascade<any>
): MessageTransports {
    if (!extension.manifest) {
        throw new Error(`unable to run extension ${JSON.stringify(extension.id)}: no manifest found`)
    }
    if (isErrorLike(extension.manifest)) {
        throw new Error(
            `unable to run extension ${JSON.stringify(extension.id)}: invalid manifest: ${extension.manifest.message}`
        )
    }

    if (extension.manifest.url) {
        try {
            const worker = new ExtensionHostWorker()
            const initData: InitData = {
                bundleURL: extension.manifest.url,
                sourcegraphURL: window.context.externalURL,
                clientApplication: 'sourcegraph',
                settingsCascade,
            }
            worker.postMessage(initData)
            return createWebWorkerMessageTransports(worker)
        } catch (err) {
            console.error(err)
        }
        throw new Error('failed to initialize extension host')
    }
    throw new Error(`unable to run extension ${JSON.stringify(extension.id)}: no "url" property in manifest`)
}
