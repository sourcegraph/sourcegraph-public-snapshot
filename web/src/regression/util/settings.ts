import { GraphQLClient } from './GraphQLClient'
import * as GQL from '../../../../shared/src/graphql/schema'
import { mutateSettings } from '../../../../shared/src/settings/edit'
import { gql, dataOrThrowErrors } from '../../../../shared/src/graphql/graphql'
import { map, filter } from 'rxjs/operators'
import { isDefined } from '../../../../shared/src/util/types'

/**
 * Fetches the user with the given username.
 */
function fetchUser(graphQLClient: GraphQLClient, username: GQL.ID): Promise<GQL.IUser> {
    return graphQLClient
        .queryGraphQL(
            gql`
                query User($username: String!) {
                    user(username: $username) {
                        id
                        settingsCascade {
                            subjects {
                                latestSettings {
                                    id
                                    contents
                                }
                            }
                        }
                    }
                }
            `,
            { username }
        )
        .pipe(
            map(dataOrThrowErrors),
            map(({ user }) => user),
            filter(isDefined)
        )
        .toPromise()
}

/**
 * Applies an edit to the user settings for the given username.
 */
export async function editUserSettings(
    username: string,
    edit: GQL.ISettingsEdit,
    graphQLClient: GraphQLClient
): Promise<void> {
    const user = await fetchUser(graphQLClient, username)
    const [{ latestSettings }] = user.settingsCascade.subjects.slice(-1)
    const lastID = latestSettings ? latestSettings.id : null
    await mutateSettings(graphQLClient, user.id, lastID, edit)
}
