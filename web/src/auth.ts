import { Observable } from 'rxjs/Observable'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { tap } from 'rxjs/operators/tap'
import { ReplaySubject } from 'rxjs/ReplaySubject'
import { gql, queryGraphQL } from './backend/graphql'

/**
 * Always represents the latest
 * state of the currently authenticated user.
 *
 * Unlike sourcegraphContext.user, the global currentUser object contains
 * locally mutable properties such as email, displayName, and avatarUrl, all
 * of which are expected to change over the course of a user's session.
 *
 * Note that currentUser is not designed to survive across changes in the
 * currently authenicated user. Sign in, sign out, and account changes are
 * all expected to refresh the app.
 */
export const currentUser = new ReplaySubject<GQL.IUser | null>(1)

/**
 * fetchCurrentUser can be called to fetch the current user, orgs, and config
 * state from the remote. Emits no items, completes when done.
 */
export function fetchCurrentUser(): Observable<never> {
    return queryGraphQL(gql`
        query CurrentAuthState {
            currentUser {
                __typename
                auth0ID
                sourcegraphID
                username
                avatarURL
                email
                username
                displayName
                verified
                orgs {
                    id
                    name
                    tags {
                        name
                    }
                }
                tags {
                    id
                    name
                }
            }
        }
    `).pipe(
        tap(({ data, errors }) => {
            if (!data) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            currentUser.next(data.currentUser)
        }),
        mergeMap(() => [])
    )
}

fetchCurrentUser().toPromise()
