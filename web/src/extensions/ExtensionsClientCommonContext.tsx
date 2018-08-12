import { ExtensionsProps as GenericExtensionsProps } from '@sourcegraph/extensions-client-common/lib/context'
import { Controller as ExtensionsContextController } from '@sourcegraph/extensions-client-common/lib/controller'
import { CXPControllerProps as GenericCXPControllerProps } from '@sourcegraph/extensions-client-common/lib/cxp/controller'
import { importScriptsBlobURL } from '@sourcegraph/extensions-client-common/lib/cxp/webWorker'
import { ConfiguredExtension } from '@sourcegraph/extensions-client-common/lib/extensions/extension'
import {
    ConfigurationCascadeProps as GenericConfigurationCascadeProps,
    ConfigurationSubject,
    gqlToCascade,
    Settings,
} from '@sourcegraph/extensions-client-common/lib/settings'
import CaretDown from '@sourcegraph/icons/lib/CaretDown'
import Loader from '@sourcegraph/icons/lib/Loader'
import Menu from '@sourcegraph/icons/lib/Menu'
import Warning from '@sourcegraph/icons/lib/Warning'
import { ClientOptions } from 'cxp/module/client/client'
import { MessageTransports } from 'cxp/module/jsonrpc2/connection'
import { createWebSocketMessageTransports } from 'cxp/module/jsonrpc2/transports/browserWebSocket'
import { createWebWorkerMessageTransports } from 'cxp/module/jsonrpc2/transports/webWorker'
import { ConfigurationUpdateParams } from 'cxp/module/protocol'
import { isEqual } from 'lodash'
import { concat, Observable } from 'rxjs'
import { distinctUntilChanged, map, switchMap, take } from 'rxjs/operators'
import { gql, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { Tooltip } from '../components/tooltip/Tooltip'
import { editConfiguration } from '../configuration/backend'
import { configurationCascade, toGQLKeyPath } from '../settings/configuration'
import { refreshConfiguration } from '../user/settings/backend'
import { isErrorLike } from '../util/errors'

export interface CXPControllerProps extends GenericCXPControllerProps<ConfigurationSubject, Settings> {}

export interface ConfigurationCascadeProps extends GenericConfigurationCascadeProps<ConfigurationSubject, Settings> {}

export interface ExtensionsProps extends GenericExtensionsProps<ConfigurationSubject, Settings> {}

export function createExtensionsContextController(): ExtensionsContextController<ConfigurationSubject, Settings> {
    return new ExtensionsContextController<ConfigurationSubject, Settings>({
        configurationCascade: configurationCascade.pipe(
            map(gqlToCascade),
            distinctUntilChanged((a, b) => isEqual(a, b))
        ),
        updateExtensionSettings,
        queryGraphQL: (request, variables) =>
            queryGraphQL(
                gql`
                    ${request}
                `,
                variables
            ),
        icons: {
            Loader: Loader as React.ComponentType<{ className: string; onClick?: () => void }>,
            Warning: Warning as React.ComponentType<{ className: string; onClick?: () => void }>,
            CaretDown: CaretDown as React.ComponentType<{ className: string; onClick?: () => void }>,
            Menu: Menu as React.ComponentType<{ className: string; onClick?: () => void }>,
        },
        forceUpdateTooltip: () => Tooltip.forceUpdate(),
    })
}

function updateExtensionSettings(
    subject: string,
    args: {
        extensionID: string
        edit?: ConfigurationUpdateParams
        enabled?: boolean
        remove?: boolean
    }
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
            if (args.edit) {
                edit = { keyPath: toGQLKeyPath(args.edit.path), value: args.edit.value }
            } else if (typeof args.enabled === 'boolean') {
                edit = { keyPath: toGQLKeyPath(['extensions', args.extensionID]), value: args.enabled }
            } else if (args.remove) {
                edit = { keyPath: toGQLKeyPath(['extensions', args.extensionID]), value: null }
            } else {
                throw new Error('no edit')
            }
            return editConfiguration(subject, lastID, edit)
        }),
        switchMap(() => concat(refreshConfiguration(), [void 0]))
    )
}

export function updateUserExtensionSettings(args: { extensionID: string; enabled?: boolean }): Observable<void> {
    return configurationCascade.pipe(
        take(1),
        switchMap(configurationCascade => {
            // Only support configuring extension settings in user settings with this action.
            const subject = configurationCascade.subjects[configurationCascade.subjects.length - 1]
            return updateExtensionSettings(subject.id, args)
        })
    )
}

export function createMessageTransports(
    extension: Pick<ConfiguredExtension, 'id' | 'manifest'>,
    options: ClientOptions
): Promise<MessageTransports> {
    if (!extension.manifest) {
        throw new Error(`unable to connect to extension ${JSON.stringify(extension.id)}: no manifest found`)
    }
    if (isErrorLike(extension.manifest)) {
        throw new Error(
            `unable to connect to extension ${JSON.stringify(extension.id)}: invalid manifest: ${
                extension.manifest.message
            }`
        )
    }
    if (extension.manifest.platform.type === 'bundle') {
        const APPLICATION_JSON_MIME_TYPE = 'application/json'
        if (
            typeof extension.manifest.platform.contentType === 'string' &&
            extension.manifest.platform.contentType !== APPLICATION_JSON_MIME_TYPE
        ) {
            // Until these are supported, prevent people from
            throw new Error(
                `unable to run extension ${JSON.stringify(extension.id)} bundle: content type ${JSON.stringify(
                    extension.manifest.platform.contentType
                )} is not supported (use ${JSON.stringify(APPLICATION_JSON_MIME_TYPE)})`
            )
        }
        const worker = new Worker(importScriptsBlobURL(extension.id, extension.manifest.platform.url))
        return Promise.resolve(createWebWorkerMessageTransports(worker))
    }

    // Include ?mode=&repo= in the url to make it easier to find the correct WebSocket connection in (e.g.) the
    // Chrome network inspector. It does not affect any behaviour.
    const url = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/.api/lsp?mode=${
        extension.id
    }&rootUri=${options.root}`
    return createWebSocketMessageTransports(new WebSocket(url))
}
