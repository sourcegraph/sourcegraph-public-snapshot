import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { gql, queryGraphQL } from '../../backend/graphql'

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
                }
            }
        `,
        { uri }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.repository) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.repository
        })
    )
}
