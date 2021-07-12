import { useMemo } from 'react'

import { ConfiguredSubjectOrError, SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { Settings } from '../../../schema/settings.schema'
import { InsightDashboard, InsightsDashboardType } from '../../core/types'
import { isSubjectInsightSupported } from '../../core/types/subjects'

import { getSubjectDashboards } from './utils'

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

        if (isErrorLike(settings) || !settings || !isSubjectInsightSupported(subject)) {
            return []
        }

        return getSubjectDashboards(subject, settings)
    })

    return [ALL_INSIGHTS_DASHBOARD, ...subjectDashboards]
}
