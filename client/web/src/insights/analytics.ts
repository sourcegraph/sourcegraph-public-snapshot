import { isEqual } from 'lodash'

import { isErrorLike, ErrorLike } from '@sourcegraph/common'
import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { Settings } from '../schema/settings.schema'

export function logCodeInsightsChanges(
    oldSettingsOrError: Settings | ErrorLike,
    newSettingsOrError: Settings | ErrorLike,
    telemetryService: TelemetryService
): void {
    try {
        const oldSettings = !isErrorLike(oldSettingsOrError) && oldSettingsOrError
        const newSettings = !isErrorLike(newSettingsOrError) && newSettingsOrError

        if (oldSettings && newSettings) {
            for (const { action, insightType } of diffCodeInsightsSettings(oldSettings, newSettings)) {
                telemetryService.log(`Insight${action}`, { insightType }, { insightType })
            }
        }
    } catch {
        // noop
    }
}

interface InsightDiff {
    action: 'Addition' | 'Removal' | 'Edit'
    insightType: string
}

const BACKEND_INSIGHTS_SETTINGS_KEY = 'insights.allrepos'

export function diffCodeInsightsSettings(oldSettings: Settings, newSettings: Settings): InsightDiff[] {
    const oldInsights = new Map<string, InsightData>()
    const newInsights = new Map<string, InsightData>()

    // Top level insights (extension insights)
    for (const key of Object.keys(oldSettings)) {
        const insightMetadata = parseInsightSettingsKey(key)
        if (insightMetadata) {
            const configuration = oldSettings[key]
            oldInsights.set(insightMetadata.name, { ...insightMetadata, configuration })
        }
    }

    // Special insights BE insight store (Backend Insights)
    for (const key of Object.keys(oldSettings[BACKEND_INSIGHTS_SETTINGS_KEY] ?? {})) {
        const insightMetadata = parseInsightSettingsKey(key)
        const configuration = oldSettings[BACKEND_INSIGHTS_SETTINGS_KEY]?.[key]
        if (insightMetadata && configuration) {
            oldInsights.set(insightMetadata.name, { ...insightMetadata, configuration })
        }
    }

    // Top level insights (extension insights)
    for (const key of Object.keys(newSettings)) {
        const insightMetadata = parseInsightSettingsKey(key)
        if (insightMetadata) {
            const configuration = newSettings[key]
            newInsights.set(insightMetadata.name, { ...insightMetadata, configuration })
        }
    }

    // Special insights BE insight store (Backend Insights)
    for (const key of Object.keys(newSettings[BACKEND_INSIGHTS_SETTINGS_KEY] ?? {})) {
        const insightMetadata = parseInsightSettingsKey(key)
        const configuration = newSettings[BACKEND_INSIGHTS_SETTINGS_KEY]?.[key]
        if (insightMetadata && configuration) {
            newInsights.set(insightMetadata.name, { ...insightMetadata, configuration })
        }
    }

    const insightDiffs: InsightDiff[] = []

    for (const [name, { insightType, configuration: oldConfiguration }] of oldInsights) {
        const newInsight = newInsights.get(name)
        if (!newInsight) {
            insightDiffs.push({ insightType, action: 'Removal' })
        } else if (!isEqual(oldConfiguration, newInsight.configuration)) {
            insightDiffs.push({ insightType, action: 'Edit' })
        }
    }

    for (const [name, { insightType }] of newInsights) {
        if (!oldInsights.has(name)) {
            insightDiffs.push({ insightType, action: 'Addition' })
        }
    }

    return insightDiffs
}

interface InsightData {
    insightType: string
    name: string
    configuration: any
}

export function parseInsightSettingsKey(key: string): Omit<InsightData, 'configuration'> | undefined {
    const [insightType, maybeInsight, insightName] = key.split('.')

    if (!insightType) {
        return undefined
    }

    if (maybeInsight !== 'insight') {
        return
    }

    // Extensions that contribute insights should define settings
    // keys with the format `$EXTENSION_NAME.insight` or `$EXTENSION_NAME.insight.$INSIGHT_NAME`
    // example key: "searchInsights.insights.graphQLTypesMigration"
    return {
        insightType,
        name: insightName ?? insightType, // fallback to type
    }
}
