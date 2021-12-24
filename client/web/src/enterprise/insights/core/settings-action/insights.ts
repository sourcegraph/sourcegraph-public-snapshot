import { get } from 'lodash'

import { isErrorLike } from '@sourcegraph/common'
import { modify, parseJSONCOrError } from '@sourcegraph/shared/src/util/jsonc'

import {
    Insight,
    InsightDashboard,
    INSIGHTS_ALL_REPOS_SETTINGS_KEY,
    InsightExecutionType,
    InsightTypePrefix,
    isLangStatsInsight,
    isVirtualDashboard,
    InsightConfiguration,
    isSearchBasedInsight,
} from '../types'
import { isCustomInsightDashboard } from '../types/dashboard/real-dashboard'

import { addInsightToDashboard } from './dashboards'

/**
 * Returns insight settings key. Since different types of insight live in different
 * places in the settings we have to derive this place (settings key) by insight types
 */
const getInsightSettingKey = (insight: Insight): string[] => {
    // Lang stats insight always lives on top level by its unique id
    if (isLangStatsInsight(insight)) {
        return [insight.id]
    }

    // Search based insight may live in two main places
    switch (insight.type) {
        // Extension based lives on top level of settings file by its id
        case InsightExecutionType.Runtime: {
            return [insight.id]
        }

        // Backend based insight lives in insights.allrepos map
        case InsightExecutionType.Backend: {
            return [INSIGHTS_ALL_REPOS_SETTINGS_KEY, insight.id]
        }
    }
}

export const addInsight = (settings: string, insight: Insight, dashboard: InsightDashboard | null): string => {
    const dashboardSettingKey =
        dashboard && !isVirtualDashboard(dashboard) && isCustomInsightDashboard(dashboard)
            ? dashboard.settingsKey
            : undefined

    const transforms = [
        (settings: string) => addInsightToSettings(settings, insight),
        (settings: string) =>
            dashboardSettingKey ? addInsightToDashboard(settings, dashboardSettingKey, insight.id) : settings,
    ]

    return transforms.reduce((settings, transformer) => transformer(settings), settings)
}

/**
 * Serializes and adds insight configurations to the settings content string (jsonc).
 * Returns settings content string with serialized insight inside.
 *
 * @param settings - original settings content string
 * @param insight - insight configuration to add in settings file
 */
export const addInsightToSettings = (settings: string, insight: Insight): string => {
    const insightSettingsKey = getInsightSettingKey(insight)

    // Add insight to the user settings
    return modify(settings, insightSettingsKey, getSanitizedInsight(insight))
}

/**
 * Returns insight configuration, removes all synthetic properties from the insight object
 */
const getSanitizedInsight = (insight: Insight): InsightConfiguration | undefined => {
    if (isLangStatsInsight(insight)) {
        const { id, visibility, type, viewType, ...originalInsight } = insight

        return originalInsight
    }

    if (isSearchBasedInsight(insight)) {
        const { id, visibility, type, viewType, ...originalInsight } = insight
        const sanitizedSeries = originalInsight.series.map(line => ({
            name: line.name,
            query: line.query,
            stroke: line.stroke,
        }))

        return { ...originalInsight, series: sanitizedSeries }
    }

    return
}

interface RemoveInsightFromSettingsInputs {
    originalSettings: string
    insightID: string
    isOldCodeStatsInsight?: boolean
}

/**
 * Return edited settings without deleted insight.
 */
export const removeInsightFromSettings = (props: RemoveInsightFromSettingsInputs): string => {
    const {
        originalSettings,
        insightID,
        // For backward compatibility with old code stats insight api we have to delete
        // this insight in a special way. See link below for more information.
        // https://github.com/sourcegraph/sourcegraph-code-stats-insights/blob/master/src/code-stats-insights.ts#L33
        isOldCodeStatsInsight = insightID === `${InsightTypePrefix.langStats}.language`,
    } = props

    if (isOldCodeStatsInsight) {
        const editedSettings = modify(
            originalSettings,
            // According to our naming convention <type>.insight.<name>
            ['codeStatsInsights.query'],
            undefined
        )

        return modify(
            editedSettings,
            // According to our naming convention <type>.insight.<name>
            ['codeStatsInsights.otherThreshold'],
            undefined
        )
    }

    // Just to be sure that we removed this insight whatever this insight is (backend or extension based)
    // Remove this insight from top level of settings file and from insights.allrepos
    const allPossibleInsightSettingsKeys = [[insightID], [INSIGHTS_ALL_REPOS_SETTINGS_KEY, insightID]]

    let editedSettings = originalSettings
    const parsedSettings = parseJSONCOrError<object>(originalSettings)

    if (isErrorLike(parsedSettings)) {
        return originalSettings
    }

    for (const settingsKey of allPossibleInsightSettingsKeys) {
        // If settings content jsonc doesn't have a value under the settingsKey
        // it fails with parsing error. We should check existence of the property that
        // we're about to remove
        if (get(parsedSettings, settingsKey)) {
            editedSettings = modify(
                originalSettings,
                // According to our naming convention <type>.insight.<name>
                settingsKey,
                undefined
            )
        }
    }

    return editedSettings
}
