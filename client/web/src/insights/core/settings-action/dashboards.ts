import { camelCase } from 'lodash'

import { InsightDashboard } from '../../../schema/settings.schema'
import { INSIGHTS_DASHBOARDS_SETTINGS_KEY } from '../types'

import { modify } from './utils'

/**
 * Adds sanitized dashboard configuration to the settings content.
 *
 * @param settings - original subject settings
 * @param dashboardConfiguration - a dashboard configurations
 */
export function addDashboardToSettings(settings: string, dashboardConfiguration: InsightDashboard): string {
    return modify(
        settings,
        [INSIGHTS_DASHBOARDS_SETTINGS_KEY, camelCase(dashboardConfiguration.title)],
        dashboardConfiguration
    )
}
