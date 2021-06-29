import { useMemo } from 'react'

import {
    ConfiguredSubjectOrError,
    SettingsCascadeOrError,
    SettingsSubject,
} from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { isDefined } from '@sourcegraph/shared/src/util/types'

import { Settings } from '../../../../../schema/settings.schema'
import {
    InsightDashboard,
    InsightBuiltInDashboard,
    InsightCustomDashboard,
    InsightDashboardOwner,
    InsightsDashboardType,
    isInsightSettingKey,
    ALL_INSIGHTS_DASHBOARD,
} from '../../../../core/types'

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

    const builtInDashboards = getBuiltInDashboards(subjects)
    const customDashboards = getCustomDashboards(subjects)

    return [...builtInDashboards, ...customDashboards]
}

/**
 * Returns built in types of insights dashboards (all, personal, org level dashboards).
 */
function getBuiltInDashboards(subjects: ConfiguredSubjectOrError<Settings>[]): InsightBuiltInDashboard[] {
    const subjectDashboards = subjects.reduce<InsightBuiltInDashboard[]>((dashboards, configuredSubject) => {
        const { settings, subject } = configuredSubject

        if (isErrorLike(settings) || settings === null) {
            return dashboards
        }

        const dashboardOwner = getDashboardOwner(subject)
        const subjectInsightIds = Object.keys(settings).filter(isInsightSettingKey)

        const subjectDashboard: InsightBuiltInDashboard = {
            type: InsightsDashboardType.BuiltIn,
            owner: dashboardOwner,
            insightIds: subjectInsightIds,
        }

        return [...dashboards, subjectDashboard]
    }, [])

    return [ALL_INSIGHTS_DASHBOARD, ...subjectDashboards]
}

/**
 * Returns list of custom insights dashboards generated from settings cascade subjects.
 */
function getCustomDashboards(subjects: ConfiguredSubjectOrError<Settings>[]): InsightCustomDashboard[] {
    return subjects.reduce<InsightCustomDashboard[]>((dashboards, configuredSubject) => {
        const { settings, subject } = configuredSubject

        if (isErrorLike(settings) || settings === null) {
            return dashboards
        }

        const subjectDashboards = Object.keys(settings['insights.dashboards'] ?? {})
            .map<InsightCustomDashboard | undefined>(dashboardKey => {
                // Select dashboard configuration from the subject settings
                const dashboardSettings = settings['insights.dashboards']?.[dashboardKey]

                if (!dashboardSettings) {
                    return undefined
                }

                // Extend settings dashboard configuration with owner info
                return {
                    type: InsightsDashboardType.Custom,
                    owner: getDashboardOwner(subject),
                    ...dashboardSettings,
                }
            })
            .filter(isDefined)

        return [...dashboards, ...subjectDashboards]
    }, [])
}

function getDashboardOwner(subject: SettingsSubject): InsightDashboardOwner {
    if (subject.__typename === 'User') {
        return {
            id: subject.id,
            name: subject.displayName ?? subject.username,
        }
    }

    if (subject.__typename === 'Org') {
        return {
            id: subject.id,
            name: subject.displayName ?? subject.name,
        }
    }

    return {
        id: null,
        name: 'UNKNOWN SUBJECT',
    }
}
