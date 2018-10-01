import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../../backend/graphql'
import * as GQL from '../../backend/graphqlschema'
import { createAggregateError } from '../../util/errors'

/**
 * Fetches a repository.
 */
export function fetchRepository(name: string): Observable<GQL.IRepository> {
    return queryGraphQL(
        gql`
            query Repository($name: String!) {
                repository(name: $name) {
                    id
                    name
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
        { name }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.repository) {
                throw createAggregateError(errors)
            }
            return data.repository
        })
    )
}
