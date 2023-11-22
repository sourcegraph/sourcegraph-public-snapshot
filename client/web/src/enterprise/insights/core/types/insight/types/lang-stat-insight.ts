import type { BaseInsight, InsightType } from '../common'

/**
 * Extended Lang Stats Insight.
 * Some fields and settings (id, visibility) do not exist implicitly in user/org settings but
 * we have to have these to operate with insight properly.
 *
 * Note: Lang stats insight works only via Extension API and doesn't have a be version
 * (like search based insight has)
 */
export interface LangStatsInsight extends BaseInsight {
    type: InsightType.LangStats
    repository: string
    otherThreshold: number
}
