import { isEqual } from 'lodash'

import { isSettingsValid, Settings, SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'

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
    insightType: string
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

    switch (insightType) {
        // Exceptional extensions that don't adhere to the standard insight settings key format
        case 'codeStatsInsights': {
            if (key.split('.')[1] !== 'query') {
                return
            }

            return {
                insightType,
                name: insightType,
            }
        }

        // Extensions that contribute insights should define settings
        // keys with the format `$EXTENSION_NAME.insight` or `$EXTENSION_NAME.insight.$INSIGHT_NAME`
        // example key: "searchInsights.insights.graphQLTypesMigration"
        default: {
            if (maybeInsight !== 'insight') {
                return
            }

            return {
                insightType,
                name: insightName ?? insightType, // fallback to type
            }
        }
    }
}
