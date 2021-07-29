import { ErrorLike, isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { Settings } from '../../../../../../../../../schema/settings.schema'
import { INSIGHTS_ALL_REPOS_SETTINGS_KEY, isInsightSettingKey } from '../../../../../../../../core/types'
import { useDistinctValue } from '../../../../../../../../hooks/use-distinct-value'

interface GetBackendInsightIdsInput {
    insightIds: string[]
    finalSettings: Settings | ErrorLike | null
}

/**
 * Returns filtered be insights only ids list.
 *
 * Dashboard insight ids field contains all insights (extension based and be based). To avoid
 * unnecessary gql BE insights request we should separate BE insights from extension like insights.
 */
export function useBackendInsightIds(input: GetBackendInsightIdsInput): string[] {
    return useDistinctValue(getBackendInsightIds(input))
}

export function getBackendInsightIds(input: GetBackendInsightIdsInput): string[] {
    const { insightIds, finalSettings } = input

    if (!finalSettings || isErrorLike(finalSettings)) {
        return insightIds
    }

    const backendBasedInsightIds = new Set(
        Object.keys((finalSettings?.[INSIGHTS_ALL_REPOS_SETTINGS_KEY] as object) ?? {}).filter(isInsightSettingKey)
    )

    return insightIds.filter(id => backendBasedInsightIds.has(id))
}
