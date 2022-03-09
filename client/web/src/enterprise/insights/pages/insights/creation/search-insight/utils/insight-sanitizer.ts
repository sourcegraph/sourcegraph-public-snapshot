import { camelCase } from 'lodash'

import { getSanitizedRepositories } from '../../../../../components/creation-ui-kit'
import {
    InsightExecutionType,
    InsightType,
    InsightTypePrefix,
    SearchBasedInsight,
    SearchBasedInsightSeries,
} from '../../../../../core/types'
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
export function getSanitizedSearchInsight(rawInsight: CreateInsightFormFields): SearchBasedInsight {
    // Backend type of insight.
    if (rawInsight.allRepos) {
        return {
            id: `${InsightTypePrefix.search}.${camelCase(rawInsight.title)}`,
            type: InsightExecutionType.Backend,
            viewType: InsightType.SearchBased,
            title: rawInsight.title,
            series: getSanitizedSeries(rawInsight.series),
            visibility: '',
            step: { [rawInsight.step]: +rawInsight.stepValue },
            dashboardReferenceCount: rawInsight.dashboardReferenceCount,
        }
    }

    return {
        // ID generated according to our naming insight convention
        // <Type of insight>.insight.<name of insight>
        id: `${InsightTypePrefix.search}.${camelCase(rawInsight.title)}`,
        type: InsightExecutionType.Runtime,
        viewType: InsightType.SearchBased,
        title: rawInsight.title,
        repositories: getSanitizedRepositories(rawInsight.repositories),
        series: getSanitizedSeries(rawInsight.series),
        step: { [rawInsight.step]: +rawInsight.stepValue },
        dashboardReferenceCount: rawInsight.dashboardReferenceCount,
    }
}
