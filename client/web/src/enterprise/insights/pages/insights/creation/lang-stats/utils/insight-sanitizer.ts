import { type MinimalLangStatsInsightData, InsightType } from '../../../../../core'
import type { LangStatsCreationFormFields } from '../types'

/**
 * Converter from creation UI form values to real insight object.
 */
export const getSanitizedLangStatsInsight = (values: LangStatsCreationFormFields): MinimalLangStatsInsightData => ({
    type: InsightType.LangStats,
    title: values.title.trim(),
    repository: values.repository.trim(),
    otherThreshold: values.threshold / 100,
    dashboards: [],
})
