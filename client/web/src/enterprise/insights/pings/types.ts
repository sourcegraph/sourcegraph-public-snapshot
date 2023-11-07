import { InsightType } from '../core'

export enum CodeInsightTrackType {
    SearchBasedInsight = 'SearchBased',
    LangStatsInsight = 'LangStats',
    CaptureGroupInsight = 'CaptureGroup',
    ComputeInsight = 'ComputeInsight',
    InProductLandingPageInsight = 'InProductLandingPageInsight',
    CloudLandingPageInsight = 'CloudLandingPageInsight',
}

export const getTrackingTypeByInsightType = (insightType: InsightType): CodeInsightTrackType => {
    switch (insightType) {
        case InsightType.CaptureGroup: {
            return CodeInsightTrackType.CaptureGroupInsight
        }
        case InsightType.SearchBased: {
            return CodeInsightTrackType.SearchBasedInsight
        }
        case InsightType.LangStats: {
            return CodeInsightTrackType.LangStatsInsight
        }
        case InsightType.Compute: {
            return CodeInsightTrackType.ComputeInsight
        }
    }
}
