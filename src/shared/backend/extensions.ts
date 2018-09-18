import { UpdateExtensionSettingsArgs } from '@sourcegraph/extensions-client-common/lib/context'
import { Controller as ExtensionsContextController } from '@sourcegraph/extensions-client-common/lib/controller'
import { ConfiguredExtension } from '@sourcegraph/extensions-client-common/lib/extensions/extension'
import { gql, graphQLContent } from '@sourcegraph/extensions-client-common/lib/graphql'
import {
    ConfigurationCascade,
    ConfigurationCascadeOrError,
    ConfigurationSubject,
    gqlToCascade,
    mergeSettings,
    Settings,
} from '@sourcegraph/extensions-client-common/lib/settings'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { applyEdits } from '@sqs/jsonc-parser'
import * as JSONC from '@sqs/jsonc-parser'
import { removeProperty, setProperty } from '@sqs/jsonc-parser/lib/edit'
import { isEqual } from 'lodash'
import Alert from 'mdi-react/AlertIcon'
import MenuDown from 'mdi-react/MenuDownIcon'
import Menu from 'mdi-react/MenuIcon'
import { combineLatest, from, Observable, throwError } from 'rxjs'
import { distinctUntilChanged, map, mergeMap, switchMap, take } from 'rxjs/operators'
import { MessageTransports } from 'sourcegraph/module/protocol/jsonrpc2/connection'
import { TextDocumentDecoration } from 'sourcegraph/module/protocol/plainTypes'
import uuid from 'uuid'
import { Disposable } from 'vscode-languageserver'
import storage, { StorageItems } from '../../browser/storage'
import { ExtensionConnectionInfo, onFirstMessage } from '../messaging'
import { getContext } from './context'
import { createAggregateError, isErrorLike } from './errors'
import { queryGraphQL } from './graphql'
import { sendLSPHTTPRequests } from './lsp'
import { createPortMessageTransports } from './PortMessageTransports'

const createPlatformMessageTransports = (connectionInfo: ExtensionConnectionInfo) =>
    new Promise<MessageTransports>((resolve, reject) => {
        const channelID = uuid.v4()
        const port = chrome.runtime.connect({ name: channelID })
        port.postMessage(connectionInfo)
        onFirstMessage(port, (response: { error?: any }) => {
            if (response.error) {
                reject(response.error)
            } else {
                resolve(createPortMessageTransports(port, connectionInfo))
            }
        })
    })

export function createMessageTransports(
    extension: Pick<ConfiguredExtension, 'id' | 'manifest'>
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
    return createPlatformMessageTransports({
        extensionID: extension.id,
        jsBundleURL: extension.manifest.url,
    }).catch(err => {
        console.error('Error connecting to', extension.id + ':', err)
        throw err
    })
}

const mergeDisposables = (...disposables: Disposable[]): Disposable => ({
    dispose: () => {
        for (const disposable of disposables) {
            disposable.dispose()
        }
    },
})

// This applies a decoration to a GitHub blob page. This doesn't work with any other code host yet.
export const applyDecoration = ({
    fileElement,
    decoration,
}: {
    fileElement: HTMLElement
    decoration: TextDocumentDecoration
}): Disposable => {
    const disposables: Disposable[] = []
    const ghLineNumber = decoration.range.start.line + 1
    const lineNumberElements: NodeListOf<HTMLElement> = fileElement.querySelectorAll(
        `td[data-line-number="${ghLineNumber}"]`
    )
    if (!lineNumberElements) {
        throw new Error(`Line number ${ghLineNumber} not found`)
    }
    if (lineNumberElements.length !== 1) {
        throw new Error(`Line number ${ghLineNumber} matched ${lineNumberElements.length} elements (expected 1)`)
    }
    const lineNumberElement = lineNumberElements[0]
    if (!lineNumberElement) {
        throw new Error(`Line number ${ghLineNumber} is falsy: ${lineNumberElement}`)
    }
    const lineElement = lineNumberElement.nextElementSibling as HTMLElement | undefined
    if (!lineElement) {
        throw new Error(`Line ${ghLineNumber} is falsy: ${lineNumberElement}`)
    }
    if (decoration.backgroundColor) {
        lineElement.style.backgroundColor = decoration.backgroundColor
        disposables.push({
            dispose: () => {
                lineElement.style.backgroundColor = null
            },
        })
    }
    if (decoration.after) {
        const linkTo = (url: string) => (e: HTMLElement): HTMLElement => {
            const link = document.createElement('a')
            link.setAttribute('href', url)
            link.style.color = decoration.after!.color || null
            link.appendChild(e)
            return link
        }
        const after = document.createElement('span')
        after.style.backgroundColor = decoration.after.backgroundColor || null
        after.textContent = decoration.after.contentText || null
        const annotation = decoration.after.linkURL ? linkTo(decoration.after.linkURL)(after) : after
        lineElement.appendChild(annotation)
        disposables.push({
            dispose: () => {
                annotation.remove()
            },
        })
    }
    return mergeDisposables(...disposables)
}

const storageConfigurationCascade: Observable<
    ConfigurationCascade<ConfigurationSubject, Settings>
> = storage.observeSync('clientSettings').pipe(
    map(clientSettingsString => JSONC.parse(clientSettingsString || '')),
    map(clientSettings => ({
        subjects: [
            {
                subject: {
                    id: 'Client',
                    settingsURL: 'N/A',
                    viewerCanAdminister: true,
                    __typename: 'Client',
                    displayName: 'Client',
                } as ConfigurationSubject,
                settings: clientSettings,
            },
        ],
        merged: clientSettings || {},
    }))
)

const mergeCascades = (
    cascadeOrError: ConfigurationCascadeOrError<ConfigurationSubject, Settings>,
    cascade: ConfigurationCascade<ConfigurationSubject, Settings>
): ConfigurationCascadeOrError<ConfigurationSubject, Settings> => ({
    subjects:
        cascadeOrError.subjects === null
            ? cascade.subjects
            : isErrorLike(cascadeOrError.subjects)
                ? cascadeOrError.subjects
                : [...cascadeOrError.subjects, ...cascade.subjects],
    merged:
        cascadeOrError.merged === null
            ? cascade.merged
            : isErrorLike(cascadeOrError.merged)
                ? cascadeOrError.merged
                : mergeSettings([cascadeOrError.merged, cascade.merged]),
})

const configurationCascadeFragment = gql`
    fragment ConfigurationCascadeFields on ConfigurationCascade {
        subjects {
            __typename
            ... on Org {
                id
                name
                displayName
            }
            ... on User {
                id
                username
                displayName
            }
            ... on Site {
                id
                siteID
            }
            latestSettings {
                id
                configuration {
                    contents
                }
            }
            settingsURL
            viewerCanAdminister
        }
        merged {
            contents
            messages
        }
    }
`

/**
 * Always represents the entire configuration cascade; i.e., it contains the
 * individual configs from the various config subjects (orgs, user, etc.).
 */
export const gqlConfigurationCascade = storage.observeSync('sourcegraphURL').pipe(
    switchMap(url =>
        queryGraphQL(
            getContext({ repoKey: '', isRepoSpecific: false }),
            gql`
                query Configuration {
                    viewerConfiguration {
                        ...ConfigurationCascadeFields
                    }
                }
                ${configurationCascadeFragment}
            `[graphQLContent],
            {},
            url
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.viewerConfiguration) {
                    throw createAggregateError(errors)
                }

                for (const subject of data.viewerConfiguration.subjects) {
                    // User/org/global settings cannot be edited from the
                    // browser extension (only client settings can).
                    subject.viewerCanAdminister = false
                }

                return data.viewerConfiguration
            })
        )
    )
)

export function createExtensionsContextController(
    sourcegraphUrl: string
): ExtensionsContextController<ConfigurationSubject, Settings> {
    const sourcegraphLanguageServerURL = new URL(sourcegraphUrl)
    sourcegraphLanguageServerURL.pathname = '.api/xlang'

    return new ExtensionsContextController<ConfigurationSubject, Settings>({
        configurationCascade: combineLatest(gqlConfigurationCascade, storageConfigurationCascade).pipe(
            map(([gqlCascade, storageCascade]) => mergeCascades(gqlToCascade(gqlCascade), storageCascade)),
            distinctUntilChanged((a, b) => isEqual(a, b))
        ),
        updateExtensionSettings,
        queryGraphQL: (request, variables) =>
            storage.observeSync('sourcegraphURL').pipe(
                take(1),
                mergeMap(url =>
                    queryGraphQL(getContext({ repoKey: '', isRepoSpecific: false }), request, variables, url)
                )
            ),
        queryLSP: requests => sendLSPHTTPRequests(requests),
        icons: {
            Loader: LoadingSpinner as React.ComponentType<{ className: string; onClick?: () => void }>,
            Warning: Alert as React.ComponentType<{ className: string; onClick?: () => void }>,
            CaretDown: MenuDown as React.ComponentType<{ className: string; onClick?: () => void }>,
            Menu: Menu as React.ComponentType<{ className: string; onClick?: () => void }>,
        },
        forceUpdateTooltip: () => {
            // TODO(sqs): implement tooltips on the browser extension
        },
        experimentalClientCapabilities: {
            sourcegraphLanguageServerURL: sourcegraphLanguageServerURL.href,
        },
    })
}

export const updateExtensionSettings = (subjectID, args: UpdateExtensionSettingsArgs): Observable<undefined> => {
    if (subjectID !== 'Client') {
        return throwError('Cannot update settings for ' + subjectID + '.')
    }
    return from(
        new Promise<StorageItems>(resolve => storage.getSync(storageItems => resolve(storageItems))).then(
            storageItems => {
                let clientSettings = storageItems.clientSettings

                const format = { tabSize: 2, insertSpaces: true, eol: '\n' }

                if ('edit' in args && args.edit) {
                    clientSettings = applyEdits(
                        clientSettings,
                        // TODO(chris): remove `.slice()` (which guards against
                        // mutation) once
                        // https://github.com/Microsoft/node-jsonc-parser/pull/12
                        // is merged in.
                        setProperty(clientSettings, args.edit.path.slice(), args.edit.value, format)
                    )
                } else if ('extensionID' in args) {
                    clientSettings = applyEdits(
                        clientSettings,
                        typeof args.enabled === 'boolean'
                            ? setProperty(clientSettings, ['extensions', args.extensionID], args.enabled, format)
                            : removeProperty(clientSettings, ['extensions', args.extensionID], format)
                    )
                }
                return new Promise<undefined>(resolve =>
                    storage.setSync({ clientSettings }, () => {
                        resolve(undefined)
                    })
                )
            }
        )
    )
}
