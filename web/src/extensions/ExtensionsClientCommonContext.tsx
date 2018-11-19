import { isEqual } from 'lodash'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import MenuIcon from 'mdi-react/MenuIcon'
import { concat, Observable } from 'rxjs'
import { distinctUntilChanged, map, switchMap, take, withLatestFrom } from 'rxjs/operators'
import ExtensionHostWorker from 'worker-loader!./extensionHost.worker'
import { InitData } from '../../../shared/src/api/extension/extensionHost'
import { SettingsCascade } from '../../../shared/src/api/protocol'
import { MessageTransports } from '../../../shared/src/api/protocol/jsonrpc2/connection'
import { createWebWorkerMessageTransports } from '../../../shared/src/api/protocol/jsonrpc2/transports/webWorker'
import { ControllerProps as GenericExtensionsControllerProps } from '../../../shared/src/client/controller'
import { ExtensionsProps as GenericExtensionsProps, UpdateExtensionSettingsArgs } from '../../../shared/src/context'
import { Controller as ExtensionsContextController } from '../../../shared/src/controller'
import { ConfiguredExtension } from '../../../shared/src/extensions/extension'
import { QueryResult } from '../../../shared/src/graphql'
import * as GQL from '../../../shared/src/graphqlschema'
import {
    gqlToCascade,
    Settings,
    SettingsCascadeProps as GenericSettingsCascadeProps,
    SettingsSubject,
} from '../../../shared/src/settings'
import { authenticatedUser } from '../auth'
import { gql, queryGraphQL } from '../backend/graphql'
import { sendLSPHTTPRequests } from '../backend/lsp'
import { Tooltip } from '../components/tooltip/Tooltip'
import { editSettings } from '../configuration/backend'
import { settingsCascade, toGQLKeyPath } from '../settings/configuration'
import { refreshSettings } from '../user/settings/backend'
import { ErrorLike, isErrorLike } from '../util/errors'

export interface ExtensionsControllerProps extends GenericExtensionsControllerProps<SettingsSubject, Settings> {}

export interface SettingsCascadeProps extends GenericSettingsCascadeProps<SettingsSubject, Settings> {}

export interface ExtensionsProps extends GenericExtensionsProps<SettingsSubject, Settings> {}

export function createExtensionsContextController(): ExtensionsContextController<SettingsSubject, Settings> {
    return new ExtensionsContextController<SettingsSubject, Settings>({
        settingsCascade: settingsCascade.pipe(
            map(gqlToCascade),
            distinctUntilChanged((a, b) => isEqual(a, b))
        ),
        updateExtensionSettings: (subject, args) => updateExtensionSettings(subject, args),
        queryGraphQL: (request, variables) =>
            queryGraphQL(
                gql`
                    ${request}
                `,
                variables
            ) as Observable<QueryResult<Pick<GQL.IQuery, 'extensionRegistry' | 'repository'>>>,
        queryLSP: requests => sendLSPHTTPRequests(requests),
        icons: {
            CaretDown: MenuDownIcon as React.ComponentType<{ className: string; onClick?: () => void }>,
            Menu: MenuIcon as React.ComponentType<{ className: string; onClick?: () => void }>,
        },
        forceUpdateTooltip: () => Tooltip.forceUpdate(),
    })
}

function updateExtensionSettings(subject: string, args: UpdateExtensionSettingsArgs): Observable<void> {
    return settingsCascade.pipe(
        take(1),
        withLatestFrom(authenticatedUser),
        switchMap(([settingsCascade, authenticatedUser]) => {
            const subjectSettings = settingsCascade.subjects.find(s => s.id === subject)
            if (!subjectSettings) {
                throw new Error(`no settings subject: ${subject}`)
            }
            const lastID = subjectSettings.latestSettings ? subjectSettings.latestSettings.id : null

            let edit: GQL.ISettingsEdit
            let editDescription: string
            if ('edit' in args && args.edit) {
                edit = { keyPath: toGQLKeyPath(args.edit.path), value: args.edit.value }
                editDescription = `update user setting ` + '`' + args.edit.path + '`'
            } else if ('extensionID' in args) {
                edit = {
                    keyPath: toGQLKeyPath(['extensions', args.extensionID]),
                    value: typeof args.enabled === 'boolean' ? args.enabled : null,
                }
                editDescription =
                    `${typeof args.enabled === 'boolean' ? 'enable' : 'disable'} extension ` +
                    '`' +
                    args.extensionID +
                    '`'
            } else {
                throw new Error('no edit')
            }

            if (!authenticatedUser) {
                const u = new URL(window.context.externalURL)
                throw new Error(
                    `Unable to ${editDescription} because you are not signed in.` +
                        '\n\n' +
                        `[**Sign into Sourcegraph${
                            u.hostname === 'sourcegraph.com' ? '' : ` on ${u.host}`
                        }**](${`${u.href.replace(/\/$/, '')}/sign-in`})`
                )
            }

            return editSettings(subject, lastID, edit)
        }),
        switchMap(() => concat(refreshSettings(), [void 0]))
    )
}

export function updateHighestPrecedenceExtensionSettings(args: {
    extensionID: string
    enabled?: boolean
}): Observable<void> {
    return settingsCascade.pipe(
        take(1),
        switchMap(settingsCascade => {
            // Only support configuring extension settings in user settings with this action.
            const subject = settingsCascade.subjects[settingsCascade.subjects.length - 1]
            return updateExtensionSettings(subject.id, args)
        })
    )
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
