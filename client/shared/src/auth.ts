import { gql } from '@sourcegraph/http-client'

import type { CurrentAuthStateResult } from './graphql-operations'

/**
 * This GraphQL should be in sync with server-side current user preloading
 * defined in `cmd/frontend/internal/app/jscontext/jscontext.go`.
 *
 * There's a standalone `SourcegraphContextCurrentUser` type derived from
 * the generated `AuthenticatedUser` type. It will make sure that we don't
 * forget to add new fields the server logic if client side query changes.
 *
 * Ideally, we need to generate the Typescript type from the golang interface
 * used in the `jscontext.go`. This is a follow-up improvement that we need
 * to look into.
 */
export const currentAuthStateQuery = gql`
    query CurrentAuthState {
        currentUser {
            __typename
            id
            databaseID
            username
            avatarURL
            displayName
            siteAdmin
            url
            settingsURL
            organizations {
                nodes {
                    __typename
                    id
                    name
                    displayName
                    url
                    settingsURL
                }
            }
            session {
                canSignOut
            }
            viewerCanAdminister
            tosAccepted
            hasVerifiedEmail
            completedPostSignup
            emails {
                email
                verified
                isPrimary
            }
            latestSettings {
                id
                contents
            }
            permissions {
                nodes {
                    id
                    displayName
                }
            }
        }
    }
`
export type AuthenticatedUser = NonNullable<CurrentAuthStateResult['currentUser']>
