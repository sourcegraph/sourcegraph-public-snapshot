import type { GraphQLClient } from '@sourcegraph/http-client'
import { TemporarySettingsStorage } from '@sourcegraph/shared/src/settings/temporary/TemporarySettingsStorage'

import { getWebGraphQLClient } from '../backend/graphql'
import { useDeveloperSettings } from '../stores'

export interface AppShellInit {
    graphqlClient: GraphQLClient
    temporarySettingsStorage: TemporarySettingsStorage
}

export async function initAppShell(): Promise<AppShellInit> {
    const graphqlClient = await getWebGraphQLClient()
    const temporarySettingsStorage = new TemporarySettingsStorage(
        graphqlClient,
        window.context.isAuthenticatedUser,
        process.env.NODE_ENV === 'development' || useDeveloperSettings.getState().enabled
    )

    return {
        graphqlClient,
        temporarySettingsStorage,
    }
}
