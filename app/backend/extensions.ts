import { Controller as ExtensionsContextController } from '@sourcegraph/extensions-client-common/lib/controller'
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
import * as JSONC from '@sqs/jsonc-parser'
import { applyEdits } from '@sqs/jsonc-parser'
import { removeProperty, setProperty } from '@sqs/jsonc-parser/lib/edit'
import { ConfigurationUpdateParams } from 'cxp/module/protocol'
import { isEqual } from 'lodash'
import Alert from 'mdi-react/AlertIcon'
import MenuDown from 'mdi-react/MenuDownIcon'
import Menu from 'mdi-react/MenuIcon'
import { combineLatest, Observable, Subject, throwError } from 'rxjs'
import { distinctUntilChanged, map, mergeMap, switchMap, take } from 'rxjs/operators'
import storage from '../../extension/storage'
import { getContext } from './context'
import { createAggregateError, isErrorLike } from './errors'
import { queryGraphQL } from './graphql'

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
        updateExtensionSettings: (
            subjectID,
            args: { extensionID: string; edit?: ConfigurationUpdateParams; enabled?: boolean; remove?: boolean }
        ) => {
            if (subjectID !== 'Client') {
                return throwError('Cannot update settings for ' + subjectID + '.')
            }
            const update = new Subject<undefined>()
            storage.getSync(storageItems => {
                let clientSettings = storageItems.clientSettings

                const format = { tabSize: 2, insertSpaces: true, eol: '\n' }

                if (args.edit) {
                    clientSettings = applyEdits(
                        clientSettings,
                        // TODO(chris): remove `.slice()` (which guards against
                        // mutation) once
                        // https://github.com/Microsoft/node-jsonc-parser/pull/12
                        // is merged in.
                        setProperty(clientSettings, args.edit.path.slice(), args.edit.value, format)
                    )
                } else if (typeof args.enabled === 'boolean') {
                    clientSettings = applyEdits(
                        clientSettings,
                        setProperty(clientSettings, ['extensions', args.extensionID], args.enabled, format)
                    )
                } else if (args.remove) {
                    clientSettings = applyEdits(
                        clientSettings,
                        removeProperty(clientSettings, ['extensions', args.extensionID], format)
                    )
                }
                storage.setSync({ clientSettings }, () => {
                    update.next(undefined)
                })
            })
            return update
        },
        queryGraphQL: (request, variables) =>
            storage.observeSync('sourcegraphURL').pipe(
                take(1),
                mergeMap(url =>
                    queryGraphQL(getContext({ repoKey: '', isRepoSpecific: false }), request, variables, url)
                )
            ),
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
