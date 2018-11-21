import { isEqual } from 'lodash'
import { distinctUntilChanged, map } from 'rxjs/operators'
import ExtensionHostWorker from 'worker-loader!./extensionHost.worker'
import { InitData } from '../../../shared/src/api/extension/extensionHost'
import { SettingsCascade } from '../../../shared/src/api/protocol'
import { MessageTransports } from '../../../shared/src/api/protocol/jsonrpc2/connection'
import { createWebWorkerMessageTransports } from '../../../shared/src/api/protocol/jsonrpc2/transports/webWorker'
import { ControllerProps as GenericExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { ConfiguredExtension } from '../../../shared/src/extensions/extension'
import { gql } from '../../../shared/src/graphql/graphql'
import {
    PlatformContext,
    PlatformContextProps as GenericPlatformContextProps,
} from '../../../shared/src/platform/context'
import { mutateSettings, updateSettings } from '../../../shared/src/settings/edit'
import {
    gqlToCascade,
    Settings,
    SettingsCascadeProps as GenericSettingsCascadeProps,
} from '../../../shared/src/settings/settings'
import { ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { requestGraphQL } from '../backend/graphql'
import { sendLSPHTTPRequests } from '../backend/lsp'
import { Tooltip } from '../components/tooltip/Tooltip'
import { settingsCascade } from '../settings/configuration'
import { refreshSettings } from '../user/settings/backend'

export interface ExtensionsControllerProps extends GenericExtensionsControllerProps {}

export interface SettingsCascadeProps extends GenericSettingsCascadeProps {}

export interface ExtensionsProps extends GenericPlatformContextProps {}

export function createPlatformContext(): PlatformContext {
    const context: PlatformContext = {
        settingsCascade: settingsCascade.pipe(
            map(gqlToCascade),
            distinctUntilChanged((a, b) => isEqual(a, b))
        ),
        updateSettings: async (subject, args) => {
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
    }
    return context
}

export function createMessageTransports(
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

/** Reports whether the given extension is mentioned (enabled or disabled) in the settings. */
export function isExtensionAdded(settings: Settings | ErrorLike | null, extensionID: string): boolean {
    return !!settings && !isErrorLike(settings) && !!settings.extensions && extensionID in settings.extensions
}
