import { InsightType, InsightTypePrefix, SyntheticInsightFields } from './common'

/**
 * Extended Lang Stats Insight.
 * Some fields and settings (id, visibility) do not exist implicitly in user/org settings but
 * we have to have these to operate with insight properly.
 *
 * Note: Lang stats insight works only via Extension API and doesn't have a be version
 * (like search based insight has)
 */
export interface LangStatsInsight extends LangStatsInsightConfiguration, SyntheticInsightFields {
    type: InsightType.Extension
}

/**
 * Lang stats insight as it is presented in user/org settings cascade.
 * See public API of code stats extension
 * https://github.com/sourcegraph/sourcegraph-code-stats-insights/blob/master/package.json#L27
 */
export interface LangStatsInsightConfiguration {
    title: string
    repository: string
    otherThreshold: number
}

/**
 * Since we use insight name conventions to distinguish between insight types.
 * Example id for the lang stats insight: "codeStatsInsights.insight.myFirstLangStatInsight"
 */
export const isLangStatsdInsightId = (id: string): boolean => id.startsWith(InsightTypePrefix.langStats)
