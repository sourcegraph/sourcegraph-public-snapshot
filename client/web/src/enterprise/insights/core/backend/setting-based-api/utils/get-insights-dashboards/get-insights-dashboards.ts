import { ErrorLike, isErrorLike } from '@sourcegraph/common'
import { ConfiguredSubjectOrError } from '@sourcegraph/shared/src/settings/settings'

import { Settings } from '../../../../../../../schema/settings.schema'
import { InsightDashboard } from '../../../../types'
import { ALL_INSIGHTS_DASHBOARD } from '../../../../types/dashboard/virtual-dashboard'
import { isSubjectInsightSupported } from '../../../../types/subjects'

import { getInsightIdsFromSettings, getSubjectDashboards } from './utils'

/**
 * Returns all valid and reachable for a user insight-dashboards.
 */
export function getInsightsDashboards(
    subjects: ConfiguredSubjectOrError<Settings>[] | null,
    final: Settings | ErrorLike | null
): InsightDashboard[] {
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

    const normalizedFinalSettings = !final || isErrorLike(final) ? {} : final

    return [
        {
            ...ALL_INSIGHTS_DASHBOARD,
            // Get all reachable insight ids from the final (merged settings)
            insightIds: getInsightIdsFromSettings(normalizedFinalSettings),
        },
        ...subjectDashboards,
    ]
}
