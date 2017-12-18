import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { queryGraphQL } from '../backend/graphql'

/**
 * Fetches all users.
 *
 * @return Observable that emits the list of users
 */
export function fetchAllUsers(): Observable<GQL.IUser[]> {
    return queryGraphQL(`query Users {
        users {
            nodes {
                id
                username
                displayName
                email
                createdAt
                latestSettings {
                    createdAt
                    configuration { contents }
                }
                orgs { name }
                tags { name }
            }
        }
    }`).pipe(
        map(({ data, errors }) => {
            if (!data || !data.users) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.users.nodes
        })
    )
}
