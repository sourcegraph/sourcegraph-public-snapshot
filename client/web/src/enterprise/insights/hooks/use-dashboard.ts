import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'

import { InsightDashboardSettingsApi } from '../core/types'
import { findDashboardByUrlId } from '../pages/dashboards/dashboard-page/components/dashboards-content/utils/find-dashboard-by-url-id'

import { useDashboards } from './use-dashboards/use-dashboards'

interface UseDashboardProps extends SettingsCascadeProps {
    dashboardId?: string
}

export function useDashboard(props: UseDashboardProps): InsightDashboardSettingsApi | null {
    const { settingsCascade, dashboardId } = props

    const dashboards = useDashboards(settingsCascade)

    return findDashboardByUrlId(dashboards, dashboardId ?? '') ?? null
}
