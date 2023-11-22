import { mutateSettings } from '@sourcegraph/shared/src/settings/edit'

import type { SettingsEdit } from '../../graphql-operations'

import { getUser } from './api'
import type { GraphQLClient } from './GraphQlClient'

/**
 * Applies an edit to the user settings for the given username.
 */
export async function editUserSettings(
    username: string,
    edit: SettingsEdit,
    graphQLClient: GraphQLClient
): Promise<void> {
    const user = await getUser(graphQLClient, username)
    if (!user) {
        throw new Error(`User not found: ${username}`)
    }
    const [{ latestSettings }] = user.settingsCascade.subjects.slice(-1)
    const lastID = latestSettings ? latestSettings.id : null
    await mutateSettings(graphQLClient, user.id, lastID, edit)
}
