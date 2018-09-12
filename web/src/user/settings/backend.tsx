import { Observable } from 'rxjs'
import { filter, map, mergeMap, take, tap } from 'rxjs/operators'
import { authRequired } from '../../auth'
import { gql, queryGraphQL } from '../../backend/graphql'
import * as GQL from '../../backend/graphqlschema'
import { gqlConfigurationCascade } from '../../settings/configuration'
import { createAggregateError } from '../../util/errors'

/**
 * Refreshes the configuration from the server, which propagates throughout the
 * app to all consumers of configuration settings.
 */
export function refreshConfiguration(): Observable<never> {
    return authRequired.pipe(
        take(1),
        filter(authRequired => !authRequired),
        mergeMap(() => fetchViewerConfiguration()),
        tap(result => gqlConfigurationCascade.next(result)),
        mergeMap(() => [])
    )
}

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
 * Fetches the viewer's configuration from the server. Callers should use refreshConfiguration instead of calling
 * this function, to ensure that the result is propagated consistently throughout the app instead of only being
 * returned to the caller.
 *
 * @return Observable that emits the configuration
 */
function fetchViewerConfiguration(): Observable<GQL.IConfigurationCascade> {
    return queryGraphQL(gql`
        query Configuration {
            viewerConfiguration {
                ...ConfigurationCascadeFields
            }
        }
        ${configurationCascadeFragment}
    `).pipe(
        map(({ data, errors }) => {
            if (!data || !data.viewerConfiguration) {
                throw createAggregateError(errors)
            }
            return data.viewerConfiguration
        })
    )
}

refreshConfiguration()
    .toPromise()
    .then(() => void 0, err => console.error(err))
