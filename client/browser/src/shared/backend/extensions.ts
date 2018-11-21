import * as JSONC from '@sqs/jsonc-parser'
import { applyEdits } from '@sqs/jsonc-parser'
import { removeProperty, setProperty } from '@sqs/jsonc-parser/lib/edit'
import { isEqual } from 'lodash'
import { combineLatest, Observable, Subject, throwError, Unsubscribable } from 'rxjs'
import { distinctUntilChanged, map, mapTo, mergeMap, startWith, switchMap, take } from 'rxjs/operators'
import uuid from 'uuid'
import { MessageTransports } from '../../../../../shared/src/api/protocol/jsonrpc2/connection'
import { TextDocumentDecoration } from '../../../../../shared/src/api/protocol/plainTypes'
import { ConfiguredExtension } from '../../../../../shared/src/extensions/extension'
import { gql, graphQLContent } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContext } from '../../../../../shared/src/platform/context'
import { mutateSettings, UpdateExtensionSettingsArgs, updateSettings } from '../../../../../shared/src/settings/edit'
import {
    gqlToCascade,
    mergeSettings,
    SettingsCascade,
    SettingsCascadeOrError,
    SettingsSubject,
} from '../../../../../shared/src/settings/settings'
import storage, { StorageItems } from '../../browser/storage'
import { ExtensionConnectionInfo, onFirstMessage } from '../../messaging'
import { canFetchForURL } from '../util/context'
import { getContext } from './context'
import { createAggregateError, isErrorLike } from './errors'
import { queryGraphQL, requestGraphQL } from './graphql'
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
    extension: Pick<ConfiguredExtension, 'id' | 'manifest'>,
    settingsCascade: SettingsCascade
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
        settingsCascade,
    }).catch(err => {
        console.error('Error connecting to', extension.id + ':', err)
        throw err
    })
}

const combineUnsubscribables = (...unsubscribables: Unsubscribable[]): Unsubscribable => ({
    unsubscribe: () => {
        for (const unsubscribable of unsubscribables) {
            unsubscribable.unsubscribe()
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
}): Unsubscribable => {
    const unsubscribables: Unsubscribable[] = []
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
        unsubscribables.push({
            unsubscribe: () => {
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
        unsubscribables.push({
            unsubscribe: () => {
                annotation.remove()
            },
        })
    }
    return combineUnsubscribables(...unsubscribables)
}

const storageSettingsCascade: Observable<SettingsCascade> = storage.observeSync('clientSettings').pipe(
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
                lastID: null,
            },
        ],
        final: clientSettings || {},
    }))
)

const mergeCascades = (cascadeOrError: SettingsCascadeOrError, cascade: SettingsCascade): SettingsCascadeOrError => ({
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
                contents
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
            retry: false,
        }).pipe(
            map(({ data, errors }) => {
                // Suppress deprecation warnings because our use of these deprecated fields is intentional (see
                // tsdoc comment).
                //
                // tslint:disable deprecation
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
                // tslint:enable deprecation
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
export const settingsCascade: Observable<SettingsCascadeOrError> = combineLatest(
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

export function createPlatformContext(sourcegraphUrl: string): PlatformContext {
    const sourcegraphLanguageServerURL = new URL(sourcegraphUrl)
    sourcegraphLanguageServerURL.pathname = '.api/xlang'

    const context: PlatformContext = {
        settingsCascade,
        updateSettings: async (subject, args) => {
            await updateSettings(
                context,
                subject,
                args,
                subject === 'Client' ? () => editClientSettings(args) : mutateSettings
            )
            settingsCascadeRefreshes.next()
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
        queryLSP: canFetchForURL(sourcegraphUrl)
            ? requests => sendLSPHTTPRequests(requests)
            : () =>
                  throwError(
                      'The queryLSP command is unavailable because the current repository does not exist on the Sourcegraph instance.'
                  ),
        forceUpdateTooltip: () => {
            // TODO(sqs): implement tooltips on the browser extension
        },
    }
    return context
}

function editClientSettings(args: UpdateExtensionSettingsArgs): Promise<void> {
    return new Promise<StorageItems>(resolve => storage.getSync(storageItems => resolve(storageItems))).then(
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
}
