import { useMemo } from 'react'

import { IOrg, IUser } from '@sourcegraph/shared/src/graphql/schema'
import {
    ConfiguredSubjectOrError,
    SettingsCascadeOrError,
    SettingsSubject,
} from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { isDefined } from '@sourcegraph/shared/src/util/types'

import { Settings } from '../../../../../../schema/settings.schema'
import {
    InsightDashboard,
    INSIGHTS_DASHBOARDS_SETTINGS_KEY,
    InsightsDashboardType,
    isInsightSettingKey,
} from '../../../../../core/types'
import { InsightDashboardOwner } from '../../../../../core/types/dashboard/core'

/**
 * Special virtual dashboard - "All Insights"
 */
export const ALL_INSIGHTS_DASHBOARD: InsightDashboard = {
    id: 'all',
    type: InsightsDashboardType.All,
}

/**
 * React hook that returns all valid and available insights dashboards.
 */
export function useDashboards(settingsCascade: SettingsCascadeOrError): InsightDashboard[] {
    const { subjects } = settingsCascade

    return useMemo(() => getInsightsDashboards(subjects), [subjects])
}

/**
 * Returns all valid and reachable for a user insight-dashboards.
 */
export function getInsightsDashboards(subjects: ConfiguredSubjectOrError<Settings>[] | null): InsightDashboard[] {
    if (subjects === null) {
        return []
    }

    const subjectDashboards = subjects.flatMap(configuredSubject => {
        const { settings, subject } = configuredSubject

        if (isErrorLike(settings) || !settings || !isSubjectSupported(subject)) {
            return []
        }

        return getSubjectDashboards(subject, settings)
    })

    return [ALL_INSIGHTS_DASHBOARD, ...subjectDashboards]
}

/**
 * Currently we support only two types of subject that can have insights dashboard.
 */
type SupportedSubject = IUser | IOrg

const SUPPORTED_TYPES_OF_SUBJECT = new Set<SettingsSubject['__typename']>(['User', 'Org'])
const isSubjectSupported = (subject: SettingsSubject): subject is SupportedSubject =>
    SUPPORTED_TYPES_OF_SUBJECT.has(subject.__typename)

/**
 * Returns all subject dashboards and one special (built-in) dashboard that includes
 * all insights from subject settings.
 */
function getSubjectDashboards(subject: SupportedSubject, settings: Settings): InsightDashboard[] {
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
        .map<InsightDashboard | undefined>(dashboardKey => {
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
        })
        .filter(isDefined)

    return [subjectBuiltInDashboard, ...subjectDashboards]
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
