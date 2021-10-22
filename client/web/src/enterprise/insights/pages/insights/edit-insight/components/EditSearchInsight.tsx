import React, { useMemo } from 'react'

import { SubmissionErrors } from '../../../../components/form/hooks/useForm'
import { InsightType, SearchBasedInsight } from '../../../../core/types'
import { isSearchBackendBasedInsight } from '../../../../core/types/insight/search-insight'
import { SupportedInsightSubject } from '../../../../core/types/subjects'
import { createDefaultEditSeries } from '../../creation/search-insight/components/search-insight-creation-content/hooks/use-editable-series'
import { SearchInsightCreationContent } from '../../creation/search-insight/components/search-insight-creation-content/SearchInsightCreationContent'
import { CreateInsightFormFields, InsightStep } from '../../creation/search-insight/types'
import { getSanitizedSearchInsight } from '../../creation/search-insight/utils/insight-sanitizer'

interface EditSearchBasedInsightProps {
    insight: SearchBasedInsight
    subjects: SupportedInsightSubject[]
    onSubmit: (insight: SearchBasedInsight) => SubmissionErrors | Promise<SubmissionErrors> | void
    onCancel: () => void
}

export const EditSearchBasedInsight: React.FunctionComponent<EditSearchBasedInsightProps> = props => {
    const { insight, subjects, onSubmit, onCancel } = props

    const insightFormValues = useMemo<CreateInsightFormFields>(() => {
        if (insight.type === InsightType.Backend) {
            return {
                title: insight.title,
                visibility: insight.visibility,
                repositories: '',
                series: insight.series.map(line => createDefaultEditSeries({ ...line, valid: true })),
                stepValue: '2',
                step: 'weeks',
                allRepos: true,
            }
        }

        return {
            title: insight.title,
            visibility: insight.visibility,
            repositories: insight.repositories.join(', '),
            series: insight.series.map(line => createDefaultEditSeries({ ...line, valid: true })),
            stepValue: Object.values(insight.step)[0]?.toString() ?? '3',
            step: Object.keys(insight.step)[0] as InsightStep,
            allRepos: false,
        }
    }, [insight])

    // Handlers
    const handleSubmit = (values: CreateInsightFormFields): SubmissionErrors | Promise<SubmissionErrors> | void => {
        const sanitizedInsight = getSanitizedSearchInsight(values)

        // Preserve backend insight filters since these filters don't represent in form fields
        // in case if editing hasn't change type of search insight.
        if (isSearchBackendBasedInsight(sanitizedInsight) && isSearchBackendBasedInsight(insight)) {
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
            className="pb-5"
            initialValue={insightFormValues}
            subjects={subjects}
            dataTestId="search-insight-edit-page-content"
            onSubmit={handleSubmit}
            onCancel={onCancel}
        />
    )
}
