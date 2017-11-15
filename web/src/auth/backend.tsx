import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { queryGraphQL } from '../backend/graphql'

export function isUsernameAvailable(username: string): Observable<boolean> {
    return queryGraphQL(
        `
        query UsernameAvailability($username: String!) {
            root {
                isUsernameAvailable(username: $username)
            }
        }
        `,
        {
            username,
        }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.root || typeof data.root.isUsernameAvailable !== 'boolean') {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.root.isUsernameAvailable
        })
    )
}
