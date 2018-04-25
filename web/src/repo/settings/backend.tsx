import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../../backend/graphql'
import * as GQL from '../../backend/graphqlschema'
import { createAggregateError } from '../../util/errors'

/**
 * Fetches a repository.
 */
export function fetchRepository(uri: string): Observable<GQL.IRepository> {
    return queryGraphQL(
        gql`
            query Repository($uri: String!) {
                repository(uri: $uri) {
                    id
                    uri
                    viewerCanAdminister
                    enabled
                    mirrorInfo {
                        remoteURL
                        cloneInProgress
                        cloneProgress
                        cloned
                        updatedAt
                    }
                }
            }
        `,
        { uri }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.repository) {
                throw createAggregateError(errors)
            }
            return data.repository
        })
    )
}
