import {
    LineChartSearchInsightDataSeriesInput,
    LineChartSearchInsightInput,
    PieChartSearchInsightInput,
} from '../../../../../../../graphql-operations'
import {
    CaptureGroupInsight,
    Insight,
    InsightDashboard,
    InsightType,
    isVirtualDashboard,
    LangStatsInsight,
    SearchBasedInsight,
} from '../../../../types'
import { isSearchBackendBasedInsight } from '../../../../types/insight/search-insight'
import { getStepInterval } from '../../utils/get-step-interval'

type CreateInsightInput = LineChartSearchInsightInput | PieChartSearchInsightInput

/**
 * Returns serialized GQL input for create insight mutation from Insight FE model.
 */
export function getInsightCreateGqlInput(insight: Insight, dashboard: InsightDashboard | null): CreateInsightInput {
    switch (insight.viewType) {
        case InsightType.SearchBased:
            return getSearchInsightCreateInput(insight, dashboard)
        case InsightType.CaptureGroup:
            return getCaptureGroupInsightCreateInput(insight, dashboard)
        case InsightType.LangStats:
            return getLangStatsInsightCreateInput(insight, dashboard)
    }
}

export function getCaptureGroupInsightCreateInput(
    insight: CaptureGroupInsight,
    dashboard: InsightDashboard | null
): LineChartSearchInsightInput {
    const [unit, value] = getStepInterval(insight.step)

    const input: LineChartSearchInsightInput = {
        dataSeries: [
            {
                query: insight.query,
                options: {},
                repositoryScope: { repositories: insight.repositories },
                timeScope: { stepInterval: { unit, value } },
                generatedFromCaptureGroups: true,
            },
        ],
        options: { title: insight.title },
    }

    if (dashboard && !isVirtualDashboard(dashboard)) {
        input.dashboards = [dashboard.id]
    }

    return input
}

export function getSearchInsightCreateInput(
    insight: SearchBasedInsight,
    dashboard: InsightDashboard | null
): LineChartSearchInsightInput {
    const repositories = !isSearchBackendBasedInsight(insight) ? insight.repositories : []

    const [unit, value] = getStepInterval(insight.step)
    const input: LineChartSearchInsightInput = {
        dataSeries: insight.series.map<LineChartSearchInsightDataSeriesInput>(series => ({
            query: series.query,
            options: {
                label: series.name,
                lineColor: series.stroke,
            },
            repositoryScope: { repositories },
            timeScope: { stepInterval: { unit, value } },
        })),
        options: { title: insight.title },
    }

    if (dashboard && !isVirtualDashboard(dashboard)) {
        input.dashboards = [dashboard.id]
    }
    return input
}

export function getLangStatsInsightCreateInput(
    insight: LangStatsInsight,
    dashboard: InsightDashboard | null
): PieChartSearchInsightInput {
    return {
        // Query do not exist as setting for this type of insight, it's predefined
        // and locked on BE.
        // TODO: Remove this field as soon as BE removes this from GQL api.
        query: '',
        repositoryScope: { repositories: [insight.repository] },
        presentationOptions: {
            title: insight.title,
            otherThreshold: insight.otherThreshold,
        },
        dashboards: [dashboard?.id ?? ''],
    }
}
