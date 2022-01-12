import { isErrorLike } from '@sourcegraph/common'
import { ConfiguredSubjectOrError, SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'

import { Settings } from '../../../../../../schema/settings.schema'
import {
    Insight,
    InsightExecutionType,
    INSIGHTS_ALL_REPOS_SETTINGS_KEY,
    InsightType,
    isInsightSettingKey,
    parseInsightTypeFromSettingId,
} from '../../../types'
import { LangStatsInsightConfiguration } from '../../../types/insight/lang-stat-insight'
import { SearchBasedExtensionInsightSettings } from '../../../types/insight/search-insight'

export function findInsightById(settingsCascade: SettingsCascadeOrError<Settings>, insightId: string): Insight | null {
    const subjects = settingsCascade.subjects

    const subject = subjects?.find(
        ({ settings }) =>
            settings &&
            !isErrorLike(settings) &&
            (settings[insightId] ||
                // Also check insights all repos map as a second place of insights
                (settings[INSIGHTS_ALL_REPOS_SETTINGS_KEY] as Record<string, Insight>)?.[insightId])
    )

    return subject ? parseInsightFromSubject(insightId, subject) : null
}

export function parseInsightFromSubject(
    insightId: string,
    subject: ConfiguredSubjectOrError<Settings>
): Insight | null {
    if (!isInsightSettingKey(insightId) || !subject?.settings || isErrorLike(subject.settings)) {
        return null
    }

    const type = parseInsightTypeFromSettingId(insightId)

    // Early return in case if we don't support this type of insight
    if (type === null) {
        return null
    }

    // Top level match means we are dealing with extension based insights
    if (subject.settings[insightId]) {
        if (type === InsightType.LangStats) {
            const insightConfiguration = subject.settings[insightId] as LangStatsInsightConfiguration

            return {
                id: insightId,
                visibility: subject.subject.id,
                type: InsightExecutionType.Runtime,
                viewType: type,
                ...insightConfiguration,
            }
        }

        if (type === InsightType.SearchBased) {
            const insightConfiguration = subject.settings[insightId] as SearchBasedExtensionInsightSettings

            return {
                id: insightId,
                visibility: subject.subject.id,
                type: InsightExecutionType.Runtime,
                viewType: type,
                ...insightConfiguration,
            }
        }
    }

    const allReposInsights = subject.settings[INSIGHTS_ALL_REPOS_SETTINGS_KEY] ?? {}

    // Match in all repos object means that we are dealing with backend search based insight.
    // At the moment we support only search based insight in setting BE insight map
    if (allReposInsights[insightId] && type === InsightType.SearchBased) {
        const insightConfiguration = allReposInsights[insightId]
        const filters = insightConfiguration.filters ?? { includeRepoRegexp: '', excludeRepoRegexp: '' }

        return {
            id: insightId,
            visibility: subject.subject.id,
            type: InsightExecutionType.Backend,
            step: { months: 1 },
            viewType: type,
            ...insightConfiguration,
            series: insightConfiguration.series?.map((line, index) => ({
                id: `${line.name}-${index}`,
                ...line,
            })),
            filters,
        }
    }

    return null
}
