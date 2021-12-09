import { LineChartSearchInsightInput, UpdateLineChartSearchInsightInput } from '../../../../../../graphql-operations'
import { CaptureGroupInsight, InsightDashboard, isVirtualDashboard } from '../../../types'
import { getStepInterval } from '../utils/insight-transformers'

/**
 * Returns gql variable input for creation capture group insight through gql api.
 */
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

export function getCaptureGroupInsightUpdateInput(insight: CaptureGroupInsight): UpdateLineChartSearchInsightInput {
    const [unit, value] = getStepInterval(insight.step)

    return {
        dataSeries: [
            {
                query: insight.query,
                options: {},
                repositoryScope: { repositories: insight.repositories },
                timeScope: { stepInterval: { unit, value } },
                generatedFromCaptureGroups: true,
            },
        ],
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
}
