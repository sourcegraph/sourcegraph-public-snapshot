import {
    LineChartSearchInsightDataSeriesInput,
    UpdateLineChartSearchInsightInput,
    UpdatePieChartSearchInsightInput,
} from '../../../../../../../graphql-operations'
import { CaptureGroupInsight, LangStatsInsight, SearchBasedInsight } from '../../../../types'
import { isSearchBackendBasedInsight } from '../../../../types/insight/search-insight'
import { getStepInterval } from '../../utils/get-step-interval'

export function getSearchInsightUpdateInput(insight: SearchBasedInsight): UpdateLineChartSearchInsightInput {
    const repositories = !isSearchBackendBasedInsight(insight) ? insight.repositories : []
    const [unit, value] = getStepInterval(insight.step)
    const filters = isSearchBackendBasedInsight(insight)
        ? {
              includeRepoRegex: insight.filters?.includeRepoRegexp,
              excludeRepoRegex: insight.filters?.excludeRepoRegexp,
          }
        : {}

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
        viewControls: { filters },
    }
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

export function getLangStatsInsightUpdateInput(insight: LangStatsInsight): UpdatePieChartSearchInsightInput {
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
