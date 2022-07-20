import { MinimalLangStatsInsightData, InsightExecutionType, InsightType } from '../../../../../core'
import { LangStatsCreationFormFields } from '../types'

/**
 * Converter from creation UI form values to real insight object.
 */
export const getSanitizedLangStatsInsight = (values: LangStatsCreationFormFields): MinimalLangStatsInsightData => ({
    executionType: InsightExecutionType.Runtime,
    type: InsightType.LangStats,
    title: values.title.trim(),
    repository: values.repository.trim(),
    otherThreshold: values.threshold / 100,
    seriesDisplayOptions: {},
    dashboards: [],
})
