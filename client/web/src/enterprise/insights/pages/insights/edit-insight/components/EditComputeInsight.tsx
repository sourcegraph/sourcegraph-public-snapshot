import { type FC, useMemo } from 'react'

import { FORM_ERROR, type SubmissionErrors } from '@sourcegraph/wildcard'

import { CodeInsightCreationMode, CodeInsightsCreationActions, createDefaultEditSeries } from '../../../../components'
import type { ComputeInsight, MinimalComputeInsightData } from '../../../../core'
import { ComputeInsightCreationContent } from '../../creation/compute/components/ComputeInsightCreationContent'
import type { CreateComputeInsightFormFields } from '../../creation/compute/types'
import { getSanitizedComputeInsight } from '../../creation/compute/utils/insight-sanitaizer'

interface EditComputeInsightProps {
    insight: ComputeInsight
    licensed: boolean
    isEditAvailable: boolean | undefined
    onSubmit: (insight: MinimalComputeInsightData) => SubmissionErrors | Promise<SubmissionErrors> | void
    onCancel: () => void
}

export const EditComputeInsight: FC<EditComputeInsightProps> = props => {
    const { insight, licensed, isEditAvailable, onCancel, onSubmit } = props

    const insightFormValues = useMemo<CreateComputeInsightFormFields>(
        () => ({
            title: insight.title,
            repositories: insight.repositories,
            series: insight.series.map(line => createDefaultEditSeries({ ...line, valid: true })),
            dashboardReferenceCount: insight.dashboardReferenceCount,
            groupBy: insight.groupBy,
        }),
        [insight]
    )

    const handleSubmit = (
        values: CreateComputeInsightFormFields
    ): SubmissionErrors | Promise<SubmissionErrors> | void => {
        const sanitizedInsight = getSanitizedComputeInsight(values)
        return onSubmit({
            ...sanitizedInsight,
            filters: insight.filters,
        })
    }

    return (
        <ComputeInsightCreationContent
            touched={true}
            initialValue={insightFormValues}
            className="pb-5"
            onSubmit={handleSubmit}
        >
            {form => (
                <CodeInsightsCreationActions
                    mode={CodeInsightCreationMode.Edit}
                    licensed={licensed}
                    available={isEditAvailable}
                    submitting={form.submitting}
                    errors={form.submitErrors?.[FORM_ERROR]}
                    clear={form.isFormClearActive}
                    onCancel={onCancel}
                />
            )}
        </ComputeInsightCreationContent>
    )
}
