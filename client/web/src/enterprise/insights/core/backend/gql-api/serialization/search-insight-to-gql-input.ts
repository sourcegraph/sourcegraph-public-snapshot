import { UpdateLineChartSearchInsightInput } from '@sourcegraph/shared/out/src/graphql-operations'

import {
    LineChartSearchInsightDataSeriesInput,
    LineChartSearchInsightInput,
} from '../../../../../../graphql-operations'
import { InsightDashboard, isVirtualDashboard, SearchBasedInsight } from '../../../types'
import { isSearchBackendBasedInsight, SearchBasedBackendFilters } from '../../../types/insight/search-insight'
import { getStepInterval } from '../utils/insight-transformers'

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

export function getSearchInsightUpdateInput(
    insight: SearchBasedInsight & { filters?: SearchBasedBackendFilters }
): UpdateLineChartSearchInsightInput {
    const repositories = !isSearchBackendBasedInsight(insight) ? insight.repositories : []

    const [unit, value] = getStepInterval(insight.step)
    const input: UpdateLineChartSearchInsightInput = {
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
        viewControls: {
            filters: {
                includeRepoRegex: insight.filters?.includeRepoRegexp,
                excludeRepoRegex: insight.filters?.excludeRepoRegexp,
            },
        },
    }
    return input
}
