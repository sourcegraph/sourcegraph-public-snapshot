import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { Settings } from '../../../../../../../../../schema/settings.schema'
import { INSIGHTS_ALL_REPOS_SETTINGS_KEY, isInsightSettingKey } from '../../../../../../../../core/types'

interface GetBackendInsightIdsInput extends SettingsCascadeProps<Settings> {
    insightIds: string[]
}

/**
 * Returns filtered be insights only ids list.
 *
 * Dashboard insight ids field contains all insights (extension based and be based). To avoid
 * unnecessary gql BE insights request we should separate BE insights from extension like insights.
 */
export function getBackendInsightIds(input: GetBackendInsightIdsInput): string[] {
    const { insightIds, settingsCascade } = input
    const { final } = settingsCascade

    if (!final || isErrorLike(final)) {
        return insightIds
    }

    const backendBasedInsightIds = new Set(
        Object.keys((final?.[INSIGHTS_ALL_REPOS_SETTINGS_KEY] as object) ?? {}).filter(isInsightSettingKey)
    )

    return insightIds.filter(id => backendBasedInsightIds.has(id))
}
