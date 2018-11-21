import { applyEdits, parse as parseJSONC } from '@sqs/jsonc-parser'
import { removeProperty, setProperty } from '@sqs/jsonc-parser/lib/edit'
import { isEqual } from 'lodash'
import { combineLatest, Observable, Subject } from 'rxjs'
import { distinctUntilChanged, map, mapTo, startWith, switchMap } from 'rxjs/operators'
import { gql, graphQLContent } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { UpdateExtensionSettingsArgs } from '../../../../shared/src/settings/edit'
import {
    gqlToCascade,
    mergeSettings,
    SettingsCascade,
    SettingsCascadeOrError,
    SettingsSubject,
} from '../../../../shared/src/settings/settings'
import { createAggregateError, isErrorLike } from '../../../../shared/src/util/errors'
import storage, { StorageItems } from '../browser/storage'
import { getContext } from '../shared/backend/context'
import { queryGraphQL } from '../shared/backend/graphql'

const storageSettingsCascade: Observable<SettingsCascade> = storage.observeSync('clientSettings').pipe(
    map(clientSettingsString => parseJSONC(clientSettingsString || '')),
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
export const settingsCascadeRefreshes = new Subject<void>()

/**
 * Always represents the entire settings cascade; i.e., it contains the individual settings from the various
 * settings subjects (orgs, user, etc.).
 *
 * TODO(sqs): This uses the DEPRECATED GraphQL Query.viewerConfiguration and ConfigurationCascade for backcompat.
 */
const gqlSettingsCascade: Observable<Pick<GQL.ISettingsCascade, 'subjects' | 'final'>> = combineLatest(
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

/**
 * Applies an edit and persists the result to client settings.
 */
export function editClientSettings(args: UpdateExtensionSettingsArgs): Promise<void> {
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
