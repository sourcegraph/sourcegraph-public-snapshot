import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { gql, queryGraphQL } from '../backend/graphql'
import { createAggregateError } from '../util/errors'

export function isUsernameAvailable(username: string): Observable<boolean> {
    return queryGraphQL(
        gql`
            query UsernameAvailability($username: String!) {
                isUsernameAvailable(username: $username)
            }
        `,
        {
            username,
        }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || typeof data.isUsernameAvailable !== 'boolean') {
                throw createAggregateError(errors)
            }
            return data.isUsernameAvailable
        })
    )
}
