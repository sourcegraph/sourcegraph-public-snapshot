import { useMemo } from 'react'

import {
    ConfiguredSubjectOrError,
    SettingsCascadeOrError,
    SettingsCascadeProps,
} from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { isDefined } from '@sourcegraph/shared/src/util/types'

import { Settings } from '../../../../schema/settings.schema'
import {
    Insight,
    InsightExecutionType,
    INSIGHTS_ALL_REPOS_SETTINGS_KEY,
    InsightType,
    isInsightSettingKey,
    parseInsightTypeFromSettingId,
} from '../../core/types'
import { LangStatsInsightConfiguration } from '../../core/types/insight/lang-stat-insight'
import { SearchBasedExtensionInsightSettings } from '../../core/types/insight/search-insight'
import { useDistinctValue } from '../use-distinct-value'

export interface UseInsightProps extends SettingsCascadeProps<Settings> {
    insightId: string
}

/**
 * Returns insight from the setting cascade.
 */
export function useInsight(props: UseInsightProps): Insight | null {
    const { settingsCascade, insightId } = props

    return useMemo(() => findInsightById(settingsCascade, insightId), [settingsCascade, insightId])
}

export interface UseInsightsProps extends SettingsCascadeProps<Settings> {
    insightIds: string[]
}

export function useInsights(props: UseInsightsProps): Insight[] {
    const { settingsCascade, insightIds } = props

    const ids = useDistinctValue(insightIds)

    return useMemo(() => ids.map(id => findInsightById(settingsCascade, id)).filter(isDefined), [settingsCascade, ids])
}

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

        return {
            id: insightId,
            visibility: subject.subject.id,
            type: InsightExecutionType.Backend,
            viewType: type,
            ...insightConfiguration,
        }
    }

    return null
}
