import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { gql, queryGraphQL } from '../backend/graphql'

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
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.isUsernameAvailable
        })
    )
}
