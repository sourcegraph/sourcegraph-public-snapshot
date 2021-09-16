import { useMemo } from 'react'

import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { InsightDashboard, Settings } from '../../../../../../schema/settings.schema'

export interface useDashboardSettingsProps extends SettingsCascadeProps<Settings> {
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
