import { InsightExecutionType, InsightType } from './common'

/**
 * Extended Lang Stats Insight.
 * Some fields and settings (id, visibility) do not exist implicitly in user/org settings but
 * we have to have these to operate with insight properly.
 *
 * Note: Lang stats insight works only via Extension API and doesn't have a be version
 * (like search based insight has)
 */
export interface LangStatsInsight {
    id: string
    executionType: InsightExecutionType.Runtime
    type: InsightType.LangStats
    title: string
    repository: string
    otherThreshold: number
    dashboardReferenceCount: number
}
