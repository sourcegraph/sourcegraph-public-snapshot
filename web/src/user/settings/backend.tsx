import { Observable } from 'rxjs'
import { filter, map, mergeMap, take, tap } from 'rxjs/operators'
import { createAggregateError } from '../../../../shared/src/errors'
import { gql } from '../../../../shared/src/graphql'
import * as GQL from '../../../../shared/src/graphqlschema'
import { authRequired } from '../../auth'
import { queryGraphQL } from '../../backend/graphql'
import { settingsCascade } from '../../settings/configuration'

/**
 * Refreshes the viewer's settings from the server, which propagates throughout the app to all consumers of
 * settings.
 */
export function refreshSettings(): Observable<never> {
    return authRequired.pipe(
        take(1),
        filter(authRequired => !authRequired),
        mergeMap(() => fetchViewerSettings()),
        tap(result => settingsCascade.next(result)),
        mergeMap(() => [])
    )
}

const settingsCascadeFragment = gql`
    fragment SettingsCascadeFields on SettingsCascade {
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
        final
    }
`

/**
 * Fetches the viewer's settings from the server. Callers should use refreshSettings instead of calling
 * this function, to ensure that the result is propagated consistently throughout the app instead of only being
 * returned to the caller.
 *
 * @return Observable that emits the settings
 */
function fetchViewerSettings(): Observable<GQL.ISettingsCascade> {
    return queryGraphQL(gql`
        query ViewerSettings {
            viewerSettings {
                ...SettingsCascadeFields
            }
        }
        ${settingsCascadeFragment}
    `).pipe(
        map(({ data, errors }) => {
            if (!data || !data.viewerSettings) {
                throw createAggregateError(errors)
            }
            return data.viewerSettings
        })
    )
}

refreshSettings()
    .toPromise()
    .then(() => void 0, err => console.error(err))
