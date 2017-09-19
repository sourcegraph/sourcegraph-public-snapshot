import 'rxjs/add/operator/map'
import { BehaviorSubject } from 'rxjs/BehaviorSubject'
import { Observable } from 'rxjs/Observable'
import { queryGraphQL } from './backend/graphql'

/**
 * currentUser is a BehaviorSubject object that always represents the latest
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
export const currentUser = new BehaviorSubject<GQL.IUser | null>(null)

/**
 * fetchCurrentUser can be called to fetch the current user and orgs
 * state from the remote.
 */
export function fetchCurrentUser(): Observable<GQL.IUser | null> {
     return queryGraphQL(`
        query CurrentAuthState {
            root {
                currentUser {
                    id
                    avatarURL
                    email
                    orgs {
                        id
                        name
                        members {
                            id
                            userID
                            username
                            email
                            displayName
                            avatarURL
                            createdAt
                        }
                    }
                }
            }
        }
    `)
        .map(result => {
            if (!result.data) {
                throw new Error('invalid response received from graphql endpoint')
            }
            return result.data.root.currentUser
        })
}
