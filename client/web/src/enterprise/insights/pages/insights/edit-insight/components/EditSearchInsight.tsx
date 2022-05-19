import React, { useMemo } from 'react'

import { SubmissionErrors } from '../../../../components/form/hooks/useForm'
import {
    MinimalSearchBasedInsightData,
    InsightExecutionType,
    SearchBasedInsight,
    isSearchBackendBasedInsight,
} from '../../../../core'
import { CreateInsightFormFields, InsightStep } from '../../creation/search-insight'
import { createDefaultEditSeries } from '../../creation/search-insight/components/search-insight-creation-content/hooks/use-editable-series'
import { SearchInsightCreationContent } from '../../creation/search-insight/components/search-insight-creation-content/SearchInsightCreationContent'
import { getSanitizedSearchInsight } from '../../creation/search-insight/utils/insight-sanitizer'

interface EditSearchBasedInsightProps {
    insight: SearchBasedInsight
    onSubmit: (insight: MinimalSearchBasedInsightData) => SubmissionErrors | Promise<SubmissionErrors> | void
    onCancel: () => void
}

export const EditSearchBasedInsight: React.FunctionComponent<
    React.PropsWithChildren<EditSearchBasedInsightProps>
> = props => {
    const { insight, onSubmit, onCancel } = props

    const insightFormValues = useMemo<CreateInsightFormFields>(() => {
        if (insight.executionType === InsightExecutionType.Backend) {
            return {
                title: insight.title,
                repositories: '',
                series: insight.series.map(line => createDefaultEditSeries({ ...line, valid: true })),
                stepValue: Object.values(insight.step)[0]?.toString() ?? '3',
                step: Object.keys(insight.step)[0] as InsightStep,
                allRepos: true,
                dashboardReferenceCount: insight.dashboardReferenceCount,
            }
        }

        return {
            title: insight.title,
            repositories: insight.repositories.join(', '),
            series: insight.series.map(line => createDefaultEditSeries({ ...line, valid: true })),
            stepValue: Object.values(insight.step)[0]?.toString() ?? '3',
            step: Object.keys(insight.step)[0] as InsightStep,
            allRepos: false,
            dashboardReferenceCount: insight.dashboardReferenceCount,
        }
    }, [insight])

    const handleSubmit = (values: CreateInsightFormFields): SubmissionErrors | Promise<SubmissionErrors> | void => {
        const sanitizedInsight = getSanitizedSearchInsight(values)

        // Preserve backend insight filters since these filters aren't represented
        // in the editing form
        if (sanitizedInsight.executionType === InsightExecutionType.Backend && isSearchBackendBasedInsight(insight)) {
            return onSubmit({
                ...sanitizedInsight,
                filters: insight.filters,
            })
        }

        return onSubmit(sanitizedInsight)
    }

    return (
        <SearchInsightCreationContent
            mode="edit"
            initialValue={insightFormValues}
            dataTestId="search-insight-edit-page-content"
            className="pb-5"
            onSubmit={handleSubmit}
            onCancel={onCancel}
            insight={insight}
        />
    )
}
