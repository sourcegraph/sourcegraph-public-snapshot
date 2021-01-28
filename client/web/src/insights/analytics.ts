import { isEqual } from 'lodash'
import { isSettingsValid, Settings, SettingsCascadeOrError } from '../../../shared/src/settings/settings'
import { TelemetryService } from '../../../shared/src/telemetry/telemetryService'

const insightTypes = ['searchInsights', 'codeStatsInsights'] as const
type InsightType = typeof insightTypes[number]

export function logCodeInsightsChanges(
    oldSettingsCascade: SettingsCascadeOrError<Settings>,
    newSettingsCascade: SettingsCascadeOrError<Settings>,
    telemetryService: TelemetryService
): void {
    try {
        const oldSettings = isSettingsValid(oldSettingsCascade) && oldSettingsCascade.final
        const newSettings = isSettingsValid(newSettingsCascade) && newSettingsCascade.final

        if (oldSettings && newSettings) {
            for (const { action, insightType } of diffCodeInsightsSettings(oldSettings, newSettings)) {
                telemetryService.log(`Insight${action}`, { insightType })
            }
        }
    } catch {
        // noop
    }
}

interface InsightDiff {
    action: 'Addition' | 'Removal' | 'Edit'
    insightType: InsightType
}

export function diffCodeInsightsSettings(oldSettings: Settings, newSettings: Settings): InsightDiff[] {
    const oldInsights = new Map<string, InsightData>()
    const newInsights = new Map<string, InsightData>()

    for (const key of Object.keys(oldSettings)) {
        const insightMetadata = parseInsightSettingsKey(key)
        if (insightMetadata) {
            const configuration = oldSettings[key]
            oldInsights.set(insightMetadata.name, { ...insightMetadata, configuration })
        }
    }

    for (const key of Object.keys(newSettings)) {
        const insightMetadata = parseInsightSettingsKey(key)
        if (insightMetadata) {
            const configuration = newSettings[key]
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

/**
 * Get the type of insight from a settings key. For example:
 * "searchInsights.searchInsights.insight.graphQLTypesMigration.insightsPage"
 * to "searchInsights"
 */
export function getInsightTypeFromSettingsKey(key: string): InsightType | undefined {
    return insightTypes.find(type => type === key.split('.')[0])
}

interface InsightData {
    insightType: InsightType
    name: string
    configuration: any
}

function parseInsightSettingsKey(key: string): Omit<InsightData, 'configuration'> | undefined {
    const insightType = getInsightTypeFromSettingsKey(key)

    if (!insightType) {
        return undefined
    }

    switch (insightType) {
        case 'searchInsights': {
            // example key: "searchInsights.insights.graphQLTypesMigration"
            const parts = key.split('.')
            if (parts[1] !== 'insight') {
                return
            }

            return {
                insightType,
                name: parts[2] ?? insightType, // fallback to type
            }
        }

        case 'codeStatsInsights':
            if (key.split('.')[1] !== 'query') {
                return
            }

            return {
                insightType,
                name: insightType,
            }
    }
}
