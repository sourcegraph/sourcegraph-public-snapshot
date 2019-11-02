import { applyEdits, parse as parseJSONC } from '@sqs/jsonc-parser'
import { setProperty } from '@sqs/jsonc-parser/lib/edit'
import { from, Observable, ObservableInput } from 'rxjs'
import { map, catchError } from 'rxjs/operators'
import { SettingsEdit } from '../../../shared/src/api/client/services/settings'
import { dataOrThrowErrors, gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { PlatformContext } from '../../../shared/src/platform/context'
import {
    mergeSettings,
    SettingsCascade,
    SettingsCascadeOrError,
    SettingsSubject,
} from '../../../shared/src/settings/settings'
import { isErrorLike } from '../../../shared/src/util/errors'
import { LocalStorageSubject } from '../../../shared/src/util/LocalStorageSubject'
import { observeStorageKey, storage } from '../browser/storage'
import { isInPage } from '../context'
import { isHTTPStatusError } from '../../../shared/src/backend/fetch'

const inPageClientSettingsKey = 'sourcegraphClientSettings'

const createStorageSettingsCascade: () => Observable<SettingsCascade> = () => {
    const storageSubject = isInPage
        ? new LocalStorageSubject<string>(inPageClientSettingsKey, '{}')
        : observeStorageKey('sync', 'clientSettings')

    const subject: SettingsSubject = {
        __typename: 'Client',
        id: 'Client',
        displayName: 'Client',
        viewerCanAdminister: true,
    }

    return storageSubject.pipe(
        map(clientSettingsString => parseJSONC(clientSettingsString || '')),
        map(clientSettings => ({
            subjects: [
                {
                    subject,
                    settings: clientSettings,
                    lastID: null,
                },
            ],
            final: clientSettings || {},
        }))
    )
}

/**
 * The settings cascade consisting solely of client settings.
 */
export const storageSettingsCascade = createStorageSettingsCascade()

/**
 * Merge two settings cascades (used to merge viewer settings and client settings).
 */
export function mergeCascades(
    cascadeOrError: SettingsCascadeOrError,
    cascade: SettingsCascade
): SettingsCascadeOrError {
    return {
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
    }
}

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

/**
 * Fetches the settings cascade for the viewer.
 *
 * TODO(sqs): This uses the DEPRECATED GraphQL Query.viewerConfiguration and ConfigurationCascade for backcompat.
 */
export function fetchViewerSettings(
    requestGraphQL: PlatformContext['requestGraphQL']
): Observable<Pick<GQL.ISettingsCascade, 'subjects' | 'final'>> {
    return from(
        requestGraphQL<GQL.IQuery>({
            request: gql`
                query ViewerConfiguration {
                    viewerConfiguration {
                        ...ConfigurationCascadeFields
                    }
                }
                ${configurationCascadeFragment}
            `,
            variables: {},
            mightContainPrivateInfo: false,
        })
    ).pipe(
        map(dataOrThrowErrors),
        // Suppress deprecation warnings because our use of these deprecated fields is intentional (see tsdoc comment).
        //
        // tslint:disable deprecation
        map(({ viewerConfiguration }) => {
            if (!viewerConfiguration) {
                throw new Error('fetchViewerSettings: empty viewerConfiguration')
            }

            for (const subject of viewerConfiguration.subjects) {
                // User/org/global settings cannot be edited from the
                // browser extension (only client settings can).
                subject.viewerCanAdminister = false
            }

            return {
                subjects: viewerConfiguration.subjects,
                final: viewerConfiguration.merged.contents,
            }
        })
        // tslint:enable deprecation
    )
}

/**
 * Fetches initial user settings, checking whether the user is logged in by
 * catching Ajax errors with a 401 status code.
 */
export const checkUserLoggedInAndFetchSettings = (
    requestGraphQL: PlatformContext['requestGraphQL']
): Observable<
    { userLoggedIn: false } | { userLoggedIn: true; settings: Pick<GQL.ISettingsCascade, 'subjects' | 'final'> }
> =>
    fetchViewerSettings(requestGraphQL).pipe(
        map(settings => ({ userLoggedIn: true, settings })),
        catchError(
            (err): ObservableInput<{ userLoggedIn: false }> => {
                if (isHTTPStatusError(err) && err.response.status === 401) {
                    return [{ userLoggedIn: false }]
                }
                // Workaround for https://github.com/mozilla/webextension-polyfill/issues/210 in the browser extension
                if (isErrorLike(err) && err.message.includes('failed with 401')) {
                    return [{ userLoggedIn: false }]
                }
                throw err
            }
        )
    )

/**
 * Applies an edit and persists the result to client settings.
 */
export async function editClientSettings(edit: SettingsEdit | string): Promise<void> {
    const getNext = (prev: string): string =>
        typeof edit === 'string'
            ? edit
            : applyEdits(
                  prev,
                  // TODO(chris): remove `.slice()` (which guards against mutation) once
                  // https://github.com/Microsoft/node-jsonc-parser/pull/12 is merged in.
                  setProperty(prev, edit.path.slice(), edit.value, {
                      tabSize: 2,
                      insertSpaces: true,
                      eol: '\n',
                  })
              )
    if (isInPage) {
        const prev = localStorage.getItem(inPageClientSettingsKey) || ''
        const next = getNext(prev)

        localStorage.setItem(inPageClientSettingsKey, next)

        return Promise.resolve()
    }

    const { clientSettings: prev = '{}' } = await storage.sync.get()
    const next = getNext(prev)

    await storage.sync.set({ clientSettings: next })
}
