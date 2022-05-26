import {
    InsightViewFiltersInput,
    LineChartSearchInsightDataSeriesInput,
    SeriesDisplayOptionsInput,
    UpdateLineChartSearchInsightInput,
    UpdatePieChartSearchInsightInput,
} from '../../../../../../../graphql-operations'
import { parseSeriesDisplayOptions } from '../../../../../components/insights-view-grid/components/backend-insight/components/drill-down-filters-panel/drill-down-filters/utils'
import {
    MinimalCaptureGroupInsightData,
    MinimalLangStatsInsightData,
    MinimalSearchBasedInsightData,
} from '../../../code-insights-backend-types'
import { getStepInterval } from '../../utils/get-step-interval'

export function getSearchInsightUpdateInput(insight: MinimalSearchBasedInsightData): UpdateLineChartSearchInsightInput {
    const repositories = insight.repositories
    const [unit, value] = getStepInterval(insight.step)
    const filters: InsightViewFiltersInput = {
        includeRepoRegex: insight.filters.includeRepoRegexp,
        excludeRepoRegex: insight.filters.excludeRepoRegexp,
        searchContexts: insight.filters.context ? [insight.filters.context] : [],
    }

    // TODO: update when sorting all insights are supported
    const seriesDisplayOptions: SeriesDisplayOptionsInput = {}

    return {
        dataSeries: insight.series.map<LineChartSearchInsightDataSeriesInput>(series => ({
            seriesId: series.id,
            query: series.query,
            options: {
                label: series.name,
                lineColor: series.stroke,
            },
            repositoryScope: { repositories },
            timeScope: { stepInterval: { unit, value } },
        })),
        presentationOptions: {
            title: insight.title,
        },
        viewControls: { filters, seriesDisplayOptions },
    }
}

export function getCaptureGroupInsightUpdateInput(
    insight: MinimalCaptureGroupInsightData
): UpdateLineChartSearchInsightInput {
    const { step, filters, query, title, repositories, seriesDisplayOptions } = insight
    const [unit, value] = getStepInterval(step)

    return {
        dataSeries: [
            {
                query,
                options: {},
                repositoryScope: { repositories },
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
                searchContexts: insight.filters.context ? [filters.context] : [],
            },
            seriesDisplayOptions: parseSeriesDisplayOptions(seriesDisplayOptions),
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
