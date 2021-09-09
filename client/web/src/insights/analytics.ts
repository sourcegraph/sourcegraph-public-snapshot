import { Duration } from 'date-fns'
import { isEqual } from 'lodash'

import {
    isSettingsValid,
    mergeSettings,
    SettingsCascade,
    SettingsCascadeOrError,
    SettingsOrgSubject,
    SettingsSiteSubject,
    SettingsSubject,
} from '@sourcegraph/shared/src/settings/settings'
import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { Settings } from '../schema/settings.schema'

const BACKEND_INSIGHTS_SETTINGS_KEY = 'insights.allrepos'

enum InsightTypePrefix {
    search = 'searchInsights.insight',
    langStats = 'codeStatsInsights.insight',
}

const isSearchBasedInsightId = (id: string): boolean => id.startsWith(InsightTypePrefix.search)
const isLangStatsdInsightId = (id: string): boolean => id.startsWith(InsightTypePrefix.langStats)

const isGlobalSubject = (subject: SettingsSubject): subject is SettingsSiteSubject => subject.__typename === 'Site'
const isOrganizationSubject = (subject: SettingsSubject): subject is SettingsOrgSubject => subject.__typename === 'Org'

export function logInsightMetrics(
    oldSettingsCascade: SettingsCascadeOrError<Settings>,
    newSettingsCascade: SettingsCascadeOrError<Settings>,
    telemetryService: TelemetryService
): void {
    logCodeInsightsCount(oldSettingsCascade, newSettingsCascade, telemetryService)
    logSearchBasedInsightStepSize(oldSettingsCascade, newSettingsCascade, telemetryService)
    logCodeInsightsChanges(oldSettingsCascade, newSettingsCascade, telemetryService)
}

export function logCodeInsightsCount(
    oldSettingsCascade: SettingsCascadeOrError<Settings>,
    newSettingsCascade: SettingsCascadeOrError<Settings>,
    telemetryService: TelemetryService
): void {
    const oldSettings = isSettingsValid(oldSettingsCascade) && oldSettingsCascade
    const newSettings = isSettingsValid(newSettingsCascade) && newSettingsCascade

    try {
        if (oldSettings && newSettings) {
            const oldGroupedInsights = getInsightsGroupedByType(oldSettings)
            const newGroupedInsights = getInsightsGroupedByType(newSettings)

            if (!isEqual(oldGroupedInsights, newGroupedInsights)) {
                telemetryService.log('InsightsGroupedCount', newGroupedInsights, newGroupedInsights)
            }
        }
    } catch {
        // noop
    }
}

export function logSearchBasedInsightStepSize(
    oldSettingsCascade: SettingsCascadeOrError<Settings>,
    newSettingsCascade: SettingsCascadeOrError<Settings>,
    telemetryService: TelemetryService
): void {
    const oldSettings = isSettingsValid(oldSettingsCascade) && oldSettingsCascade
    const newSettings = isSettingsValid(newSettingsCascade) && newSettingsCascade

    try {
        if (oldSettings && newSettings) {
            const oldGroupedStepSizes = getGroupedStepSizes(oldSettings.final)
            const newGroupedStepSizes = getGroupedStepSizes(newSettings.final)

            if (!isEqual(oldGroupedStepSizes, newGroupedStepSizes)) {
                telemetryService.log('InsightsGroupedStepSizes', newGroupedStepSizes, newGroupedStepSizes)
            }
        }
    } catch {
        // noop
    }
}

/**
 * Collect number current insights that are org-visible by type of insight.
 */
export function getGroupedStepSizes(settings: Settings): number[] {
    return Object.keys(settings)
        .filter(key => key.startsWith(InsightTypePrefix.search))
        .reduce<number[]>((stepsInDays, key) => {
            const insight = settings[key] as { step: Duration }

            return [...stepsInDays, getDaysFromInsightStep(insight.step)]
        }, [])
}

export function getDaysFromInsightStep(step: Duration): number {
    return (Object.keys(step) as (keyof Duration)[]).reduce((days, stepKey) => {
        const stepValue = step[stepKey] ?? 0

        switch (stepKey) {
            case 'years': {
                return days + stepValue * 365
            }

            case 'months': {
                return days + stepValue * 30
            }

            case 'weeks': {
                return days + stepValue * 7
            }

            case 'days': {
                return days + stepValue
            }

            default:
                return days
        }
    }, 0)
}

interface InsightGroups {
    codeStatsInsights: number
    searchBasedInsights: number
    searchBasedBackendInsights: number
    searchBasedExtensionInsights: number
}

/**
 * Collect insights count statistic from orgs settings according to
 * auth user organization list.
 */
export function getInsightsGroupedByType(settingsCascade: SettingsCascade<Settings>): InsightGroups {
    const { subjects } = settingsCascade

    const globalSubjects = subjects.filter(configuredSubject => isGlobalSubject(configuredSubject.subject))
    const orgSubjects = subjects.filter(configuredSubject => isOrganizationSubject(configuredSubject.subject))

    const finalSettingsOfAllPublicSubjects = [...globalSubjects, ...orgSubjects].reduce((finalSettings, orgSubject) => {
        const orgSettings = orgSubject.settings

        if (!orgSettings) {
            return finalSettings
        }

        const mergedSettings = mergeSettings([finalSettings, orgSettings])

        return mergedSettings ?? finalSettings
    }, {} as Settings)

    const codeStatsInsightCount = Object.keys(finalSettingsOfAllPublicSubjects).filter(isLangStatsdInsightId).length

    const searchBasedExtensionInsightCount = Object.keys(finalSettingsOfAllPublicSubjects).filter(
        isSearchBasedInsightId
    ).length

    const searchBasedBackendInsightCount = Object.keys(
        finalSettingsOfAllPublicSubjects[BACKEND_INSIGHTS_SETTINGS_KEY] ?? {}
    ).filter(isSearchBasedInsightId).length

    return {
        codeStatsInsights: codeStatsInsightCount,
        searchBasedInsights: searchBasedExtensionInsightCount + searchBasedBackendInsightCount,
        searchBasedExtensionInsights: searchBasedExtensionInsightCount,
        searchBasedBackendInsights: searchBasedBackendInsightCount,
    }
}

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
