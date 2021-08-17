import { camelCase } from 'lodash'
import { useContext } from 'react'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'

import { InsightsApiContext } from '../../../../../core/backend/api-provider'
import { addInsightToDashboard } from '../../../../../core/settings-action/dashboards'
import { addInsightToSettings } from '../../../../../core/settings-action/insights'
import { InsightDashboard, InsightTypePrefix, isVirtualDashboard } from '../../../../../core/types'
import { isSettingsBasedInsightsDashboard } from '../../../../../core/types/dashboard/real-dashboard'
import { SearchBackendBasedInsight, SearchBasedBackendFilters } from '../../../../../core/types/insight/search-insight'

interface CreateInsightInputs {
    insightName: string
    originalInsight: SearchBackendBasedInsight
    dashboard: InsightDashboard
    filters: SearchBasedBackendFilters
}

export interface UseInsightFilterCreationApi {
    create: (inputs: CreateInsightInputs) => Promise<void>
}

export interface UseInsightFilterCreationProps extends PlatformContextProps<'updateSettings'> {}

export function useInsightFilterCreation(props: UseInsightFilterCreationProps): UseInsightFilterCreationApi {
    const { platformContext } = props
    const { getSubjectSettings, updateSubjectSettings } = useContext(InsightsApiContext)

    const createInsightWithFilters = async (inputs: CreateInsightInputs): Promise<void> => {
        const { dashboard, insightName, originalInsight, filters } = inputs
        // Get id of insight setting subject (owner of it insight)
        const subjectId = isVirtualDashboard(dashboard) ? originalInsight.visibility : dashboard.owner.id

        // Create new insight by name and last valid filters value
        const newInsight: SearchBackendBasedInsight = {
            ...originalInsight,
            id: `${InsightTypePrefix.search}.${camelCase(insightName)}`,
            title: insightName,
            filters,
        }

        const settings = await getSubjectSettings(subjectId).toPromise()

        const updatedSettings = [
            (settings: string) => addInsightToSettings(settings, newInsight),
            (settings: string) => {
                // Virtual and built-in dashboards calculate their insight dynamically in runtime
                // no need to store insight list for them explicitly
                if (isVirtualDashboard(dashboard) || !isSettingsBasedInsightsDashboard(dashboard)) {
                    return settings
                }

                return addInsightToDashboard(settings, dashboard.settingsKey, newInsight.id)
            },
        ].reduce((settings, transformer) => transformer(settings), settings.contents)

        await updateSubjectSettings(platformContext, subjectId, updatedSettings).toPromise()

        return
    }

    return { create: createInsightWithFilters }
}
