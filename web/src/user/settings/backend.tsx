import { Observable, Subject } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../../shared/src/util/errors'
import { queryGraphQL } from '../../backend/graphql'

export const settingsRefreshes = new Subject<void>()

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
 * Fetches the viewer's settings from the server. Callers should use settingsRefreshes#next instead of calling
 * this function, to ensure that the result is propagated consistently throughout the app instead of only being
 * returned to the caller.
 *
 * @return Observable that emits the settings
 */
export function fetchViewerSettings(): Observable<GQL.ISettingsCascade> {
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
