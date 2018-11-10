import { UpdateExtensionSettingsArgs } from '@sourcegraph/extensions-client-common/lib/context'
import { Controller as ExtensionsContextController } from '@sourcegraph/extensions-client-common/lib/controller'
import { ConfiguredExtension } from '@sourcegraph/extensions-client-common/lib/extensions/extension'
import { gql, graphQLContent } from '@sourcegraph/extensions-client-common/lib/graphql'
import {
    gqlToCascade,
    mergeSettings,
    Settings,
    SettingsCascade,
    SettingsCascadeOrError,
    SettingsSubject,
} from '@sourcegraph/extensions-client-common/lib/settings'
import { applyEdits } from '@sqs/jsonc-parser'
import * as JSONC from '@sqs/jsonc-parser'
import { removeProperty, setProperty } from '@sqs/jsonc-parser/lib/edit'
import { isEqual } from 'lodash'
import MenuDown from 'mdi-react/MenuDownIcon'
import Menu from 'mdi-react/MenuIcon'
import { combineLatest, from, Observable, Subject, throwError } from 'rxjs'
import { distinctUntilChanged, map, mapTo, mergeMap, startWith, switchMap, take, tap } from 'rxjs/operators'
import { MessageTransports } from 'sourcegraph/module/protocol/jsonrpc2/connection'
import { TextDocumentDecoration } from 'sourcegraph/module/protocol/plainTypes'
import uuid from 'uuid'
import { Disposable } from 'vscode-languageserver'
import storage, { StorageItems } from '../../browser/storage'
import { ExtensionConnectionInfo, onFirstMessage } from '../../messaging'
import { GQL } from '../../types/gqlschema'
import { canFetchForURL } from '../util/context'
import { getContext } from './context'
import { createAggregateError, isErrorLike } from './errors'
import { mutateGraphQL, queryGraphQL } from './graphql'
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
                resolve(createPortMessageTransports(port))
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

const storageSettingsCascade: Observable<SettingsCascade<SettingsSubject, Settings>> = storage
    .observeSync('clientSettings')
    .pipe(
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
                    } as SettingsSubject,
                    settings: clientSettings,
                },
            ],
            final: clientSettings || {},
        }))
    )

const mergeCascades = (
    cascadeOrError: SettingsCascadeOrError<SettingsSubject, Settings>,
    cascade: SettingsCascade<SettingsSubject, Settings>
): SettingsCascadeOrError<SettingsSubject, Settings> => ({
    subjects:
        cascadeOrError.subjects === null
            ? cascade.subjects
            : isErrorLike(cascadeOrError.subjects)
                ? cascadeOrError.subjects
                : [...cascadeOrError.subjects, ...cascade.subjects],
    final:
        cascadeOrError.final === null
            ? cascade.final
            : isErrorLike(cascadeOrError.final)
                ? cascadeOrError.final
                : mergeSettings([cascadeOrError.final, cascade.final]),
})

// This is a fragment on the DEPRECATED GraphQL API type ConfigurationCascade (not SettingsCascade) for backcompat.
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

/** A subject that emits whenever the settings cascade must be refreshed from the Sourcegraph instance. */
const settingsCascadeRefreshes = new Subject<void>()

/**
 * Always represents the entire settings cascade; i.e., it contains the individual settings from the various
 * settings subjects (orgs, user, etc.).
 *
 * TODO(sqs): This uses the DEPRECATED GraphQL Query.viewerConfiguration and ConfigurationCascade for backcompat.
 */
export const gqlSettingsCascade: Observable<Pick<GQL.ISettingsCascade, 'subjects' | 'final'>> = combineLatest(
    storage.observeSync('sourcegraphURL'),
    settingsCascadeRefreshes.pipe(
        mapTo(null),
        startWith(null)
    )
).pipe(
    switchMap(([url]) =>
        queryGraphQL({
            ctx: getContext({ repoKey: '', isRepoSpecific: false }),
            request: gql`
                query ViewerConfiguration {
                    viewerConfiguration {
                        ...ConfigurationCascadeFields
                    }
                }
                ${configurationCascadeFragment}
            `[graphQLContent],
            url,
            requestMightContainPrivateInfo: false,
        }).pipe(
            map(({ data, errors }) => {
                if (!data || !data.viewerConfiguration) {
                    throw createAggregateError(errors)
                }

                for (const subject of data.viewerConfiguration.subjects) {
                    // User/org/global settings cannot be edited from the
                    // browser extension (only client settings can).
                    subject.viewerCanAdminister = false
                }

                return {
                    subjects: data.viewerConfiguration.subjects,
                    final: data.viewerConfiguration.merged.contents,
                }
            })
        )
    )
)

const EMPTY_CONFIGURATION_CASCADE: SettingsCascade = { subjects: [], final: {} }

/**
 * The active settings cascade.
 *
 * - For unauthenticated users, this is the GraphQL settings plus client settings (which are stored locally in the
 *   browser extension.
 * - For authenticated users, this is just the GraphQL settings (client settings are ignored to simplify the UX).
 */
export const settingsCascade: Observable<SettingsCascadeOrError<SettingsSubject, Settings>> = combineLatest(
    gqlSettingsCascade,
    storageSettingsCascade
).pipe(
    map(([gqlCascade, storageCascade]) =>
        mergeCascades(
            gqlToCascade(gqlCascade),
            gqlCascade.subjects.some(subject => subject.__typename === 'User')
                ? EMPTY_CONFIGURATION_CASCADE
                : storageCascade
        )
    ),
    distinctUntilChanged((a, b) => isEqual(a, b))
)

export function createExtensionsContextController(
    sourcegraphUrl: string
): ExtensionsContextController<SettingsSubject, Settings> {
    const sourcegraphLanguageServerURL = new URL(sourcegraphUrl)
    sourcegraphLanguageServerURL.pathname = '.api/xlang'

    return new ExtensionsContextController<SettingsSubject, Settings>({
        settingsCascade,
        updateExtensionSettings,
        queryGraphQL: (request, variables, requestMightContainPrivateInfo) =>
            storage.observeSync('sourcegraphURL').pipe(
                take(1),
                mergeMap(url =>
                    queryGraphQL({
                        ctx: getContext({ repoKey: '', isRepoSpecific: false }),
                        request,
                        variables,
                        url,
                        requestMightContainPrivateInfo,
                    })
                )
            ),
        queryLSP: canFetchForURL(sourcegraphUrl)
            ? requests => sendLSPHTTPRequests(requests)
            : () =>
                  throwError(
                      'The queryLSP command is unavailable because the current repository does not exist on the Sourcegraph instance.'
                  ),
        icons: {
            CaretDown: MenuDown as React.ComponentType<{ className: string; onClick?: () => void }>,
            Menu: Menu as React.ComponentType<{ className: string; onClick?: () => void }>,
        },
        forceUpdateTooltip: () => {
            // TODO(sqs): implement tooltips on the browser extension
        },
    })
}

// TODO(sqs): copied from sourcegraph/sourcegraph temporarily
function updateUserSettings(subject: string, args: UpdateExtensionSettingsArgs): Observable<void> {
    return gqlSettingsCascade.pipe(
        take(1),
        switchMap(gqlSettingsCascade => {
            const subjectSettings = gqlSettingsCascade.subjects.find(s => s.id === subject)
            if (!subjectSettings) {
                throw new Error(`no settings subject: ${subject}`)
            }
            const lastID = subjectSettings.latestSettings ? subjectSettings.latestSettings.id : null

            let edit: GQL.ISettingsEdit
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

            return editSettings(subject, lastID, edit)
        })
    )
}

// TODO(sqs): copied from sourcegraph/sourcegraph temporarily
//
// NOTE: uses configurationMutation and editConfiguration for backcompat
function editSettings(subject: GQL.ID, lastID: number | null, edit: GQL.IConfigurationEdit): Observable<void> {
    return mutateGraphQL({
        ctx: getContext({ repoKey: '', isRepoSpecific: false }),
        request: `
            mutation EditConfiguration($subject: ID!, $lastID: Int, $edit: ConfigurationEdit!) {
                configurationMutation(input: { subject: $subject, lastID: $lastID }) {
                    editConfiguration(edit: $edit) {
                        empty {
                            alwaysNil
                        }
                    }
                }
            }
        `,
        variables: { subject, lastID, edit },
    }).pipe(
        map(({ errors }) => {
            if (errors && errors.length > 0) {
                throw createAggregateError(errors)
            }
        }),
        map(() => undefined),
        tap(() => settingsCascadeRefreshes.next())
    )
}

// TODO(sqs): copied from sourcegraph/sourcegraph temporarily
function toGQLKeyPath(keyPath: (string | number)[]): GQL.IKeyPathSegment[] {
    return keyPath.map(v => (typeof v === 'string' ? { property: v } : { index: v }))
}

const updateClientSettings = (subjectID: 'Client', args: UpdateExtensionSettingsArgs): Observable<void> =>
    from(
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

export const updateExtensionSettings = (subjectID: string, args: UpdateExtensionSettingsArgs): Observable<void> => {
    if (subjectID === 'Client') {
        return updateClientSettings(subjectID, args)
    }
    return updateUserSettings(subjectID, args)
}
