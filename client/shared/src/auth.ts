import { CurrentAuthStateResult } from './graphql-operations'
import { gql } from './graphql/graphql'

export const currentAuthStateQuery = gql`
    query CurrentAuthState {
        currentUser {
            __typename
            id
            databaseID
            username
            avatarURL
            email
            displayName
            siteAdmin
            tags
            url
            settingsURL
            organizations {
                nodes {
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
            tags
        }
    }
`
export type AuthenticatedUser = NonNullable<CurrentAuthStateResult['currentUser']>
