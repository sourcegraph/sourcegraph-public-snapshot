import { useMemo } from 'react'

import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { InsightDashboard, Settings } from '../../../../../schema/settings.schema'

export interface useDashboardSettingsProps extends SettingsCascadeProps<Settings> {
    /**
     * Final settings used below as a store of all existing dashboards
     * Usually we have a validation step for the title of dashboard because
     * users can't have two dashboards with the same name/id. In edit mode
     * we should allow users to have insight with id (camelCase(dashboard name))
     * which already exists in the settings. For turning off this id/title
     * validation we remove current dashboard from the final settings.
     */
    excludeDashboardIds?: string[]
}

/**
 * Returns dashboard final (merged) configurations map with all insight dashboards.
 */
export function useDashboardSettings(props: useDashboardSettingsProps): Record<string, InsightDashboard> {
    const { settingsCascade, excludeDashboardIds = [] } = props

    return useMemo(() => {
        if (isErrorLike(settingsCascade.final) || !settingsCascade.final) {
            return {}
        }

        const dashboardSettings = { ...settingsCascade.final['insights.dashboards'] }

        for (const id of excludeDashboardIds) {
            delete dashboardSettings?.[id]
        }

        return dashboardSettings
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [settingsCascade.final, ...excludeDashboardIds])
}
