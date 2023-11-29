import {
    type LineChartSearchInsightDataSeriesInput,
    type LineChartSearchInsightInput,
    type PieChartSearchInsightInput,
    TimeIntervalStepUnit,
} from '../../../../../../../graphql-operations'
import { InsightType } from '../../../../types'
import type {
    CreationInsightInput,
    MinimalCaptureGroupInsightData,
    MinimalComputeInsightData,
    MinimalLangStatsInsightData,
    MinimalSearchBasedInsightData,
} from '../../../code-insights-backend-types'
import { getStepInterval } from '../../utils/get-step-interval'

type CreateInsightInput = LineChartSearchInsightInput | PieChartSearchInsightInput

/**
 * Returns serialized GQL input for create insight mutation from Insight FE model.
 */
export function getInsightCreateGqlInput(
    insight: CreationInsightInput,
    dashboardId: string | null
): CreateInsightInput {
    switch (insight.type) {
        case InsightType.SearchBased: {
            return getSearchInsightCreateInput(insight, dashboardId)
        }
        case InsightType.CaptureGroup: {
            return getCaptureGroupInsightCreateInput(insight, dashboardId)
        }
        case InsightType.Compute: {
            return getComputeInsightCreateInput(insight, dashboardId)
        }
        case InsightType.LangStats: {
            return getLangStatsInsightCreateInput(insight, dashboardId)
        }
    }
}

export function getCaptureGroupInsightCreateInput(
    insight: MinimalCaptureGroupInsightData,
    dashboardId: string | null
): LineChartSearchInsightInput {
    const { step, repoQuery, filters, title } = insight
    const [unit, value] = getStepInterval(step)

    const input: LineChartSearchInsightInput = {
        repositoryScope: {
            repositories: insight.repositories,
            repositoryCriteria: repoQuery || null,
        },
        dataSeries: [
            {
                query: insight.query,
                options: {},
                timeScope: { stepInterval: { unit, value } },
                generatedFromCaptureGroups: true,
            },
        ],
        options: { title },
        viewControls: {
            seriesDisplayOptions: filters.seriesDisplayOptions,
            filters: {
                searchContexts: [filters.context],
                excludeRepoRegex: filters.excludeRepoRegexp,
                includeRepoRegex: filters.includeRepoRegexp,
            },
        },
    }

    if (dashboardId) {
        input.dashboards = [dashboardId]
    }

    return input
}

export function getSearchInsightCreateInput(
    insight: MinimalSearchBasedInsightData,
    dashboardId: string | null
): LineChartSearchInsightInput {
    const { step, repositories, repoQuery, filters, title } = insight
    const [unit, value] = getStepInterval(step)

    const input: LineChartSearchInsightInput = {
        repositoryScope: {
            repositories,
            repositoryCriteria: repoQuery || null,
        },
        dataSeries: insight.series.map<LineChartSearchInsightDataSeriesInput>(series => ({
            query: series.query,
            options: {
                label: series.name,
                lineColor: series.stroke,
            },
            timeScope: { stepInterval: { unit, value } },
        })),
        options: { title },
        viewControls: {
            seriesDisplayOptions: filters.seriesDisplayOptions,
            filters: {
                searchContexts: [filters.context],
                excludeRepoRegex: filters.excludeRepoRegexp,
                includeRepoRegex: filters.includeRepoRegexp,
            },
        },
    }

    if (dashboardId) {
        input.dashboards = [dashboardId]
    }

    return input
}

export function getLangStatsInsightCreateInput(
    insight: MinimalLangStatsInsightData,
    dashboardId: string | null
): PieChartSearchInsightInput {
    const input: PieChartSearchInsightInput = {
        // Query do not exist as setting for this type of insight, it's predefined
        // and locked on BE.
        // TODO: Remove this field as soon as BE removes this from GQL api.
        query: '',
        repositoryScope: { repositories: [insight.repository] },
        presentationOptions: {
            title: insight.title,
            otherThreshold: insight.otherThreshold,
        },
    }

    if (dashboardId) {
        input.dashboards = [dashboardId]
    }

    return input
}

export function getComputeInsightCreateInput(
    insight: MinimalComputeInsightData,
    dashboardId: string | null
): LineChartSearchInsightInput {
    const { repositories, filters, groupBy, title, series } = insight
    const input: LineChartSearchInsightInput = {
        repositoryScope: { repositories },
        dataSeries: series.map<LineChartSearchInsightDataSeriesInput>(series => ({
            query: series.query,
            options: {
                label: series.name,
                lineColor: series.stroke,
            },
            groupBy,
            timeScope: { stepInterval: { unit: TimeIntervalStepUnit.WEEK, value: 2 } },
            generatedFromCaptureGroups: true,
        })),
        options: { title },
        viewControls: {
            seriesDisplayOptions: filters.seriesDisplayOptions,
            filters: {
                searchContexts: [filters.context],
                excludeRepoRegex: filters.excludeRepoRegexp,
                includeRepoRegex: filters.includeRepoRegexp,
            },
        },
    }

    if (dashboardId) {
        input.dashboards = [dashboardId]
    }

    return input
}
