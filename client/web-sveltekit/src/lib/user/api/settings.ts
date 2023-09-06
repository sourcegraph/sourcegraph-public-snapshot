import { viewerSettingsQuery } from '@sourcegraph/shared/src/backend/settings'

import { gql, query } from '$lib/graphql'
import type { ViewerSettingsResult } from '$lib/graphql/shared'
import { gqlToCascade } from '$lib/shared'
import type { SettingsCascadeOrError } from '$lib/shared'

export { viewerSettingsQuery }

export async function fetchUserSettings(): Promise<SettingsCascadeOrError> {
    const data = await query<ViewerSettingsResult>(gql(viewerSettingsQuery))
    if (!data?.viewerSettings) {
        throw new Error('Unable to fetch user settings')
    }

    return gqlToCascade(data.viewerSettings)
}
