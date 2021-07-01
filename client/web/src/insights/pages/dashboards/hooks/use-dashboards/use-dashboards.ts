import { useMemo } from 'react'

import {
    ConfiguredSubjectOrError,
    SettingsCascadeOrError,
    SettingsSubject,
} from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { isDefined } from '@sourcegraph/shared/src/util/types'

import { Settings } from '../../../../../schema/settings.schema'
import { InsightDashboard, InsightDashboardOwner } from '../../../../core/types'

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

    return subjects.reduce<InsightDashboard[]>((dashboards, subject) => {
        const settings = subject.settings

        if (isErrorLike(settings) || settings === null) {
            return dashboards
        }

        const subjectDashboards = Object.keys(settings['insights.dashboards'] ?? {})
            .map<InsightDashboard | undefined>(dashboardKey => {
                // Select dashboard configuration from the subject settings
                const dashboardSettings = settings['insights.dashboards']?.[dashboardKey]

                if (!dashboardSettings) {
                    return undefined
                }

                // Extend settings dashboard configuration with owner info
                return {
                    owner: getDashboardOwner(subject.subject),
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
