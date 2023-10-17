import {
    type LineChartSearchInsightDataSeriesInput,
    TimeIntervalStepUnit,
    type UpdateLineChartSearchInsightInput,
    type UpdatePieChartSearchInsightInput,
} from '../../../../../../../graphql-operations'
import type {
    MinimalCaptureGroupInsightData,
    MinimalComputeInsightData,
    MinimalLangStatsInsightData,
    MinimalSearchBasedInsightData,
} from '../../../code-insights-backend-types'
import { getStepInterval } from '../../utils/get-step-interval'

export function getSearchInsightUpdateInput(insight: MinimalSearchBasedInsightData): UpdateLineChartSearchInsightInput {
    const { title, repositories, repoQuery, series, filters, step } = insight

    const [unit, value] = getStepInterval(step)

    return {
        repositoryScope: {
            repositories,
            repositoryCriteria: repoQuery || null,
        },
        dataSeries: series.map<LineChartSearchInsightDataSeriesInput>(series => ({
            seriesId: series.id,
            query: series.query,
            options: {
                label: series.name,
                lineColor: series.stroke,
            },
            timeScope: { stepInterval: { unit, value } },
        })),
        presentationOptions: { title },
        viewControls: {
            filters: {
                includeRepoRegex: filters.includeRepoRegexp,
                excludeRepoRegex: filters.excludeRepoRegexp,
                searchContexts: filters.context ? [filters.context] : [],
            },
            seriesDisplayOptions: filters.seriesDisplayOptions,
        },
    }
}

export function getCaptureGroupInsightUpdateInput(
    insight: MinimalCaptureGroupInsightData
): UpdateLineChartSearchInsightInput {
    const { step, filters, query, title, repositories, repoQuery } = insight
    const [unit, value] = getStepInterval(step)

    return {
        repositoryScope: { repositories, repositoryCriteria: repoQuery || null },
        dataSeries: [
            {
                query,
                options: {},
                timeScope: { stepInterval: { unit, value } },
                generatedFromCaptureGroups: true,
            },
        ],
        presentationOptions: {
            title,
        },
        viewControls: {
            filters: {
                includeRepoRegex: filters.includeRepoRegexp,
                excludeRepoRegex: filters.excludeRepoRegexp,
                searchContexts: filters.context ? [filters.context] : [],
            },
            seriesDisplayOptions: filters.seriesDisplayOptions,
        },
    }
}

export function getComputeInsightUpdateInput(insight: MinimalComputeInsightData): UpdateLineChartSearchInsightInput {
    const { repositories, filters, groupBy } = insight

    return {
        dataSeries: insight.series.map<LineChartSearchInsightDataSeriesInput>(series => ({
            seriesId: series.id,
            query: series.query,
            options: {
                label: series.name,
                lineColor: series.stroke,
            },
            groupBy,
            repositoryScope: { repositories },
            // TODO: Remove this when BE supports seperate mutation for compute-powered insight
            timeScope: { stepInterval: { unit: TimeIntervalStepUnit.WEEK, value: 2 } },
            generatedFromCaptureGroups: true,
        })),
        presentationOptions: {
            title: insight.title,
        },
        // TODO: update when sorting all insights are supported
        viewControls: {
            filters: {
                includeRepoRegex: filters.includeRepoRegexp,
                excludeRepoRegex: filters.excludeRepoRegexp,
                searchContexts: filters.context ? [filters.context] : [],
            },
            seriesDisplayOptions: {},
        },
    }
}

export function getLangStatsInsightUpdateInput(insight: MinimalLangStatsInsightData): UpdatePieChartSearchInsightInput {
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
    }
}
