import { isDefined } from '@sourcegraph/common'
import { SettingsSubject } from '@sourcegraph/shared/src/settings/settings'

import { Settings } from '../../../../../../../schema/settings.schema'
import {
    CustomInsightDashboard,
    InsightDashboard,
    InsightDashboardOwner,
    INSIGHTS_ALL_REPOS_SETTINGS_KEY,
    INSIGHTS_DASHBOARDS_SETTINGS_KEY,
    InsightsDashboardScope,
    InsightsDashboardType,
    isInsightSettingKey,
} from '../../../../types'
import { isSearchBasedInsightId, SearchBasedExtensionInsightSettings } from '../../../../types/insight/search-insight'
import { isSubjectInsightSupported, SupportedInsightSubject } from '../../../../types/subjects'

/**
 * Returns all insights (extension based and be based) ids.
 */
export function getInsightIdsFromSettings(settings: Settings): string[] {
    // On top level of settings we store extension based insights only
    const extensionBasedInsightIds = Object.keys(settings)
        .filter(isInsightSettingKey)
        .filter(id => {
            if (isSearchBasedInsightId(id)) {
                const searchBasedInsightRepositories = (settings[id] as SearchBasedExtensionInsightSettings)
                    ?.repositories

                // Since dashboards currently work only with the insights page
                // Filter out all insights that are not supposed to be on insights page
                // by checking the repositories property.
                // See details https://github.com/sourcegraph/sourcegraph-search-insights/blob/1b204a579160bab4208a1266cf4ad6e735cdd774/package.json#L77
                return Array.isArray(searchBasedInsightRepositories)
            }

            return true
        })

    // BE insights live in the 'insights.allrepos' map
    const backendBasedInsightIds = Object.keys(settings?.[INSIGHTS_ALL_REPOS_SETTINGS_KEY] ?? {}).filter(
        isInsightSettingKey
    )

    return [...backendBasedInsightIds, ...extensionBasedInsightIds]
}

/**
 * Returns all subject dashboards and one special (built-in) dashboard that includes
 * all insights from subject settings.
 */
export function getSubjectDashboards(subject: SupportedInsightSubject, settings: Settings): InsightDashboard[] {
    const { dashboardScope, ...owner } = getDashboardOwnerInfo(subject)

    const subjectBuiltInDashboard: InsightDashboard = {
        owner,
        id: owner.id,
        type: InsightsDashboardType.BuiltIn,
        title: owner.name,
        scope: dashboardScope,
        insightIds: getInsightIdsFromSettings(settings),
    }

    // Find all subject insights dashboards
    const subjectDashboards = Object.keys(settings[INSIGHTS_DASHBOARDS_SETTINGS_KEY] ?? {})
        .map(dashboardKey => getSubjectDashboardByID(subject, settings, dashboardKey))
        .filter(isDefined)

    return [subjectBuiltInDashboard, ...subjectDashboards]
}

/**
 * Returns settings based dashboard from subject settings by id (key).
 *
 * @param subject - settings subject
 * @param settings - settings map of current subject
 * @param dashboardKey - possible dashboard key (id)
 */
export function getSubjectDashboardByID(
    subject: SettingsSubject,
    settings: Settings,
    dashboardKey: string
): CustomInsightDashboard | undefined {
    if (!isSubjectInsightSupported(subject)) {
        return undefined
    }

    const { dashboardScope, ...owner } = getDashboardOwnerInfo(subject)

    // Select dashboard configuration from the subject settings
    const dashboardSettings = settings[INSIGHTS_DASHBOARDS_SETTINGS_KEY]?.[dashboardKey]

    if (!dashboardSettings) {
        return undefined
    }

    return {
        owner,
        scope: dashboardScope,
        type: InsightsDashboardType.Custom,
        settingsKey: dashboardKey,
        ...dashboardSettings,
    }
}

interface DashboardOwnerInfo extends InsightDashboardOwner {
    /**
     * Currently we support three types of subject that can have insights dashboard.
     */
    dashboardScope: InsightsDashboardScope
}

/**
 * Return dashboard owner info by subject configuration
 *
 * @param subject - subject settings (User, Organization, Site, Client)
 */
export function getDashboardOwnerInfo(subject: SupportedInsightSubject): DashboardOwnerInfo {
    switch (subject.__typename) {
        case 'Site': {
            return {
                id: subject.id,
                name: 'Global',
                dashboardScope: InsightsDashboardScope.Global,
            }
        }
        case 'Org':
            return {
                id: subject.id,
                name: subject.displayName ?? subject.name,
                dashboardScope: InsightsDashboardScope.Organization,
            }

        case 'User':
            return {
                id: subject.id,
                name: subject.displayName ?? subject.username,
                dashboardScope: InsightsDashboardScope.Personal,
            }
    }
}
