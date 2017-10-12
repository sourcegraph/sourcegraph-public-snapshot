import 'rxjs/add/operator/do'
import 'rxjs/add/operator/mergeMap'
import { Observable } from 'rxjs/Observable'
import { ReplaySubject } from 'rxjs/ReplaySubject'
import { queryGraphQL } from './backend/graphql'

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
 * fetchCurrentUser can be called to fetch the current user and orgs state from the remote.
 * Emits no items, completes when done.
 */
export function fetchCurrentUser(): Observable<never> {
    return queryGraphQL(`
        query CurrentAuthState {
            root {
                currentUser {
                    id
                    sourcegraphID
                    username
                    avatarURL
                    email
                    username
                    displayName
                    orgs {
                        id
                        name
                    }
                    tags {
                        id
                        name
                    }
                }
            }
        }
    `)
        .do(result => {
            if (!result.data) {
                throw new Error('invalid response received from graphql endpoint')
            }
            currentUser.next(result.data.root.currentUser)
        })
        .mergeMap(() => [])
}
