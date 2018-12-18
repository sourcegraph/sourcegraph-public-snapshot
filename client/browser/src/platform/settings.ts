import { applyEdits, parse as parseJSONC } from '@sqs/jsonc-parser'
import { setProperty } from '@sqs/jsonc-parser/lib/edit'
import { Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'
import { SettingsEdit } from '../../../../shared/src/api/client/services/settings'
import { gql, graphQLContent } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import {
    mergeSettings,
    SettingsCascade,
    SettingsCascadeOrError,
    SettingsSubject,
} from '../../../../shared/src/settings/settings'
import { createAggregateError, isErrorLike } from '../../../../shared/src/util/errors'
import storage, { StorageItems } from '../browser/storage'
import { isInPage } from '../context'
import { getContext } from '../shared/backend/context'
import { queryGraphQL } from '../shared/backend/graphql'
import { sourcegraphUrl } from '../shared/util/context'

/**
 * The settings cascade consisting solely of client settings.
 */
export const storageSettingsCascade: Observable<SettingsCascade> = isInPage
    ? of({ subjects: [], final: {} })
    : storage.observeSync('clientSettings').pipe(
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
export function fetchViewerSettings(): Observable<Pick<GQL.ISettingsCascade, 'subjects' | 'final'>> {
    return queryGraphQL({
        ctx: getContext({ repoKey: '', isRepoSpecific: false }),
        request: gql`
            query ViewerConfiguration {
                viewerConfiguration {
                    ...ConfigurationCascadeFields
                }
            }
            ${configurationCascadeFragment}
        `[graphQLContent],
        url: sourcegraphUrl,
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
}

/**
 * Applies an edit and persists the result to client settings.
 */
export function editClientSettings(edit: SettingsEdit | string): Promise<void> {
    return new Promise<StorageItems>(resolve => storage.getSync(storageItems => resolve(storageItems))).then(
        storageItems => {
            const prev = storageItems.clientSettings
            const next =
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
            return new Promise<undefined>(resolve =>
                storage.setSync({ clientSettings: next }, () => {
                    resolve(undefined)
                })
            )
        }
    )
}
