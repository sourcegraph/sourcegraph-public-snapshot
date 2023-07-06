import { viewerSettingsQuery } from '@sourcegraph/shared/src/backend/settings'

import { createAggregateError } from '$lib/common'
import type { ViewerSettingsResult } from '$lib/graphql/shared'
import { getDocumentNode, type GraphQLClient } from '$lib/http-client'
import { gqlToCascade } from '$lib/shared'
import type { SettingsCascadeOrError } from '$lib/shared'

export { viewerSettingsQuery }

export async function fetchUserSettings(client: GraphQLClient): Promise<SettingsCascadeOrError> {
    const response = await client.query<ViewerSettingsResult>({ query: getDocumentNode(viewerSettingsQuery) })
    if (!response.data?.viewerSettings) {
        throw createAggregateError(response.errors)
    }

    return gqlToCascade(response.data.viewerSettings)
}
