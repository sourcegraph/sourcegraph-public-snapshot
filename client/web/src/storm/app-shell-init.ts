import type { GraphQLClient } from '@sourcegraph/http-client'
import { TemporarySettingsStorage } from '@sourcegraph/shared/src/settings/temporary/TemporarySettingsStorage'
import { TelemetryRecorder, noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'

import { getWebGraphQLClient } from '../backend/graphql'

export interface AppShellInit {
    graphqlClient: GraphQLClient
    temporarySettingsStorage: TemporarySettingsStorage
    telemetryRecorder: TelemetryRecorder
}

export async function initAppShell(): Promise<AppShellInit> {
    const graphqlClient = await getWebGraphQLClient()
    const temporarySettingsStorage = new TemporarySettingsStorage(graphqlClient, window.context.isAuthenticatedUser)
    const telemetryRecorder = noOpTelemetryRecorder

    return {
        graphqlClient,
        temporarySettingsStorage,
        telemetryRecorder,
    }
}
