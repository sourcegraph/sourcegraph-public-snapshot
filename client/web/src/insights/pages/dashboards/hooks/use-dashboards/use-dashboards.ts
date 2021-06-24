import { useMemo } from 'react'

import {
    ConfiguredSubjectOrError,
    SettingsCascadeOrError,
    SettingsSubject,
    Settings,
} from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { INSIGHT_DASHBOARD_PREFIX, InsightDashboard, InsightDashboardConfiguration } from '../../../../core/types'

interface InsightDashboardOwner {
    id: string | null
    name: string
}

interface InsightDashboardInfo extends InsightDashboard {
    /**
     * Subject that has particular dashboard, it can be personal setting
     * or organization setting subject.
     */
    owner: InsightDashboardOwner
}

/**
 * React hook that returns all valid and available insights dashboards.
 */
export function useDashboards(settingsCascade: SettingsCascadeOrError): InsightDashboardInfo[] {
    const { subjects } = settingsCascade

    return useMemo(() => getInsightsDashboards(subjects), [subjects])
}

/**
 * Returns all valid and reachable for the user insight dashboards.
 */
export function getInsightsDashboards(subjects: ConfiguredSubjectOrError<Settings>[] | null): InsightDashboardInfo[] {
    if (subjects === null) {
        return []
    }

    return subjects.reduce<InsightDashboardInfo[]>((dashboards, subject) => {
        const settings = subject.settings

        if (isErrorLike(settings) || settings === null) {
            return dashboards
        }

        const subjectDashboards = Object.keys(settings)
            .filter(key => key.startsWith(INSIGHT_DASHBOARD_PREFIX))
            .map(dashboardKey => {
                const dashboardSettings: InsightDashboardConfiguration = settings[dashboardKey]

                return {
                    id: dashboardKey,
                    visibility: subject.subject.id,
                    owner: getDashboardOwner(subject.subject),
                    ...dashboardSettings,
                }
            })

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
