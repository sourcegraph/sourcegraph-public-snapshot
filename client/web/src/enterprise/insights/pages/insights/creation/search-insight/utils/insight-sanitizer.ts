import { getSanitizedRepositories } from '../../../../../components/creation-ui-kit'
import {
    MinimalSearchBasedInsightData,
    InsightExecutionType,
    InsightType,
    SearchBasedInsightSeries,
} from '../../../../../core'
import { CreateInsightFormFields, EditableDataSeries } from '../types'

export function getSanitizedLine(line: EditableDataSeries): SearchBasedInsightSeries {
    return {
        id: line.id,
        name: line.name.trim(),
        stroke: line.stroke,
        // Query field is a reg exp field for code insight query setting
        // Native html input element adds escape symbols by itself
        // to prevent this behavior below we replace double escaping
        // with just one series of escape characters e.g. - //
        query: line.query.replace(/\\\\/g, '\\'),
    }
}

export function getSanitizedSeries(rawSeries: EditableDataSeries[]): SearchBasedInsightSeries[] {
    return rawSeries.map(getSanitizedLine)
}

/**
 * Function converter from form shape insight to insight as it is
 * presented in user/org settings.
 */
export function getSanitizedSearchInsight(rawInsight: CreateInsightFormFields): MinimalSearchBasedInsightData {
    if (rawInsight.allRepos) {
        return {
            executionType: InsightExecutionType.Backend,
            repositories: [],
            type: InsightType.SearchBased,
            title: rawInsight.title,
            series: getSanitizedSeries(rawInsight.series),
            step: { [rawInsight.step]: +rawInsight.stepValue },
            dashboards: [],
            filters: {
                excludeRepoRegexp: '',
                includeRepoRegexp: '',
                context: '',
            },
        }
    }

    return {
        executionType: InsightExecutionType.Backend,
        type: InsightType.SearchBased,
        title: rawInsight.title,
        repositories: getSanitizedRepositories(rawInsight.repositories),
        series: getSanitizedSeries(rawInsight.series),
        step: { [rawInsight.step]: +rawInsight.stepValue },
        dashboards: [],
        filters: {
            excludeRepoRegexp: '',
            includeRepoRegexp: '',
            context: '',
        },
    }
}
