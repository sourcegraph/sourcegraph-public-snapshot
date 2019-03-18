import { ControllerProps as GenericExtensionsControllerProps } from '@sourcegraph/extensions-client-common/lib/client/controller'
import {
    ExtensionsProps as GenericExtensionsProps,
    UpdateExtensionSettingsArgs,
} from '@sourcegraph/extensions-client-common/lib/context'
import { Controller as ExtensionsContextController } from '@sourcegraph/extensions-client-common/lib/controller'
import { ConfiguredExtension } from '@sourcegraph/extensions-client-common/lib/extensions/extension'
import { QueryResult } from '@sourcegraph/extensions-client-common/lib/graphql'
import { ClientConnection } from '@sourcegraph/extensions-client-common/lib/messaging'
import * as ECCGQL from '@sourcegraph/extensions-client-common/lib/schema/graphqlschema'
import {
    ConfigurationCascadeProps as GenericConfigurationCascadeProps,
    ConfigurationSubject,
    gqlToCascade,
    Settings,
} from '@sourcegraph/extensions-client-common/lib/settings'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { isEqual } from 'lodash'
import AddIcon from 'mdi-react/AddIcon'
import InfoIcon from 'mdi-react/InformationIcon'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import MenuIcon from 'mdi-react/MenuIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import WarningIcon from 'mdi-react/WarningIcon'
import { concat, from, Observable } from 'rxjs'
import { distinctUntilChanged, map, mapTo, switchMap, take } from 'rxjs/operators'
import { InitData } from 'sourcegraph/module/extension/extensionHost'
import { MessageTransports } from 'sourcegraph/module/protocol/jsonrpc2/connection'
import { createWebWorkerMessageTransports } from 'sourcegraph/module/protocol/jsonrpc2/transports/webWorker'
import ExtensionHostWorker from 'worker-loader!./extensionHost.worker'
import { gql, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { sendLSPHTTPRequests } from '../backend/lsp'
import { Tooltip } from '../components/tooltip/Tooltip'
import { editConfiguration } from '../configuration/backend'
import { configurationCascade, toGQLKeyPath } from '../settings/configuration'
import { refreshConfiguration } from '../user/settings/backend'
import { isErrorLike } from '../util/errors'

export interface ExtensionsControllerProps extends GenericExtensionsControllerProps<ConfigurationSubject, Settings> {}

export interface ConfigurationCascadeProps extends GenericConfigurationCascadeProps<ConfigurationSubject, Settings> {}

export interface ExtensionsProps extends GenericExtensionsProps<ConfigurationSubject, Settings> {}

export function createExtensionsContextController(
    clientConnection: Promise<ClientConnection>
): ExtensionsContextController<ConfigurationSubject, Settings> {
    return new ExtensionsContextController<ConfigurationSubject, Settings>({
        configurationCascade: configurationCascade.pipe(
            map(gqlToCascade),
            distinctUntilChanged((a, b) => isEqual(a, b))
        ),
        updateExtensionSettings: (subject, args) => updateExtensionSettings(subject, args, clientConnection),
        queryGraphQL: (request, variables) =>
            queryGraphQL(
                gql`
                    ${request}
                `,
                variables
            ) as Observable<QueryResult<Pick<ECCGQL.IQuery, 'extensionRegistry'>>>,
        queryLSP: requests => sendLSPHTTPRequests(requests),
        icons: {
            Loader: LoadingSpinner as React.ComponentType<{ className: string; onClick?: () => void }>,
            Warning: WarningIcon as React.ComponentType<{ className: string; onClick?: () => void }>,
            Info: InfoIcon as React.ComponentType<{ className: string; onClick?: () => void }>,
            CaretDown: MenuDownIcon as React.ComponentType<{ className: string; onClick?: () => void }>,
            Menu: MenuIcon as React.ComponentType<{ className: string; onClick?: () => void }>,
            Add: AddIcon as React.ComponentType<{ className: string; onClick?: () => void }>,
            Settings: SettingsIcon as React.ComponentType<{ className: string; onClick?: () => void }>,
        },
        forceUpdateTooltip: () => Tooltip.forceUpdate(),
    })
}

function updateExtensionSettings(
    subject: string,
    args: UpdateExtensionSettingsArgs,
    clientConnection: Promise<ClientConnection>
): Observable<void> {
    return configurationCascade.pipe(
        take(1),
        switchMap(configurationCascade => {
            const subjectConfig = configurationCascade.subjects.find(s => s.id === subject)
            if (!subjectConfig) {
                throw new Error(`no configuration subject: ${subject}`)
            }
            const lastID = subjectConfig.latestSettings ? subjectConfig.latestSettings.id : null

            let edit: GQL.IConfigurationEdit
            if ('edit' in args && args.edit) {
                edit = { keyPath: toGQLKeyPath(args.edit.path), value: args.edit.value }
            } else if ('extensionID' in args) {
                edit = {
                    keyPath: toGQLKeyPath(['extensions', args.extensionID]),
                    value: typeof args.enabled === 'boolean' ? args.enabled : null,
                }
            } else {
                throw new Error('no edit')
            }

            if (subject === 'Client') {
                return from(clientConnection.then(connection => connection.editSetting(args))).pipe(mapTo(undefined))
            }

            return editConfiguration(subject, lastID, edit)
        }),
        switchMap(() => concat(refreshConfiguration(), [void 0]))
    )
}

export function updateHighestPrecedenceExtensionSettings(
    args: {
        extensionID: string
        enabled?: boolean
    },
    clientConnection: Promise<ClientConnection>
): Observable<void> {
    return configurationCascade.pipe(
        take(1),
        switchMap(configurationCascade => {
            // Only support configuring extension settings in user settings with this action.
            const subject = configurationCascade.subjects[configurationCascade.subjects.length - 1]
            return updateExtensionSettings(subject.id, args, clientConnection)
        })
    )
}

export function createMessageTransports(
    extension: Pick<ConfiguredExtension, 'id' | 'manifest'>
): Promise<MessageTransports> {
    if (!extension.manifest) {
        throw new Error(`unable to run extension ${JSON.stringify(extension.id)}: no manifest found`)
    }
    if (isErrorLike(extension.manifest)) {
        throw new Error(
            `unable to run extension ${JSON.stringify(extension.id)}: invalid manifest: ${extension.manifest.message}`
        )
    }

    if (extension.manifest.url) {
        const url = extension.manifest.url
        return fetch(url, { credentials: 'same-origin' })
            .then(resp => {
                if (resp.status !== 200) {
                    return resp
                        .text()
                        .then(text => Promise.reject(new Error(`loading bundle from ${url} failed: ${text}`)))
                }
                return resp.text()
            })
            .then(bundleSource => {
                const blobURL = window.URL.createObjectURL(
                    new Blob([bundleSource], {
                        type: 'application/javascript',
                    })
                )
                try {
                    const worker = new ExtensionHostWorker()
                    const initData: InitData = { bundleURL: blobURL }
                    worker.postMessage(initData)
                    return createWebWorkerMessageTransports(worker)
                } catch (err) {
                    console.error(err)
                }
                throw new Error('failed to initialize extension host')
            })
    }
    throw new Error(`unable to run extension ${JSON.stringify(extension.id)}: no "url" property in manifest`)
}
