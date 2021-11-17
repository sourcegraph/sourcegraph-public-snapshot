import { UpdateLineChartSearchInsightInput } from '@sourcegraph/shared/src/graphql-operations'

import { LineChartSearchInsightDataSeriesInput, LineChartSearchInsightInput } from '../../../../../graphql-operations'
import { InsightDashboard, SearchBasedInsight } from '../../types'
import { isSearchBackendBasedInsight, SearchBasedBackendFilters } from '../../types/insight/search-insight'

import { getStepInterval } from './insight-transformers'

export function prepareSearchInsightCreateInput(
    insight: SearchBasedInsight,
    dashboard: InsightDashboard | null
): LineChartSearchInsightInput {
    const repositories = !isSearchBackendBasedInsight(insight) ? insight.repositories : []

    const [unit, value] = getStepInterval(insight)
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

    if (dashboard?.id) {
        input.dashboards = [dashboard.id]
    }
    return input
}

export function prepareSearchInsightUpdateInput(
    insight: SearchBasedInsight & { filters?: SearchBasedBackendFilters }
): UpdateLineChartSearchInsightInput {
    const repositories = !isSearchBackendBasedInsight(insight) ? insight.repositories : []

    const [unit, value] = getStepInterval(insight)
    const input: UpdateLineChartSearchInsightInput = {
        dataSeries: insight.series.map<LineChartSearchInsightDataSeriesInput>(series => ({
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
        viewControls: {
            filters: {
                includeRepoRegex: insight.filters?.includeRepoRegexp,
                excludeRepoRegex: insight.filters?.excludeRepoRegexp,
            },
        },
    }
    return input
}
