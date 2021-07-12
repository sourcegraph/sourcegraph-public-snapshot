import { IOrg, IUser } from '@sourcegraph/shared/src/graphql/schema'
import { SettingsSubject } from '@sourcegraph/shared/src/settings/settings'
import { isDefined } from '@sourcegraph/shared/src/util/types'

import { Settings } from '../../../../../../schema/settings.schema'
import {
    INSIGHTS_DASHBOARDS_SETTINGS_KEY,
    InsightsDashboardType,
    InsightDashboard,
    isInsightSettingKey,
    SettingsBasedInsightDashboard,
    InsightDashboardOwner,
} from '../../../../../core/types'

/**
 * Currently we support only two types of subject that can have insights dashboard.
 */
type SupportedSubject = IUser | IOrg

const SUPPORTED_TYPES_OF_SUBJECT = new Set<SettingsSubject['__typename']>(['User', 'Org'])

export const isSubjectSupported = (subject: SettingsSubject): subject is SupportedSubject =>
    SUPPORTED_TYPES_OF_SUBJECT.has(subject.__typename)

/**
 * Returns all subject dashboards and one special (built-in) dashboard that includes
 * all insights from subject settings.
 */
export function getSubjectDashboards(subject: SupportedSubject, settings: Settings): InsightDashboard[] {
    const { dashboardType, ...owner } = getDashboardOwnerInfo(subject)

    const subjectBuiltInDashboard: InsightDashboard = {
        owner,
        id: owner.id,
        builtIn: true,
        title: owner.name,
        type: dashboardType,
        insightIds: Object.keys(settings).filter(isInsightSettingKey),
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
): SettingsBasedInsightDashboard | undefined {
    if (!isSubjectSupported(subject)) {
        return undefined
    }

    const { dashboardType, ...owner } = getDashboardOwnerInfo(subject)

    // Select dashboard configuration from the subject settings
    const dashboardSettings = settings[INSIGHTS_DASHBOARDS_SETTINGS_KEY]?.[dashboardKey]

    if (!dashboardSettings) {
        return undefined
    }

    return {
        owner,
        type: dashboardType,
        settingsKey: dashboardKey,
        ...dashboardSettings,
    }
}

interface DashboardOwnerInfo extends InsightDashboardOwner {
    /** Currently we support only two types of subject that can have insights dashboard. */
    dashboardType: InsightsDashboardType.Personal | InsightsDashboardType.Organization
}

/**
 * Return dashboard owner info by subject configuration
 *
 * @param subject - subject settings (User, Organization, Site, Client)
 */
function getDashboardOwnerInfo(subject: SupportedSubject): DashboardOwnerInfo {
    switch (subject.__typename) {
        case 'Org':
            return {
                id: subject.id,
                name: subject.displayName ?? subject.name,
                dashboardType: InsightsDashboardType.Organization,
            }

        case 'User':
            return {
                id: subject.id,
                name: subject.displayName ?? subject.username,
                dashboardType: InsightsDashboardType.Personal,
            }
    }
}
